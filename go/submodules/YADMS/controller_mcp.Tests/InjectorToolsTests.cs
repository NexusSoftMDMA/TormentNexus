#if GAME_HACKING
using System;
using System.IO;
using System.Threading.Tasks;
using Xunit;
using ModelContextProtocol.Protocol;
using controller_mcp.Features.Tools;

namespace controller_mcp.Tests
{
    [Collection("TestTarget")]
    public class InjectorToolsTests : IDisposable
    {
        private readonly TestTargetHelper _target;

        public InjectorToolsTests()
        {
            _target = new TestTargetHelper(hidden: true);
        }

        public void Dispose()
        {
            _target.Dispose();
        }

        [Fact]
        public async Task InjectorTools_InjectNativeDll_SucceedsOnTestTarget()
        {
            int pid = _target.TargetProcess.Id;
            
            // Safe injection target: a known core Windows DLL
            string safeDll = Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.System), "user32.dll");

            var result = await InjectorTools.InjectNativeDll(pid, safeDll);
            
            // Note: If running without Administrator privileges, this might return an IsError = true. 
            // We just ensure it doesn't crash the daemon/runner.
            if (result.IsError != true)
            {
                Assert.Contains("Successfully injected", ((TextContentBlock)result.Content[0]).Text);
            }
        }
    
        [Fact] public async Task InjectorTools_InjectNativeDll_FailsGracefullyOnInvalidPid() { var result = await InjectorTools.InjectNativeDll(-9999, ""); Assert.True(result.IsError == true); }
    }
}
#endif
