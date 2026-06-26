#if GAME_HACKING
using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Diagnostics;
using System.IO;
using System.Runtime.InteropServices;
using System.Text;
using System.Threading.Tasks;
using ModelContextProtocol;
using ModelContextProtocol.Protocol;
using ModelContextProtocol.Server;

namespace controller_mcp.Features.Tools
{
    public static class InjectorTools
    {
        [DllImport("kernel32.dll", SetLastError = true)]
        private static extern IntPtr OpenProcess(int dwDesiredAccess, bool bInheritHandle, int dwProcessId);

        [DllImport("kernel32.dll", SetLastError = true)]
        private static extern IntPtr GetModuleHandle(string lpModuleName);

        [DllImport("kernel32.dll", SetLastError = true)]
        private static extern IntPtr GetProcAddress(IntPtr hModule, string procName);

        [DllImport("kernel32.dll", SetLastError = true, ExactSpelling = true)]
        private static extern IntPtr VirtualAllocEx(IntPtr hProcess, IntPtr lpAddress, uint dwSize, uint flAllocationType, uint flProtect);

        [DllImport("kernel32.dll", SetLastError = true)]
        private static extern bool WriteProcessMemory(IntPtr hProcess, IntPtr lpBaseAddress, byte[] lpBuffer, uint nSize, out UIntPtr lpNumberOfBytesWritten);

        [DllImport("kernel32.dll")]
        private static extern IntPtr CreateRemoteThread(IntPtr hProcess, IntPtr lpThreadAttributes, uint dwStackSize, IntPtr lpStartAddress, IntPtr lpParameter, uint dwCreationFlags, IntPtr lpThreadId);

        [DllImport("kernel32.dll", SetLastError = true)]
        [return: MarshalAs(UnmanagedType.Bool)]
        private static extern bool CloseHandle(IntPtr hObject);

        private const int PROCESS_CREATE_THREAD = 0x0002;
        private const int PROCESS_QUERY_INFORMATION = 0x0400;
        private const int PROCESS_VM_OPERATION = 0x0008;
        private const int PROCESS_VM_WRITE = 0x0020;
        private const int PROCESS_VM_READ = 0x0010;

        private const uint MEM_COMMIT = 0x00001000;
        private const uint MEM_RESERVE = 0x00002000;
        private const uint PAGE_READWRITE = 0x04;

        [McpServerTool, Description("Forcefully injects an unmanaged C++ .dll into a live running process using VirtualAllocEx and CreateRemoteThread. Used for advanced game modding, overlay injection, and live reverse engineering. WARNING: May trigger Antivirus.")]
        public static async Task<CallToolResult> InjectNativeDll(int target_pid, string dll_absolute_path)
        {
            return await Task.Run(() =>
            {
                IntPtr hProcess = IntPtr.Zero;
                IntPtr hThread = IntPtr.Zero;
                try
                {
                    string safeDllPath = InputValidator.ValidateFilePath(dll_absolute_path, nameof(dll_absolute_path));
                    if (!File.Exists(safeDllPath))
                        return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"DLL not found at path: {safeDllPath}" } } };

                    using (Process targetProcess = Process.GetProcessById(target_pid))
                    {
                        // 1. Get a handle to the target process
                        hProcess = OpenProcess(PROCESS_CREATE_THREAD | PROCESS_QUERY_INFORMATION | PROCESS_VM_OPERATION | PROCESS_VM_WRITE | PROCESS_VM_READ, false, targetProcess.Id);
                    }
                    if (hProcess == IntPtr.Zero)
                        return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Failed to open process {target_pid}. You may need Administrator privileges." } } };

                    // 2. Get address of LoadLibraryA
                    IntPtr loadLibraryAddr = GetProcAddress(GetModuleHandle("kernel32.dll"), "LoadLibraryA");
                    if (loadLibraryAddr == IntPtr.Zero)
                        return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = "Failed to find LoadLibraryA in kernel32.dll" } } };

                    // 3. Allocate memory in the target process for our DLL path
                    byte[] dllNameBytes = Encoding.ASCII.GetBytes(safeDllPath + "\0");
                    IntPtr allocMemAddress = VirtualAllocEx(hProcess, IntPtr.Zero, (uint)dllNameBytes.Length, MEM_COMMIT | MEM_RESERVE, PAGE_READWRITE);
                    if (allocMemAddress == IntPtr.Zero)
                        return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = "Failed to allocate memory in target process." } } };

                    // 4. Write the DLL path into the allocated memory
                    if (!WriteProcessMemory(hProcess, allocMemAddress, dllNameBytes, (uint)dllNameBytes.Length, out UIntPtr bytesWritten))
                        return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = "Failed to write DLL path into target process memory." } } };

                    // 5. Force the target process to spin up a thread running LoadLibraryA, pointing at our injected string
                    hThread = CreateRemoteThread(hProcess, IntPtr.Zero, 0, loadLibraryAddr, allocMemAddress, 0, IntPtr.Zero);
                    if (hThread == IntPtr.Zero)
                        return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = "Failed to create remote thread. Injection blocked (possibly by Anti-Cheat or Antivirus)." } } };

                    return new CallToolResult { Content = new List<ContentBlock> { new TextContentBlock { Text = $"Successfully injected '{safeDllPath}' into PID {target_pid}." } } };
                }
                catch (UnauthorizedAccessException uex)
                {
                    AuditLogger.Log(LogLevel.ERROR, "InjectorTools", uex.ToString());
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Access Denied: Injecting into PID {target_pid} requires Administrator privileges. {uex.Message}" } } };
                }
                catch (Exception ex)
                {
                    AuditLogger.Log(LogLevel.ERROR, "InjectorTools", ex.ToString());
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Injection failed: {ex.Message}" } } };
                }
                finally
                {
                    if (hThread != IntPtr.Zero)
                        CloseHandle(hThread);
                    if (hProcess != IntPtr.Zero)
                        CloseHandle(hProcess);
                }
            });
        }
    }
}
#endif
