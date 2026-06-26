using System;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.ComponentModel;
using System.Diagnostics;
using System.IO;
using System.Text.Json;
using System.Threading.Tasks;
using ModelContextProtocol;
using ModelContextProtocol.Protocol;
using ModelContextProtocol.Server;

namespace controller_mcp.Features.Tools
{
    public class TerminalSession : IDisposable
    {
        public string Id { get; set; }
        public string Command { get; set; }
        public Process Process { get; set; }
        public ConcurrentQueue<string> OutputBuffer { get; set; } = new ConcurrentQueue<string>();
        public StreamWriter InputWriter => Process?.StandardInput;

        public void Dispose()
        {
            try
            {
                if (Process != null && !Process.HasExited)
                {
                    Process.Kill();
                    Process.Dispose();
                }
            }
            catch { }
        }
    }

    public static class TerminalTools
    {
        private static readonly ConcurrentDictionary<string, TerminalSession> _sessions = new ConcurrentDictionary<string, TerminalSession>();

        public static void StopAll()
        {
            foreach (var kvp in _sessions)
            {
                try { kvp.Value.Dispose(); } catch { }
            }
            _sessions.Clear();
        }

        [McpServerTool, Description("Starts a long-running, persistent terminal process in the background. It returns a terminal_id. The background process captures standard output and error into a mailbox.")]
        public static CallToolResult TerminalStart(string command = "cmd.exe", string working_directory = "")
        {
            try
            {
                string safeCommand = InputValidator.SanitizeCommand(command);
                string safeWorkingDir = string.IsNullOrEmpty(working_directory) ? Environment.CurrentDirectory : InputValidator.ValidateFilePath(working_directory, nameof(working_directory));

                string id = Guid.NewGuid().ToString();
                
                string exe = "cmd.exe";
                string args = "";

                if (safeCommand.StartsWith("cmd.exe ", StringComparison.OrdinalIgnoreCase))
                {
                    args = safeCommand.Substring(8);
                }
                else if (safeCommand.StartsWith("powershell.exe ", StringComparison.OrdinalIgnoreCase))
                {
                    exe = "powershell.exe";
                    args = safeCommand.Substring(15);
                }
                else if (!safeCommand.Equals("cmd.exe", StringComparison.OrdinalIgnoreCase) && !safeCommand.Equals("powershell.exe", StringComparison.OrdinalIgnoreCase))
                {
                    exe = safeCommand;
                }

                ProcessStartInfo psi = new ProcessStartInfo
                {
                    FileName = exe,
                    Arguments = args,
                    UseShellExecute = false,
                    RedirectStandardOutput = true,
                    RedirectStandardError = true,
                    RedirectStandardInput = true,
                    CreateNoWindow = true,
                    WorkingDirectory = safeWorkingDir
                };

                Process process = new Process { StartInfo = psi };
                var session = new TerminalSession { Id = id, Command = safeCommand, Process = process };

                process.OutputDataReceived += (s, e) => { if (e.Data != null) { session.OutputBuffer.Enqueue(e.Data); while (session.OutputBuffer.Count > 10000) session.OutputBuffer.TryDequeue(out _); } };
                process.ErrorDataReceived += (s, e) => { if (e.Data != null) { session.OutputBuffer.Enqueue($"[ERROR] {e.Data}"); while (session.OutputBuffer.Count > 10000) session.OutputBuffer.TryDequeue(out _); } };
                process.Exited += (s, e) => { session.OutputBuffer.Enqueue("[PROCESS EXITED]"); while (session.OutputBuffer.Count > 10000) session.OutputBuffer.TryDequeue(out _); };

                process.EnableRaisingEvents = true;
                process.Start();
                process.BeginOutputReadLine();
                process.BeginErrorReadLine();

                _sessions.TryAdd(id, session);
                StateBackupManager.AddTerminal(safeCommand, safeWorkingDir, process.Id);

                return new CallToolResult
                {
                    Content = new List<ContentBlock> { 
                        new TextContentBlock { Text = $"{{\"status\":\"started\", \"terminal_id\":\"{id}\"}}" } 
                    }
                };
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Failed to start terminal: {ex.Message}" } } };
            }
        }

        [McpServerTool, Description("Sends a string command directly into the standard input of an active background terminal process.")]
        public static CallToolResult TerminalSend(string terminal_id, string input)
        {
            if (_sessions.TryGetValue(terminal_id, out TerminalSession session))
            {
                if (session.Process.HasExited)
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Terminal {terminal_id} has already exited." } } };

                try
                {
                    session.InputWriter.WriteLine(input);
                    session.InputWriter.Flush();
                    return new CallToolResult
                    {
                        Content = new List<ContentBlock> { new TextContentBlock { Text = "Command sent." } }
                    };
                }
                catch (Exception ex)
                {
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Send failed: {ex.Message}" } } };
                }
            }

            return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"No active terminal found with ID '{terminal_id}'." } } };
        }

        [McpServerTool, Description("Retrieves all standard output and error captured from the terminal since the last time this was called.")]
        public static CallToolResult TerminalReceive(string terminal_id, bool clear_buffer = true)
        {
            if (_sessions.TryGetValue(terminal_id, out TerminalSession session))
            {
                List<string> retrievedLines = new List<string>();

                if (clear_buffer)
                {
                    while (session.OutputBuffer.TryDequeue(out string line))
                    {
                        retrievedLines.Add(line);
                    }
                }
                else
                {
                    retrievedLines.AddRange(session.OutputBuffer);
                }

                string json = JsonSerializer.Serialize(new 
                {
                    status = session.Process.HasExited ? "exited" : "running",
                    lines_count = retrievedLines.Count,
                    output = string.Join("\n", retrievedLines)
                });

                return new CallToolResult
                {
                    Content = new List<ContentBlock> { new TextContentBlock { Text = json } }
                };
            }

            return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"No active terminal found with ID '{terminal_id}'." } } };
        }

        [McpServerTool, Description("Forcefully kills an active background terminal process.")]
        public static CallToolResult TerminalKill(string terminal_id)
        {
            if (_sessions.TryRemove(terminal_id, out TerminalSession session))
            {
                StateBackupManager.RemoveTerminal(session.Command);
                session.Dispose();
                return new CallToolResult
                {
                    Content = new List<ContentBlock> { new TextContentBlock { Text = $"Terminal {terminal_id} killed successfully." } }
                };
            }

            return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"No active terminal found with ID '{terminal_id}'." } } };
        }
    }
}
