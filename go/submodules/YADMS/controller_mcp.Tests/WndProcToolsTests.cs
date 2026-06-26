using System;
using System.Threading.Tasks;
using Xunit;
using ModelContextProtocol.Protocol;
using controller_mcp.Features.Tools;

namespace controller_mcp.Tests
{
    [Collection("TestTarget")]
    public class WndProcToolsTests : IDisposable
    {
        private readonly TestTargetHelper _target;

        public WndProcToolsTests()
        {
            _target = new TestTargetHelper(hidden: false);
        }

        public void Dispose()
        {
            _target.Dispose();
        }

        [Fact]
        public async Task WndProcTools_SendRawWindowMessage_SucceedsOnTestTarget()
        {
            await Task.Delay(1000); // Give time for window handle to create
            
            // Send WM_NULL (0x0000), which does nothing but proves message injection works
            var result = await WndProcTools.SendRawWindowMessage("YADMS_TEST_WINDOW", 0x0000, 0, 0);
            
            Assert.True(result.IsError != true, $"Failed to inject message: {((TextContentBlock)result.Content?[0])?.Text}");
            Assert.Contains("Successfully posted message 0", ((TextContentBlock)result.Content[0]).Text);
        }
    
        [Fact] public async Task WndProcTools_SendRawWindowMessage_FailsGracefullyOnInvalidWindow() { var result = await WndProcTools.SendRawWindowMessage("INVALID_WINDOW_XYZ", 0, 0, 0); Assert.True(result.IsError == true); }
    }
}
