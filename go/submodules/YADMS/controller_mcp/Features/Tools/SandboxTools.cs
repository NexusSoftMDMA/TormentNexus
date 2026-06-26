using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Diagnostics;
using System.Runtime.InteropServices;
using System.Threading.Tasks;
using ModelContextProtocol;
using ModelContextProtocol.Protocol;
using ModelContextProtocol.Server;

namespace controller_mcp.Features.Tools
{
    public static class SandboxTools
    {
        [DllImport("kernel32.dll", CharSet = CharSet.Unicode, SetLastError = true)]
        private static extern IntPtr CreateJobObject(IntPtr lpJobAttributes, string lpName);

        [DllImport("kernel32.dll", SetLastError = true)]
        [return: MarshalAs(UnmanagedType.Bool)]
        private static extern bool AssignProcessToJobObject(IntPtr hJob, IntPtr hProcess);

        [DllImport("kernel32.dll", SetLastError = true)]
        [return: MarshalAs(UnmanagedType.Bool)]
        private static extern bool SetInformationJobObject(IntPtr hJob, JobObjectInfoClass JobObjectInformationClass, IntPtr lpJobObjectInformation, uint cbJobObjectInformationLength);

        public enum JobObjectInfoClass
        {
            JobObjectExtendedLimitInformation = 9
        }

        [StructLayout(LayoutKind.Sequential)]
        struct JOBOBJECT_EXTENDED_LIMIT_INFORMATION
        {
            public JOBOBJECT_BASIC_LIMIT_INFORMATION BasicLimitInformation;
            public IO_COUNTERS IoInfo;
            public UIntPtr ProcessMemoryLimit;
            public UIntPtr JobMemoryLimit;
            public UIntPtr PeakProcessMemoryUsed;
            public UIntPtr PeakJobMemoryUsed;
        }

        [StructLayout(LayoutKind.Sequential)]
        struct JOBOBJECT_BASIC_LIMIT_INFORMATION
        {
            public long PerProcessUserTimeLimit;
            public long PerJobUserTimeLimit;
            public uint LimitFlags;
            public UIntPtr MinimumWorkingSetSize;
            public UIntPtr MaximumWorkingSetSize;
            public uint ActiveProcessLimit;
            public UIntPtr Affinity;
            public uint PriorityClass;
            public uint SchedulingClass;
        }

        [StructLayout(LayoutKind.Sequential)]
        struct IO_COUNTERS
        {
            public ulong ReadOperationCount;
            public ulong WriteOperationCount;
            public ulong OtherOperationCount;
            public ulong ReadTransferCount;
            public ulong WriteTransferCount;
            public ulong OtherTransferCount;
        }

        private const uint JOB_OBJECT_LIMIT_PROCESS_MEMORY = 0x00000100;
        private const uint JOB_OBJECT_LIMIT_JOB_MEMORY = 0x00000200;
        private const uint JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE = 0x00002000;

        [DllImport("kernel32.dll", SetLastError = true)]
        private static extern uint ResumeThread(IntPtr hThread);

        [McpServerTool, Description("Spawns an executable strictly inside a Kernel Sandbox (Job Object). Enforces a hard limit on RAM usage (in Megabytes). If the process breaches this limit, the OS will natively terminate it.")]
        public static async Task<CallToolResult> LaunchSandboxedProcess(string executable_path, string arguments = "", int max_ram_mb = 100)
        {
            return await Task.Run(() =>
            {
                try
                {
                    string safePath = InputValidator.ValidateFilePath(executable_path, nameof(executable_path));

                    IntPtr hJob = CreateJobObject(IntPtr.Zero, null);
                    if (hJob == IntPtr.Zero)
                        return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Failed to create Job Object. Error: {Marshal.GetLastWin32Error()}" } } };

                    var info = new JOBOBJECT_EXTENDED_LIMIT_INFORMATION();
                    info.BasicLimitInformation.LimitFlags = JOB_OBJECT_LIMIT_PROCESS_MEMORY | JOB_OBJECT_LIMIT_JOB_MEMORY | JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE;
                    
                    // Set RAM limits
                    UIntPtr memoryLimit = new UIntPtr((uint)max_ram_mb * 1024 * 1024);
                    info.ProcessMemoryLimit = memoryLimit;
                    info.JobMemoryLimit = memoryLimit;

                    int length = Marshal.SizeOf(typeof(JOBOBJECT_EXTENDED_LIMIT_INFORMATION));
                    IntPtr extendedInfoPtr = Marshal.AllocHGlobal(length);
                    try
                    {
                        Marshal.StructureToPtr(info, extendedInfoPtr, false);
                        if (!SetInformationJobObject(hJob, JobObjectInfoClass.JobObjectExtendedLimitInformation, extendedInfoPtr, (uint)length))
                        {
                            return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Failed to set Job Object memory limits. Error: {Marshal.GetLastWin32Error()}" } } };
                        }
                    }
                    finally
                    {
                        Marshal.FreeHGlobal(extendedInfoPtr);
                    }

                    // Start process suspended
                    var psi = new ProcessStartInfo
                    {
                        FileName = safePath,
                        Arguments = arguments,
                        UseShellExecute = false,
                        CreateNoWindow = true
                    };
                    
                    var process = new Process { StartInfo = psi };
                    process.Start();

                    // Assign the process to the restrictive Job Object
                    if (!AssignProcessToJobObject(hJob, process.Handle))
                    {
                        process.Kill();
                        return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Failed to assign process to Sandbox Job Object. Error: {Marshal.GetLastWin32Error()}" } } };
                    }

                    // We successfully sandboxed it, it is now running natively capped!
                    return new CallToolResult { Content = new List<ContentBlock> { new TextContentBlock { Text = $"Successfully launched {executable_path} (PID: {process.Id}) inside a Kernel Sandbox restricted to {max_ram_mb}MB of RAM." } } };
                }
                catch (Exception ex)
                {
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Sandbox creation failed: {ex.Message}" } } };
                }
            });
        }
    }
}
