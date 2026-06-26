using System;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.ComponentModel;
using System.Net.WebSockets;
using System.Text;
using System.Text.Json;
using System.Threading;
using System.Threading.Tasks;
using ModelContextProtocol;
using ModelContextProtocol.Protocol;
using ModelContextProtocol.Server;

namespace controller_mcp.Features.Tools
{
    public class WebSocketSession : IDisposable
    {
        public string Id { get; set; }
        public ClientWebSocket Socket { get; set; }
        public ConcurrentQueue<string> Messages { get; set; } = new ConcurrentQueue<string>();
        public CancellationTokenSource CancellationTokenSource { get; set; } = new CancellationTokenSource();
        public Task ReceiveTask { get; set; }

        public void Dispose()
        {
            CancellationTokenSource.Cancel();
            try
            {
                if (Socket.State == WebSocketState.Open || Socket.State == WebSocketState.CloseReceived || Socket.State == WebSocketState.CloseSent)
                {
                    Socket.Abort();
                }
                Socket.Dispose();
            }
            catch { }
            CancellationTokenSource.Dispose();
        }
    }

    public static class WebSocketTools
    {
        private static readonly ConcurrentDictionary<string, WebSocketSession> _sessions = new ConcurrentDictionary<string, WebSocketSession>();

        public static void StopAll()
        {
            foreach (var kvp in _sessions)
            {
                try { kvp.Value.Dispose(); } catch { }
            }
            _sessions.Clear();
        }

        private static async Task ReceiveLoop(WebSocketSession session)
        {
            var buffer = new byte[8192];
            try
            {
                while (session.Socket.State == WebSocketState.Open && !session.CancellationTokenSource.IsCancellationRequested)
                {
                    var result = await session.Socket.ReceiveAsync(new ArraySegment<byte>(buffer), session.CancellationTokenSource.Token);

                    if (result.MessageType == WebSocketMessageType.Close)
                    {
                        session.Messages.Enqueue("[System: Connection closed by remote host]");
                        await session.Socket.CloseAsync(WebSocketCloseStatus.NormalClosure, "Closed by host", CancellationToken.None);
                        break;
                    }

                    if (result.MessageType == WebSocketMessageType.Text)
                    {
                        var sb = new StringBuilder();
                        sb.Append(Encoding.UTF8.GetString(buffer, 0, result.Count));

                        while (!result.EndOfMessage)
                        {
                            result = await session.Socket.ReceiveAsync(new ArraySegment<byte>(buffer), session.CancellationTokenSource.Token);
                            sb.Append(Encoding.UTF8.GetString(buffer, 0, result.Count));
                        }

                        session.Messages.Enqueue(sb.ToString());
                        while (session.Messages.Count > 10000) session.Messages.TryDequeue(out _);
                    }
                }
            }
            catch (OperationCanceledException) { }
            catch (Exception ex)
            {
                session.Messages.Enqueue($"[System Error: {ex.Message}]");
            }
            finally
            {
                if (_sessions.TryRemove(session.Id, out _))
                {
                    session.Dispose();
                }
            }
        }

        [McpServerTool, Description("Establishes a persistent connection to a WebSocket server and starts listening in the background. Returns a connection_id.")]
        public static async Task<CallToolResult> WebsocketConnect(string url)
        {
            try
            {
                string safeUrl = InputValidator.ValidateUrl(url, nameof(url));
                var uri = new Uri(safeUrl);
                var socket = new ClientWebSocket();
                
                await socket.ConnectAsync(uri, CancellationToken.None);

                string id = Guid.NewGuid().ToString();
                var session = new WebSocketSession
                {
                    Id = id,
                    Socket = socket
                };

                // Start the background receive loop without awaiting it
                session.ReceiveTask = Task.Run(() => ReceiveLoop(session));

                _sessions.TryAdd(id, session);
                StateBackupManager.AddWebSocket(safeUrl);

                return new CallToolResult
                {
                    Content = new List<ContentBlock> { 
                        new TextContentBlock { Text = $"{{\"status\":\"connected\", \"connection_id\":\"{id}\"}}" } 
                    }
                };
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Connection failed: {ex.Message}" } } };
            }
        }

        [McpServerTool, Description("Sends a text message over an open WebSocket connection.")]
        public static async Task<CallToolResult> WebsocketSend(string connection_id, string message)
        {
            if (_sessions.TryGetValue(connection_id, out WebSocketSession session))
            {
                if (session.Socket.State != WebSocketState.Open)
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Cannot send: Socket state is {session.Socket.State}" } } };

                try
                {
                    byte[] bytes = Encoding.UTF8.GetBytes(message);
                    await session.Socket.SendAsync(new ArraySegment<byte>(bytes), WebSocketMessageType.Text, true, CancellationToken.None);

                    return new CallToolResult
                    {
                        Content = new List<ContentBlock> { new TextContentBlock { Text = "Message sent successfully." } }
                    };
                }
                catch (Exception ex)
                {
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Failed to send: {ex.Message}" } } };
                }
            }

            return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"No active connection found with ID '{connection_id}'." } } };
        }

        [McpServerTool, Description("Retrieves all text messages received by the WebSocket since the last time this was called. Clears the queue by default.")]
        public static CallToolResult WebsocketReceive(string connection_id, bool clear_buffer = true)
        {
            if (_sessions.TryGetValue(connection_id, out WebSocketSession session))
            {
                List<string> retrievedMessages = new List<string>();

                if (clear_buffer)
                {
                    while (session.Messages.TryDequeue(out string msg))
                    {
                        retrievedMessages.Add(msg);
                    }
                }
                else
                {
                    retrievedMessages.AddRange(session.Messages);
                }

                string json = JsonSerializer.Serialize(new 
                {
                    status = session.Socket.State.ToString(),
                    message_count = retrievedMessages.Count,
                    messages = retrievedMessages
                });

                return new CallToolResult
                {
                    Content = new List<ContentBlock> { new TextContentBlock { Text = json } }
                };
            }

            return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"No active connection found with ID '{connection_id}'." } } };
        }

        [McpServerTool, Description("Gracefully shuts down a WebSocket connection and frees resources.")]
        public static async Task<CallToolResult> WebsocketClose(string connection_id)
        {
            if (_sessions.TryRemove(connection_id, out WebSocketSession session))
            {
                StateBackupManager.RemoveWebSocket(connection_id);
                try
                {
                    if (session.Socket.State == WebSocketState.Open)
                    {
                        await session.Socket.CloseAsync(WebSocketCloseStatus.NormalClosure, "Client closing connection", CancellationToken.None);
                    }
                }
                catch { }
                finally
                {
                    session.Dispose();
                }

                return new CallToolResult
                {
                    Content = new List<ContentBlock> { new TextContentBlock { Text = $"Connection {connection_id} closed." } }
                };
            }

            return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"No active connection found with ID '{connection_id}'." } } };
        }
    }
}
