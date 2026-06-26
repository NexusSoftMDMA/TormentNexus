#if GAME_HACKING
using System;
using System.Threading.Tasks;
using Xunit;
using ModelContextProtocol.Protocol;
using controller_mcp.Features.Tools;
using System.Text.Json;

namespace controller_mcp.Tests
{
    [Collection("TestTarget")]
    public class EngineToolsTests : IDisposable
    {
        private readonly TestTargetHelper _target;

        public EngineToolsTests()
        {
            _target = new TestTargetHelper(hidden: true);
        }

        public void Dispose()
        {
            _target.Dispose();
        }

        [Fact]
        public async Task EngineTools_ScanGameEngine_SucceedsOnTestTarget()
        {
            int pid = _target.TargetProcess.Id;

            var result = await EngineTools.ScanGameEngine(pid);
            Assert.True(result.IsError != true, $"Engine scan failed: {((TextContentBlock)result.Content?[0])?.Text}");
            
            string json = ((TextContentBlock)result.Content[0]).Text;
            var doc = JsonDocument.Parse(json);
            
            string engineDetected = doc.RootElement.GetProperty("engine_detected").GetString();
            Assert.Equal("Unknown", engineDetected); // TestTarget is not Unity/Unreal
        }
    
        [Fact] public async Task EngineTools_ScanGameEngine_FailsGracefullyOnInvalidPid() { var result = await EngineTools.ScanGameEngine(-9999); Assert.True(result.IsError == true); }
    }
}
#endif
