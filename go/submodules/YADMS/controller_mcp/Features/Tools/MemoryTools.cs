#if GAME_HACKING
using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Diagnostics;
using System.Runtime.InteropServices;
using System.Text;
using System.Threading.Tasks;
using ModelContextProtocol;
using ModelContextProtocol.Protocol;
using ModelContextProtocol.Server;

namespace controller_mcp.Features.Tools
{
    public static class MemoryTools
    {
        [DllImport("kernel32.dll")]
        private static extern IntPtr OpenProcess(int dwDesiredAccess, bool bInheritHandle, int dwProcessId);

        [DllImport("kernel32.dll", SetLastError = true)]
        private static extern bool ReadProcessMemory(IntPtr hProcess, IntPtr lpBaseAddress, [Out] byte[] lpBuffer, int dwSize, out IntPtr lpNumberOfBytesRead);

        [DllImport("kernel32.dll", SetLastError = true)]
        private static extern bool WriteProcessMemory(IntPtr hProcess, IntPtr lpBaseAddress, byte[] lpBuffer, int nSize, out IntPtr lpNumberOfBytesWritten);

        [DllImport("kernel32.dll")]
        private static extern int VirtualQueryEx(IntPtr hProcess, IntPtr lpAddress, out MEMORY_BASIC_INFORMATION lpBuffer, uint dwLength);

        [DllImport("kernel32.dll", SetLastError = true)]
        [return: MarshalAs(UnmanagedType.Bool)]
        private static extern bool CloseHandle(IntPtr hObject);

        private const int PROCESS_ALL_ACCESS = 0x1F0FFF;
        private const uint MEM_COMMIT = 0x1000;
        private const uint PAGE_READWRITE = 0x04;

        [StructLayout(LayoutKind.Sequential)]
        private struct MEMORY_BASIC_INFORMATION
        {
            public IntPtr BaseAddress;
            public IntPtr AllocationBase;
            public uint AllocationProtect;
            public IntPtr RegionSize;
            public uint State;
            public uint Protect;
            public uint Type;
        }

        [McpServerTool, Description("Acts like Cheat Engine: Scans the live RAM of a target process to find all memory addresses containing a specific integer value. Useful for live reverse-engineering.")]
        public static async Task<CallToolResult> ScanLiveMemoryInt32(int target_pid, int search_value)
        {
            return await Task.Run(() =>
            {
                IntPtr processHandle = IntPtr.Zero;
                try
                {
                    processHandle = OpenProcess(PROCESS_ALL_ACCESS, false, target_pid);
                    if (processHandle == IntPtr.Zero)
                        return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Failed to open process. Ensure it is running and you have Administrator rights." } } };

                    List<string> foundAddresses = new List<string>();
                    IntPtr currentAddress = IntPtr.Zero;
                    MEMORY_BASIC_INFORMATION memInfo = new MEMORY_BASIC_INFORMATION();
                    
                    byte[] searchBytes = BitConverter.GetBytes(search_value);

                    // Scan memory regions
                    while (VirtualQueryEx(processHandle, currentAddress, out memInfo, (uint)Marshal.SizeOf(typeof(MEMORY_BASIC_INFORMATION))) != 0)
                    {
                        if (memInfo.State == MEM_COMMIT && memInfo.Protect == PAGE_READWRITE)
                        {
                            byte[] buffer = new byte[(long)memInfo.RegionSize];
                            if (ReadProcessMemory(processHandle, memInfo.BaseAddress, buffer, buffer.Length, out IntPtr bytesRead))
                            {
                                for (int i = 0; i <= (int)bytesRead - searchBytes.Length; i++)
                                {
                                    if (buffer[i] == searchBytes[0] &&
                                        buffer[i + 1] == searchBytes[1] &&
                                        buffer[i + 2] == searchBytes[2] &&
                                        buffer[i + 3] == searchBytes[3])
                                    {
                                        IntPtr matchAddress = IntPtr.Add(memInfo.BaseAddress, i);
                                        foundAddresses.Add($"0x{matchAddress.ToInt64():X}");
                                        if (foundAddresses.Count >= 50) // Limit to 50 results to avoid massive JSON payloads
                                            break;
                                    }
                                }
                            }
                        }
                        
                        if (foundAddresses.Count >= 50) break;
                        currentAddress = IntPtr.Add(memInfo.BaseAddress, (int)memInfo.RegionSize);
                    }

                    string json = System.Text.Json.JsonSerializer.Serialize(new {
                        status = "scan_complete",
                        search_value = search_value,
                        matches_found = foundAddresses.Count,
                        addresses = foundAddresses
                    });

                    return new CallToolResult { Content = new List<ContentBlock> { new TextContentBlock { Text = json } } };
                }
                catch (UnauthorizedAccessException uex)
                {
                    AuditLogger.Log(LogLevel.ERROR, "MemoryTools", uex.ToString());
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Access Denied: Administrator privileges required. {uex.Message}" } } };
                }
                catch (Exception ex)
                {
                    AuditLogger.Log(LogLevel.ERROR, "MemoryTools", ex.ToString());
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Memory scan failed: {ex.Message}" } } };
                }
                finally
                {
                    if (processHandle != IntPtr.Zero) CloseHandle(processHandle);
                }
            });
        }

        [McpServerTool, Description("Acts like Cheat Engine: Overwrites a specific live memory address inside a target process with a new integer value.")]
        public static async Task<CallToolResult> WriteLiveMemoryInt32(int target_pid, string hex_address, int new_value)
        {
            return await Task.Run(() =>
            {
                IntPtr processHandle = IntPtr.Zero;
                try
                {
                    processHandle = OpenProcess(PROCESS_ALL_ACCESS, false, target_pid);
                    if (processHandle == IntPtr.Zero)
                        return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = "Failed to open process." } } };

                    long addressLong = Convert.ToInt64(hex_address, 16);
                    IntPtr address = new IntPtr(addressLong);
                    
                    byte[] newBytes = BitConverter.GetBytes(new_value);
                    
                    if (WriteProcessMemory(processHandle, address, newBytes, newBytes.Length, out IntPtr bytesWritten))
                    {
                        return new CallToolResult { Content = new List<ContentBlock> { new TextContentBlock { Text = $"Successfully wrote '{new_value}' to memory address {hex_address}." } } };
                    }
                    else
                    {
                        return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Failed to write to memory. Error: {Marshal.GetLastWin32Error()}" } } };
                    }
                }
                catch (UnauthorizedAccessException uex)
                {
                    AuditLogger.Log(LogLevel.ERROR, "MemoryTools", uex.ToString());
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Access Denied: Administrator privileges required. {uex.Message}" } } };
                }
                catch (Exception ex)
                {
                    AuditLogger.Log(LogLevel.ERROR, "MemoryTools", ex.ToString());
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Memory write failed: {ex.Message}" } } };
                }
                finally
                {
                    if (processHandle != IntPtr.Zero) CloseHandle(processHandle);
                }
            });
        }
    }
}
#endif
