using System;
using System.Diagnostics;
using System.IO;
using System.IO.Pipes;
using System.Linq;
using System.Text;
using System.Threading;

namespace controller_mcp.Updater
{
    class Program
    {
        static void Main(string[] args)
        {
            if (args.Length < 2 || args[0] != "/install")
            {
                Console.WriteLine("YADMS Updater - Not meant to be run manually.");
                return;
            }

            int targetPid = -1;
            string installerPath = string.Empty;

            for (int i = 0; i < args.Length; i++)
            {
                if (args[i] == "/install" && i + 1 < args.Length) installerPath = args[i + 1];
                if (args[i] == "/pid" && i + 1 < args.Length) int.TryParse(args[i + 1], out targetPid);
            }

            if (string.IsNullOrEmpty(installerPath) || !File.Exists(installerPath))
            {
                Console.WriteLine("Installer payload not found.");
                return;
            }

            Console.WriteLine("Initiating YADMS Update Sequence...");

            // Step 1: Tell Daemon to Shutdown Gracefully via IPC
            bool killSent = false;
            try
            {
                using (var pipeClient = new NamedPipeClientStream(".", "ControllerMCP_IPC", PipeDirection.Out))
                {
                    pipeClient.Connect(3000);
                    using (var writer = new StreamWriter(pipeClient, Encoding.UTF8))
                    {
                        writer.WriteLine("EXIT");
                        writer.Flush();
                        killSent = true;
                        Console.WriteLine("Graceful shutdown command sent to daemon.");
                    }
                }
            }
            catch (Exception ex)
            {
                Console.WriteLine("Failed to communicate with daemon. It may already be offline. " + ex.Message);
            }

            // Step 2: Wait for the daemon process to exit to prevent file locks
            Console.WriteLine("Waiting for parent daemon process to terminate...");
            if (targetPid != -1)
            {
                try
                {
                    Process target = Process.GetProcessById(targetPid);
                    if (!target.HasExited)
                    {
                        if (killSent)
                        {
                            target.WaitForExit(10000); // Give it 10 seconds to shutdown gracefully
                        }
                        
                        if (!target.HasExited)
                        {
                            Console.WriteLine($"Force killing stuck parent process {target.Id}...");
                            target.Kill();
                        }
                    }
                }
                catch
                {
                    // Process already dead or not found, perfectly safe to proceed
                    Console.WriteLine("Parent process already terminated or not found.");
                }
            }
            
            // Extra buffer to ensure Windows releases file handles
            Thread.Sleep(1000);

            // Step 3: Launch Installer safely
            Console.WriteLine("Launching installer payload...");
            try
            {
                ProcessStartInfo psi = new ProcessStartInfo
                {
                    FileName = installerPath,
                    Arguments = "/VERYSILENT /SUPPRESSMSGBOXES /FORCECLOSEAPPLICATIONS /NORESTART",
                    UseShellExecute = true
                };

                Process.Start(psi);
            }
            catch (Exception ex)
            {
                Console.WriteLine("Failed to launch installer: " + ex.Message);
                // UAC might have been cancelled
            }
        }
    }
}
