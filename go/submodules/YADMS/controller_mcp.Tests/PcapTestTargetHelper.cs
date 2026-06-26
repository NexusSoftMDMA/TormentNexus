using System;
using System.Diagnostics;
using System.IO;
using System.Threading;

namespace controller_mcp.Tests
{
    public class PcapTestTargetHelper : IDisposable
    {
        public Process TargetProcess { get; private set; }

        public PcapTestTargetHelper()
        {
            string exePath = Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "..", "..", "..", "..", "controller_mcp.PcapTestTarget", "bin", "Debug", "net472", "controller_mcp.PcapTestTarget.exe");
            
            // Fallback for different build structures
            if (!File.Exists(exePath))
            {
                exePath = Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "controller_mcp.PcapTestTarget.exe");
            }

            if (!File.Exists(exePath))
            {
                throw new FileNotFoundException("PcapTestTarget binary not found!", exePath);
            }

            ProcessStartInfo psi = new ProcessStartInfo
            {
                FileName = exePath,
                UseShellExecute = false,
                RedirectStandardOutput = true,
                CreateNoWindow = true
            };

            TargetProcess = Process.Start(psi);

            // Wait for it to output "Bound to 127.0.0.1:13337" to ensure the UDP socket is open
            while (!TargetProcess.StandardOutput.EndOfStream)
            {
                string line = TargetProcess.StandardOutput.ReadLine();
                if (line != null && line.Contains("Bound to"))
                {
                    break;
                }
            }

            // Give netstat another moment to register the binding
            Thread.Sleep(1000); 
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
