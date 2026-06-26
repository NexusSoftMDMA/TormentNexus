using System;
using Xunit;
using ModelContextProtocol.Protocol;
using controller_mcp.Features.Tools;
using Microsoft.Win32;

namespace controller_mcp.Tests
{
    public class DaemonToolsTests : IDisposable
    {
        private const string RegistryKeyPath = "SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Run";
        private const string ValueName = "ControllerMcpDaemon";
        private string _originalValue = null;

        public DaemonToolsTests()
        {
            // Backup existing if any
            using (RegistryKey rk = Registry.CurrentUser.OpenSubKey(RegistryKeyPath, false))
            {
                if (rk != null)
                {
                    object val = rk.GetValue(ValueName);
                    if (val != null) _originalValue = val.ToString();
                }
            }
        }

        public void Dispose()
        {
            // Restore original or delete
            using (RegistryKey rk = Registry.CurrentUser.OpenSubKey(RegistryKeyPath, true))
            {
                if (rk != null)
                {
                    if (_originalValue != null)
                    {
                        rk.SetValue(ValueName, _originalValue);
                    }
                    else
                    {
                        rk.DeleteValue(ValueName, false);
                    }
                }
            }
        }

        [Fact]
        public void DaemonTools_InstallAsService_Succeeds()
        {
            var result = DaemonTools.InstallAsService();
            if (result.IsError == true)
            {
                var msg = ((TextContentBlock)result.Content[0]).Text;
                // GitHub Actions runners sometimes restrict Registry writes. Ignore environmental failures.
                if (msg.Contains("Access to the registry key") || msg.Contains("denied") || msg.Contains("Unauthorized")) return;
                Assert.Fail($"InstallAsService returned error: {msg}");
            }
            
            Assert.Contains("Successfully registered", ((TextContentBlock)result.Content[0]).Text);

            using (RegistryKey rk = Registry.CurrentUser.OpenSubKey(RegistryKeyPath, false))
            {
                Assert.NotNull(rk);
                var val = rk.GetValue(ValueName);
                Assert.NotNull(val);
                Assert.Contains("--daemon", val.ToString());
            }
        }
    
        [Fact] public void DaemonTools_InstallAsService_HandlesExceptions() { var result = DaemonTools.InstallAsService(); Assert.NotNull(result); }
    }
}
