using System;
using System.Diagnostics;
using System.IO;
using System.Threading;

namespace controller_mcp.Tests
{
    public class TestTargetHelper : IDisposable
    {
        public Process TargetProcess { get; private set; }
        public IntPtr MainWindowHandle { get; private set; }

        public TestTargetHelper(bool hidden = true)
        {
            string exePath = Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "..", "..", "..", "..", "controller_mcp.TestTarget", "bin", "Debug", "net472", "controller_mcp.TestTarget.exe");
            
            // Fallback for different build structures
            if (!File.Exists(exePath))
            {
                exePath = Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "controller_mcp.TestTarget.exe");
            }

            if (!File.Exists(exePath))
            {
                throw new FileNotFoundException("TestTarget binary not found!", exePath);
            }

            ProcessStartInfo psi = new ProcessStartInfo
            {
                FileName = exePath,
                Arguments = hidden ? "--hidden" : "",
                UseShellExecute = false,
                RedirectStandardOutput = !hidden,
                CreateNoWindow = true
            };

            TargetProcess = Process.Start(psi);
            
            if (!hidden)
            {
                // Read the HWND from standard output
                while (!TargetProcess.StandardOutput.EndOfStream)
                {
                    string line = TargetProcess.StandardOutput.ReadLine();
                    if (line != null && line.StartsWith("HWND:"))
                    {
                        if (long.TryParse(line.Substring(5), out long hwnd))
                        {
                            MainWindowHandle = new IntPtr(hwnd);
                            break;
                        }
                    }
                }
            }

            // Give it time to initialize memory
            Thread.Sleep(500); 
        }

        public void Dispose()
        {
            if (TargetProcess != null && !TargetProcess.HasExited)
            {
                try
                {
                    TargetProcess.Kill();
                    TargetProcess.WaitForExit();
                }
                catch { }
            }
        }
    }
}
