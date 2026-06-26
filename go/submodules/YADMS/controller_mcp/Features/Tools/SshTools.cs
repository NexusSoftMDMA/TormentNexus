using System;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.ComponentModel;
using System.IO;
using System.Text;
using System.Text.Json;
using System.Threading;
using System.Threading.Tasks;
using Renci.SshNet;
using ModelContextProtocol;
using ModelContextProtocol.Protocol;
using ModelContextProtocol.Server;

namespace controller_mcp.Features.Tools
{
    public class SshSession : IDisposable
    {
        public string Id { get; set; }
        public string Host { get; set; }
        public SshClient Client { get; set; }
        public ShellStream Shell { get; set; }
        public ConcurrentQueue<string> OutputBuffer { get; set; } = new ConcurrentQueue<string>();
        public CancellationTokenSource Cts { get; set; }

        public void Dispose()
        {
            try
            {
                Cts?.Cancel();
                Shell?.Dispose();
                if (Client != null && Client.IsConnected)
                    Client.Disconnect();
                Client?.Dispose();
            }
            catch { }
        }
    }

    public static class SshTools
    {
        private static readonly ConcurrentDictionary<string, SshSession> _sessions = new ConcurrentDictionary<string, SshSession>();

        public static void StopAll()
        {
            foreach (var kvp in _sessions)
            {
                try { kvp.Value.Dispose(); } catch { }
            }
            _sessions.Clear();
        }

        private static async Task ReadLoop(SshSession session)
        {
            byte[] buffer = new byte[4096];
            try
            {
                while (!session.Cts.Token.IsCancellationRequested && session.Client.IsConnected && session.Shell != null)
                {
                    if (session.Shell.DataAvailable)
                    {
                        int bytesRead = await session.Shell.ReadAsync(buffer, 0, buffer.Length, session.Cts.Token);
                        if (bytesRead > 0)
                        {
                            string text = Encoding.UTF8.GetString(buffer, 0, bytesRead);
                            session.OutputBuffer.Enqueue(text);
                            while (session.OutputBuffer.Count > 10000) session.OutputBuffer.TryDequeue(out _);
                        }
                    }
                    else
                    {
                        await Task.Delay(100, session.Cts.Token);
                    }
                }
            }
            catch (OperationCanceledException) { }
            catch (Exception ex)
            {
                session.OutputBuffer.Enqueue($"[SSH READ ERROR] {ex.Message}");
            }
        }

        [McpServerTool, Description("Establishes a persistent SSH connection to a remote host, returning a connection_id. Supports either password or private_key_path. Opens a live shell stream that captures output into a mailbox.")]
        public static CallToolResult SshConnect(string host, int port, string username, string password = "", string private_key_path = "")
        {
            try
            {
                string id = Guid.NewGuid().ToString();
                
                AuthenticationMethod authMethod;
                if (!string.IsNullOrEmpty(private_key_path))
                {
                    string safeKeyPath = InputValidator.ValidateFilePath(private_key_path, nameof(private_key_path));

                    if (!File.Exists(safeKeyPath))
                        return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Private key file not found: {safeKeyPath}" } } };
                    
                    PrivateKeyFile pkf = new PrivateKeyFile(safeKeyPath);
                    authMethod = new PrivateKeyAuthenticationMethod(username, pkf);
                }
                else if (!string.IsNullOrEmpty(password))
                {
                    authMethod = new PasswordAuthenticationMethod(username, password);
                }
                else
                {
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = "Must provide either password or private_key_path." } } };
                }

                ConnectionInfo connInfo = new ConnectionInfo(host, port, username, authMethod);
                SshClient client = new SshClient(connInfo);
                client.KeepAliveInterval = TimeSpan.FromSeconds(60);
                client.Connect();

                // Create an interactive shell
                ShellStream shell = client.CreateShellStream("mcp-terminal", 80, 24, 800, 600, 1024);
                
                var session = new SshSession
                {
                    Id = id,
                    Host = host,
                    Client = client,
                    Shell = shell,
                    Cts = new CancellationTokenSource()
                };

                _sessions.TryAdd(id, session);
                StateBackupManager.AddSsh(host, port, username, password, private_key_path);

                _ = Task.Run(() => ReadLoop(session), session.Cts.Token);

                return new CallToolResult
                {
                    Content = new List<ContentBlock> { 
                        new TextContentBlock { Text = $"{{\"status\":\"connected\", \"connection_id\":\"{id}\", \"host\":\"{host}\"}}" } 
                    }
                };
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"SSH Connection failed: {ex.Message}" } } };
            }
        }

        [McpServerTool, Description("Sends a raw string command directly into the active remote SSH shell. Remember to append a newline (\\n) if you want to execute a shell command.")]
        public static CallToolResult SshSend(string connection_id, string command)
        {
            if (_sessions.TryGetValue(connection_id, out SshSession session))
            {
                if (!session.Client.IsConnected)
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"SSH Connection {connection_id} was dropped." } } };

                try
                {
                    session.Shell.Write(command);
                    session.Shell.Flush();
                    return new CallToolResult
                    {
                        Content = new List<ContentBlock> { new TextContentBlock { Text = "Data sent to SSH shell." } }
                    };
                }
                catch (Exception ex)
                {
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Send failed: {ex.Message}" } } };
                }
            }

            return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"No active SSH session found with ID '{connection_id}'." } } };
        }

        [McpServerTool, Description("Retrieves all standard output captured from the remote SSH shell since the last time this was called.")]
        public static CallToolResult SshReceive(string connection_id, bool clear_buffer = true)
        {
            if (_sessions.TryGetValue(connection_id, out SshSession session))
            {
                List<string> retrievedChunks = new List<string>();

                if (clear_buffer)
                {
                    while (session.OutputBuffer.TryDequeue(out string chunk))
                    {
                        retrievedChunks.Add(chunk);
                    }
                }
                else
                {
                    retrievedChunks.AddRange(session.OutputBuffer);
                }

                string combinedOutput = string.Join("", retrievedChunks);

                string json = JsonSerializer.Serialize(new 
                {
                    status = session.Client.IsConnected ? "connected" : "disconnected",
                    bytes_received = Encoding.UTF8.GetByteCount(combinedOutput),
                    output = combinedOutput
                });

                return new CallToolResult
                {
                    Content = new List<ContentBlock> { new TextContentBlock { Text = json } }
                };
            }

            return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"No active SSH session found with ID '{connection_id}'." } } };
        }

        [McpServerTool, Description("Gracefully terminates an active SSH session.")]
        public static CallToolResult SshDisconnect(string connection_id)
        {
            if (_sessions.TryRemove(connection_id, out SshSession session))
            {
                StateBackupManager.RemoveSsh(session.Host);
                session.Dispose();
                return new CallToolResult
                {
                    Content = new List<ContentBlock> { new TextContentBlock { Text = $"SSH connection {connection_id} disconnected." } }
                };
            }

            return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"No active SSH session found with ID '{connection_id}'." } } };
        }
    }
}
