using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Diagnostics;
using System.IO;
using System.Runtime.InteropServices;
using System.Threading.Tasks;
using Microsoft.Diagnostics.Runtime;
using ModelContextProtocol;
using ModelContextProtocol.Protocol;
using ModelContextProtocol.Server;

namespace controller_mcp.Features.Tools
{
    public static class DumpTools
    {
        [DllImport("dbghelp.dll", EntryPoint = "MiniDumpWriteDump", CallingConvention = CallingConvention.StdCall, CharSet = CharSet.Unicode, ExactSpelling = true, SetLastError = true)]
        static extern bool MiniDumpWriteDump(IntPtr hProcess, uint processId, IntPtr hFile, uint dumpType, IntPtr expParam, IntPtr userStreamParam, IntPtr callbackParam);

        [McpServerTool, Description("Generates a MiniDump (.dmp) file for a running process, allowing you to capture its memory state.")]
        public static async Task<CallToolResult> CreateCrashDump(int pid)
        {
            return await Task.Run(() =>
            {
                try
                {
                    using (var process = Process.GetProcessById(pid))
                    {
                        string safeDir = InputValidator.ValidateFilePath(Environment.CurrentDirectory, "CurrentDirectory");
                        string safeName = InputValidator.ValidateFilePath($"dump_{process.ProcessName}_{pid}.dmp", "DumpName");
                        string dumpPath = Path.Combine(safeDir, safeName);

                        using (FileStream fs = new FileStream(dumpPath, FileMode.Create, FileAccess.ReadWrite, FileShare.Write))
                        {
                            // 2 = MiniDumpWithFullMemory
                            bool success = MiniDumpWriteDump(process.Handle, (uint)process.Id, fs.SafeFileHandle.DangerousGetHandle(), 2, IntPtr.Zero, IntPtr.Zero, IntPtr.Zero);
                            
                            if (success)
                            {
                                return new CallToolResult { Content = new List<ContentBlock> { new TextContentBlock { Text = $"Successfully generated dump file: {dumpPath}" } } };
                            }
                            else
                            {
                                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Failed to write dump. Error: {Marshal.GetLastWin32Error()}" } } };
                            }
                        }
                    }
                }
                catch (Exception ex)
                {
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Dump creation failed: {ex.Message}" } } };
                }
            });
        }

        [McpServerTool, Description("Analyzes a .NET crash dump file using ClrMD and extracts the exception stack traces.")]
        public static async Task<CallToolResult> AnalyzeDotNetDump(string dump_path)
        {
            return await Task.Run(() =>
            {
                try
                {
                    string safeDumpPath = InputValidator.ValidateFilePath(dump_path, nameof(dump_path));
                    if (!File.Exists(safeDumpPath))
                        return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = "Dump file not found." } } };

                    using (DataTarget dataTarget = DataTarget.LoadDump(safeDumpPath))
                    {
                        if (dataTarget.ClrVersions.Length == 0)
                            return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = "No .NET runtime found in the dump. This tool only supports managed (.NET) crash dumps." } } };

                        ClrRuntime runtime = dataTarget.ClrVersions[0].CreateRuntime();
                        string report = "=== .NET Crash Dump Analysis ===\n";

                        foreach (ClrThread thread in runtime.Threads)
                        {
                            if (thread.CurrentException != null)
                            {
                                report += $"\nException found on Thread {thread.OSThreadId}:\n";
                                report += $"Type: {thread.CurrentException.Type.Name}\n";
                                report += $"Message: {thread.CurrentException.Message}\n";
                                
                                foreach (var frame in thread.CurrentException.StackTrace)
                                {
                                    report += $"  at {frame.Method?.Signature}\n";
                                }
                            }
                        }

                        if (report == "=== .NET Crash Dump Analysis ===\n")
                            report += "\nNo active exceptions found on any managed thread.";

                        return new CallToolResult { Content = new List<ContentBlock> { new TextContentBlock { Text = report } } };
                    }
                }
                catch (Exception ex)
                {
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Dump analysis failed: {ex.Message}" } } };
                }
            });
        }
    }
}
