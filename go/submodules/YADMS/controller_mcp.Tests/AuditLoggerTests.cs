using System;
using System.IO;
using System.Threading;
using Xunit;
using controller_mcp.Features.Tools;

namespace controller_mcp.Tests
{
    [Collection("AuditLogger")]
    public class AuditLoggerTests : IDisposable
    {
        private string _tempLogDir;

        public AuditLoggerTests()
        {
            _tempLogDir = Path.Combine(Path.GetTempPath(), "AuditLogger_Tests_" + Guid.NewGuid());
            Directory.CreateDirectory(_tempLogDir);
        }

        public void Dispose()
        {
            if (Directory.Exists(_tempLogDir))
            {
                Directory.Delete(_tempLogDir, true);
            }
        }

        [Fact]
        public void AuditLogger_Reconfigure_CreatesDirectory()
        {
            string newDir = Path.Combine(_tempLogDir, "SubLogs");
            AuditLogger.Reconfigure(newDir);

            Assert.True(Directory.Exists(newDir));
        }

        [Fact]
        public void AuditLogger_Log_WritesToDisk()
        {
            AuditLogger.Initialize(_tempLogDir);
            
            string uniqueCategory = "TEST_CATEGORY_" + Guid.NewGuid().ToString();
            AuditLogger.Log(LogLevel.INFO, uniqueCategory, "Test Message");
            
            // Allow background thread to write
            System.Threading.SpinWait.SpinUntil(() => 
            {
                string[] currentFiles = Directory.GetFiles(_tempLogDir, "*.log");
                foreach(var file in currentFiles)
                {
                    try {
                        using (var fs = new FileStream(file, FileMode.Open, FileAccess.Read, FileShare.ReadWrite))
                        using (var sr = new StreamReader(fs))
                        {
                            if (sr.ReadToEnd().Contains(uniqueCategory)) return true;
                        }
                    } catch { }
                }
                return false;
            }, 10000);

            string[] files = Directory.GetFiles(_tempLogDir, "*.log");
            Assert.True(files.Length > 0);

            bool found = false;
            foreach(var f in files) {
                using (var fs = new FileStream(f, FileMode.Open, FileAccess.Read, FileShare.ReadWrite))
                using (var sr = new StreamReader(fs))
                {
                    if (sr.ReadToEnd().Contains(uniqueCategory)) { found = true; break; }
                }
            }
            Assert.True(found, "Log was not written to disk in time.");
        }
    

    }
}
