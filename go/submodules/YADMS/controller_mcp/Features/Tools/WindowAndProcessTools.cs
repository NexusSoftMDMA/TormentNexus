using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Diagnostics;
using System.Linq;
using System.Text.Json;
using System.Threading.Tasks;
using ModelContextProtocol;
using ModelContextProtocol.Protocol;
using ModelContextProtocol.Server;

namespace controller_mcp.Features.Tools
{
    public static class WindowAndProcessTools
    {
        [McpServerTool, Description("Returns a JSON array of all currently open applications, including their Process IDs (PID) and Window Titles.")]
        public static CallToolResult ListOpenWindows()
        {
            var processes = Process.GetProcesses()
                .Where(p => !string.IsNullOrEmpty(p.MainWindowTitle))
                .Select(p => new
                {
                    PID = p.Id,
                    ProcessName = p.ProcessName,
                    WindowTitle = p.MainWindowTitle
                }).ToList();

            string json = JsonSerializer.Serialize(processes);

            return new CallToolResult
            {
                Content = new List<ContentBlock> { new TextContentBlock { Text = json } }
            };
        }

        [McpServerTool, Description("Forcefully terminates a process by its PID.")]
        public static CallToolResult KillProcess(int pid)
        {
            try
            {
                using (var process = Process.GetProcessById(pid))
                {
                    process.Kill();
                    return new CallToolResult
                    {
                        Content = new List<ContentBlock> { new TextContentBlock { Text = $"Successfully killed process {pid} ({process.ProcessName})." } }
                    };
                }
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Failed to kill process {pid}: {ex.Message}" } } };
            }
        }

        [McpServerTool, Description("Gracefully sends a close signal to a window's main UI thread by PID.")]
        public static CallToolResult CloseWindow(int pid)
        {
            try
            {
                using (var process = Process.GetProcessById(pid))
                {
                    bool closed = process.CloseMainWindow();
                    if (closed)
                    {
                        return new CallToolResult
                        {
                            Content = new List<ContentBlock> { new TextContentBlock { Text = $"Successfully sent close signal to process {pid} ({process.ProcessName})." } }
                        };
                    }
                    else
                    {
                        return new CallToolResult
                        {
                            Content = new List<ContentBlock> { new TextContentBlock { Text = $"Process {pid} has no main window or did not respond to the close signal." } }
                        };
                    }
                }
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Failed to close window for process {pid}: {ex.Message}" } } };
            }
        }

        [McpServerTool, Description("Checks what process is currently using a specific local port (e.g., 3000) and returns its PID and Name.")]
        public static async Task<CallToolResult> CheckPortStatus(int port)
        {
            ProcessStartInfo psi = new ProcessStartInfo
            {
                FileName = "cmd.exe",
                Arguments = $"/c netstat -ano | findstr :{port}",
                RedirectStandardOutput = true,
                UseShellExecute = false,
                CreateNoWindow = true
            };

            using (Process process = new Process { StartInfo = psi })
            {
                process.Start();
                string output = await process.StandardOutput.ReadToEndAsync();
                process.WaitForExit();

                if (string.IsNullOrWhiteSpace(output))
                {
                    return new CallToolResult
                    {
                        Content = new List<ContentBlock> { new TextContentBlock { Text = $"Port {port} is not currently in use." } }
                    };
                }

                // Parse netstat output to find PID (last column)
                // Example line:  TCP    0.0.0.0:3000      0.0.0.0:0              LISTENING       1234
                string[] lines = output.Split(new[] { '\r', '\n' }, StringSplitOptions.RemoveEmptyEntries);
                List<object> results = new List<object>();

                foreach (var line in lines)
                {
                    if (line.Contains($":{port} "))
                    {
                        string[] parts = line.Split(new[] { ' ' }, StringSplitOptions.RemoveEmptyEntries);
                        if (parts.Length >= 5)
                        {
                            string pidStr = parts[parts.Length - 1];
                            if (int.TryParse(pidStr, out int pid))
                            {
                                string procName = "Unknown";
                                try
                                {
                                    using (var proc = Process.GetProcessById(pid))
                                    {
                                        procName = proc.ProcessName;
                                    }
                                }
                                catch { }

                                results.Add(new
                                {
                                    Protocol = parts[0],
                                    LocalAddress = parts[1],
                                    State = parts.Length > 4 ? parts[3] : "",
                                    PID = pid,
                                    ProcessName = procName
                                });
                            }
                        }
                    }
                }

                string json = JsonSerializer.Serialize(results);
                return new CallToolResult
                {
                    Content = new List<ContentBlock> { new TextContentBlock { Text = json } }
                };
            }
        }

        [McpServerTool, Description("Starts an application detached from the MCP server so it continues running even if the server restarts. Useful for launching VS Code or browsers.")]
        public static CallToolResult StartDetachedProcess(string executable_path, string arguments = "")
        {
            AuditLogger.Log(LogLevel.DEBUG, "ExecutionStart", $"StartDetachedProcess: {executable_path}");
            var sw = System.Diagnostics.Stopwatch.StartNew();
            try
            {
                string safePath = InputValidator.ValidateFilePath(executable_path, nameof(executable_path));

                if (!System.IO.File.Exists(safePath))
                {
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Executable not found at {safePath}" } } };
                }
                var psi = new ProcessStartInfo
                {
                    FileName = safePath,
                    Arguments = arguments,
                    UseShellExecute = true
                };
                Process.Start(psi);
                sw.Stop();
                AuditLogger.Log(LogLevel.DEBUG, "ExecutionEnd", $"StartDetachedProcess completed in {sw.ElapsedMilliseconds}ms");
                return new CallToolResult { Content = new List<ContentBlock> { new TextContentBlock { Text = $"Successfully launched {executable_path} detached." } } };
            }
            catch (System.ComponentModel.Win32Exception wex)
            {
                AuditLogger.Log(LogLevel.ERROR, "WindowAndProcessTools", $"Win32 Exception starting process: {wex}");
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"OS blocked the execution or invalid executable: {wex.Message}" } } };
            }
            catch (Exception ex)
            {
                AuditLogger.Log(LogLevel.ERROR, "WindowAndProcessTools", ex.ToString());
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Failed to start detached process: {ex.Message}" } } };
            }
        }

        [McpServerTool, Description("Gets deep diagnostics about a running process, including memory, loaded modules (.dll), and threads.")]
        public static CallToolResult GetProcessDetails(int pid)
        {
            try
            {
                using (var p = Process.GetProcessById(pid))
                {
                    var details = new
                    {
                        Name = p.ProcessName,
                        Id = p.Id,
                        WorkingSet64 = p.WorkingSet64,
                        PrivateMemorySize64 = p.PrivateMemorySize64,
                        ThreadsCount = p.Threads.Count,
                        StartTime = p.StartTime.ToString("o")
                    };

                    return new CallToolResult
                    {
                        Content = new List<ContentBlock> { new TextContentBlock { Text = System.Text.Json.JsonSerializer.Serialize(details, new System.Text.Json.JsonSerializerOptions { WriteIndented = true }) } }
                    };
                }
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Failed to get process details: {ex.Message}" } } };
            }
        }
    }
}
