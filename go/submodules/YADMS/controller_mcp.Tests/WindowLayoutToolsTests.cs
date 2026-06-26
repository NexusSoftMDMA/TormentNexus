using System;
using System.Threading.Tasks;
using Xunit;
using ModelContextProtocol.Protocol;
using controller_mcp.Features.Tools;

namespace controller_mcp.Tests
{
    [Collection("TestTarget")]
    public class WindowLayoutToolsTests : IDisposable
    {
        private readonly TestTargetHelper _target;

        public WindowLayoutToolsTests()
        {
            // We need a visible window for this test to work
            _target = new TestTargetHelper(hidden: false);
        }

        public void Dispose()
        {
            _target.Dispose();
        }

        [Fact]
        public async Task WindowLayoutTools_SetWindowPosition_SucceedsOnTestTarget()
        {
            // Give it time to fully render the UI handle
            await Task.Delay(1000);

            var result = await WindowLayoutTools.SetWindowPosition("YADMS_TEST_WINDOW", 100, 100, 800, 600);
            
            Assert.True(result.IsError != true, $"Failed to move window: {((TextContentBlock)result.Content?[0])?.Text}");
            Assert.Contains("Successfully moved", ((TextContentBlock)result.Content[0]).Text);
            Assert.Contains("800x600", ((TextContentBlock)result.Content[0]).Text);
        }
    
        [Fact] public async Task WindowLayoutTools_SetWindowPosition_FailsGracefullyOnInvalidWindow() { var result = await WindowLayoutTools.SetWindowPosition("INVALID_WINDOW_XYZ", 0, 0, 100, 100); Assert.True(result.IsError == true); }
    }
}
