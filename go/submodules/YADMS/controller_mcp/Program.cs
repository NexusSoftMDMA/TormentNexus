using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using System.Windows.Forms;

namespace controller_mcp
{
    internal static class Program
    {
        /// <summary>
        /// The main entry point for the application.
        /// </summary>
        [STAThread]
        static void Main(string[] args)
        {
            string customLogPath = null;
            bool isDaemon = false;

            for (int i = 0; i < args.Length; i++)
            {
                if (args[i] == "--daemon")
                {
                    isDaemon = true;
                }
                if (args[i] == "--log-output" && i + 1 < args.Length)
                {
                    customLogPath = args[i + 1];
                }
                if (args[i] == "--encryption-key" && i + 1 < args.Length)
                {
                    try
                    {
                        controller_mcp.Features.Tools.StateBackupManager.ImportMasterKey(args[i + 1]);
                    }
                    catch (Exception ex)
                    {
                        MessageBox.Show($"Failed to import master key from CLI: {ex.Message}", "Encryption Error", MessageBoxButtons.OK, MessageBoxIcon.Error);
                        Environment.Exit(1);
                    }
                }
            }

            var settings = AppSettings.Load();
            controller_mcp.Features.Tools.AuditLogger.Initialize(settings.LogDirectory);
            
            // Global Failsafes
            Application.ThreadException += (sender, e) => {
                controller_mcp.Features.Tools.AuditLogger.Log(controller_mcp.Features.Tools.LogLevel.FATAL, "CRITICAL_CRASH", $"UI Thread Exception: {e.Exception}");
                if (!isDaemon) {
                    MessageBox.Show($"The MCP Server encountered a fatal UI error and must close.\n\nError: {e.Exception.Message}\n\nCheck logs for details.", "Fatal Error", MessageBoxButtons.OK, MessageBoxIcon.Error);
                }
                Environment.Exit(1);
            };
            Application.SetUnhandledExceptionMode(UnhandledExceptionMode.CatchException);

            AppDomain.CurrentDomain.UnhandledException += (sender, e) => {
                var ex = e.ExceptionObject as Exception;
                controller_mcp.Features.Tools.AuditLogger.Log(controller_mcp.Features.Tools.LogLevel.FATAL, "CRITICAL_CRASH", $"Background Thread Exception: {ex}");
                if (!isDaemon && !e.IsTerminating) {
                    MessageBox.Show($"The MCP Server encountered a fatal background error.\n\nError: {ex?.Message}\n\nCheck logs for details.", "Fatal Error", MessageBoxButtons.OK, MessageBoxIcon.Error);
                }
                // We cannot stop termination of unhandled background thread exceptions in .NET, but we logged it.
            };
            
            TaskScheduler.UnobservedTaskException += (sender, e) => {
                controller_mcp.Features.Tools.AuditLogger.Log(controller_mcp.Features.Tools.LogLevel.FATAL, "CRITICAL_CRASH", $"Unobserved Task Exception: {e.Exception}");
                e.SetObserved();
            };

            Application.EnableVisualStyles();
            Application.SetCompatibleTextRenderingDefault(false);
            Application.Run(new Form1(isDaemon));
        }
    }
}
