using System;
using System.IO;
using System.Net;
using System.Text;
using System.Text.Json;
using System.Threading;
using System.Threading.Tasks;
using System.Threading.Channels;
using ModelContextProtocol.Protocol;

namespace controller_mcp
{
    public class HttpListenerSseServerTransport : ITransport
    {
        private readonly HttpListener _listener;
        private readonly Channel<JsonRpcMessage> _messageChannel;
        private HttpListenerContext _sseContext;
        private CancellationTokenSource _cts;
        private CancellationTokenSource _activeClientCts;

        public HttpListenerSseServerTransport(int port)
        {
            _listener = new HttpListener();
            _listener.Prefixes.Add($"http://localhost:{port}/mcp/");
            _messageChannel = Channel.CreateUnbounded<JsonRpcMessage>();
            SessionId = Guid.NewGuid().ToString();
        }

        public string SessionId { get; }

        public ChannelReader<JsonRpcMessage> MessageReader => _messageChannel.Reader;

        public Action<string> OnLog { get; set; }

        public async Task StartAsync(CancellationToken cancellationToken)
        {
            _cts = CancellationTokenSource.CreateLinkedTokenSource(cancellationToken);
            _listener.Start();
            OnLog?.Invoke($"Listening on {string.Join(", ", _listener.Prefixes)}");

            _ = Task.Run(() => AcceptConnectionsAsync(_cts.Token), _cts.Token);
        }

        private async Task AcceptConnectionsAsync(CancellationToken ct)
        {
            while (!ct.IsCancellationRequested)
            {
                try
                {
                    var context = await _listener.GetContextAsync();
                    _ = Task.Run(() => HandleRequestAsync(context, ct), ct);
                }
                catch (HttpListenerException) when (ct.IsCancellationRequested)
                {
                    // Ignore expected exception on shutdown
                }
                catch (ObjectDisposedException)
                {
                    // Ignore expected exception on shutdown when the listener is disposed
                }
                catch (Exception ex)
                {
                    if (!ct.IsCancellationRequested)
                    {
                        OnLog?.Invoke($"Error accepting connection: {ex.Message}");
                    }
                }
            }
        }

        private async Task HandleRequestAsync(HttpListenerContext context, CancellationToken ct)
        {
            var req = context.Request;
            var res = context.Response;

            OnLog?.Invoke($"[HTTP IN] {req.HttpMethod} {req.Url.AbsolutePath}");
            if (req.Headers.Count > 0)
            {
                var sb = new StringBuilder();
                foreach (string key in req.Headers.AllKeys)
                    sb.Append($"\n  {key}: {req.Headers[key]}");
                OnLog?.Invoke($"[HTTP IN HEADERS]{sb.ToString()}");
            }

            if (req.Url.AbsolutePath.EndsWith("/mcp/sse", StringComparison.OrdinalIgnoreCase) && req.HttpMethod == "GET")
            {
                res.ContentType = "text/event-stream";
                res.Headers.Add("Cache-Control", "no-cache");
                res.Headers.Add("Connection", "keep-alive");
                res.Headers.Add("Access-Control-Allow-Origin", "*");
                res.SendChunked = true;

                if (res.Headers.Count > 0)
                {
                    var sb = new StringBuilder();
                    foreach (string key in res.Headers.AllKeys)
                        sb.Append($"\n  {key}: {res.Headers[key]}");
                    OnLog?.Invoke($"[HTTP OUT HEADERS]{sb.ToString()}");
                }

                _activeClientCts?.Cancel();
                _activeClientCts?.Dispose();
                _activeClientCts = CancellationTokenSource.CreateLinkedTokenSource(ct);
                var clientCt = _activeClientCts.Token;

                _sseContext = context;
                OnLog?.Invoke("Client connected to SSE stream via GET.");

                try
                {
                    string absolutePostUrl = $"http://{req.UserHostName}/mcp/messages";
                    string initStr = $"event: endpoint\ndata: {absolutePostUrl}\n\n";
                    var initMessage = Encoding.UTF8.GetBytes(initStr);
                    await res.OutputStream.WriteAsync(initMessage, 0, initMessage.Length, clientCt);
                    await res.OutputStream.FlushAsync();
                    
                    OnLog?.Invoke($"[SSE OUT] {initStr.TrimEnd()}");

                    // Hold the connection open using the client specific token
                    while (!clientCt.IsCancellationRequested)
                    {
                        await Task.Delay(15000, clientCt); // 15s heartbeat
                        var pingMsg = Encoding.UTF8.GetBytes(": ping\n\n");
                        await res.OutputStream.WriteAsync(pingMsg, 0, pingMsg.Length, clientCt);
                        await res.OutputStream.FlushAsync();
                    }
                }
                catch (Exception ex)
                {
                    OnLog?.Invoke($"SSE connection error: {ex.Message}");
                }
            }
            else if (req.HttpMethod == "POST")
            {
                string body = null;
                try
                {
                    using var reader = new StreamReader(req.InputStream, req.ContentEncoding ?? Encoding.UTF8);
                    body = await reader.ReadToEndAsync();
                }
                catch (Exception ex)
                {
                    OnLog?.Invoke($"[HTTP IN PAYLOAD ERROR] Client disconnected or stream failed: {ex.Message}");
                    res.StatusCode = 400; // Bad Request
                    res.Close();
                    return;
                }
                
                OnLog?.Invoke($"[HTTP IN PAYLOAD] {body}");
                
                JsonRpcMessage msg = null;
                try
                {
                    var opts = new JsonSerializerOptions { PropertyNameCaseInsensitive = true };
                    msg = JsonSerializer.Deserialize<JsonRpcMessage>(body, opts);
                    controller_mcp.Features.Tools.AuditLogger.LogJsonRpc("IN", body);
                    controller_mcp.Features.Tools.AnalyticsManager.TrackRequest();
                    controller_mcp.Features.Tools.AnalyticsManager.TrackBytesReceived(req.ContentLength64 > 0 ? req.ContentLength64 : body.Length);
                    
                    try 
                    {
                        using (var doc = System.Text.Json.JsonDocument.Parse(body))
                        {
                            if (doc.RootElement.TryGetProperty("method", out var methodProp) && methodProp.GetString() == "tools/call")
                            {
                                if (doc.RootElement.TryGetProperty("params", out var paramsProp) && paramsProp.TryGetProperty("name", out var nameProp))
                                {
                                    controller_mcp.Features.Tools.AnalyticsManager.TrackToolInvocation(nameProp.GetString());
                                }
                            }
                        }
                    } catch { }
                }
                catch (Exception ex)
                {
                    OnLog?.Invoke($"Error parsing message: {ex.Message}");
                    controller_mcp.Features.Tools.AnalyticsManager.TrackError();
                    res.StatusCode = 400; // Bad Request
                    res.Close();
                    return;
                }

                // Streamable HTTP: If a POST comes in to /mcp/sse and we have no SSE connection, upgrade this POST to the SSE stream!
                if (_sseContext == null && msg != null && req.Url.AbsolutePath.EndsWith("/mcp/sse", StringComparison.OrdinalIgnoreCase))
                {
                    OnLog?.Invoke("Upgrading initial POST request to SSE stream (Cursor Streamable HTTP).");
                    res.ContentType = "text/event-stream";
                    res.Headers.Add("Cache-Control", "no-cache");
                    res.Headers.Add("Connection", "keep-alive");
                    res.Headers.Add("Access-Control-Allow-Origin", "*");
                    res.SendChunked = true;

                    _activeClientCts?.Cancel();
                    _activeClientCts?.Dispose();
                    _activeClientCts = CancellationTokenSource.CreateLinkedTokenSource(ct);
                    var clientCt = _activeClientCts.Token;

                    _sseContext = context;

                    try
                    {
                        string absolutePostUrl = $"http://{req.UserHostName}/mcp/messages";
                        string initStr = $"event: endpoint\ndata: {absolutePostUrl}\n\n";
                        var initMessage = Encoding.UTF8.GetBytes(initStr);
                        await res.OutputStream.WriteAsync(initMessage, 0, initMessage.Length, clientCt);
                        await res.OutputStream.FlushAsync();
                        
                        OnLog?.Invoke($"[SSE OUT] {initStr.TrimEnd()}");

                        // Forward the message to the server
                        await _messageChannel.Writer.WriteAsync(msg, clientCt);

                        // Keep the POST connection alive forever as the SSE stream
                        await Task.Delay(-1, clientCt);
                    }
                    catch (Exception ex)
                    {
                        OnLog?.Invoke($"Streamable HTTP connection error: {ex.Message}");
                    }
                }
                else
                {
                    // Standard POST message routing
                    if (msg != null)
                    {
                        await _messageChannel.Writer.WriteAsync(msg, ct);
                    }
                    
                    res.Headers.Add("Access-Control-Allow-Origin", "*");
                    res.StatusCode = 202; // Accepted
                    res.StatusDescription = "Accepted";
                    res.Close();
                }
            }
            else if (req.HttpMethod == "OPTIONS")
            {
                // Handle CORS preflight requests
                res.Headers.Add("Access-Control-Allow-Origin", "*");
                res.Headers.Add("Access-Control-Allow-Methods", "GET, POST, OPTIONS");
                res.Headers.Add("Access-Control-Allow-Headers", "Content-Type");
                res.StatusCode = 204;
                res.Close();
            }
            else
            {
                res.StatusCode = 404;
                res.Close();
            }
        }

        public async Task SendMessageAsync(JsonRpcMessage message, CancellationToken cancellationToken = default)
        {
            if (_sseContext == null)
            {
                OnLog?.Invoke("Cannot send message, no SSE client connected.");
                return;
            }

            try
            {
                // Serialize using JsonSerializer to string. ModelContextProtocol might provide specific JSON context, but using System.Text.Json default is usually enough.
                var json = JsonSerializer.Serialize(message);
                controller_mcp.Features.Tools.AuditLogger.LogJsonRpc("OUT", json);
                OnLog?.Invoke($"[SSE OUT PAYLOAD] {json}");
                var payload = Encoding.UTF8.GetBytes($"event: message\ndata: {json}\n\n");
                await _sseContext.Response.OutputStream.WriteAsync(payload, 0, payload.Length, cancellationToken);
                await _sseContext.Response.OutputStream.FlushAsync();
                controller_mcp.Features.Tools.AnalyticsManager.TrackBytesSent(payload.Length);
            }
            catch (Exception ex)
            {
                OnLog?.Invoke($"Failed to send message over SSE: {ex.Message}");
            }
        }

        public ValueTask DisposeAsync()
        {
            _activeClientCts?.Cancel();
            _activeClientCts?.Dispose();
            _cts?.Cancel();
            if (_sseContext != null)
            {
                _sseContext.Response.Close();
            }
            _listener.Stop();
            _listener.Close();
            _messageChannel.Writer.TryComplete();
            return default;
        }
    }
}
