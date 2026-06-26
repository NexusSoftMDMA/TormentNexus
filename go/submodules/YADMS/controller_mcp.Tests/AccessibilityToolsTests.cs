using System;
using System.Threading.Tasks;
using Xunit;
using ModelContextProtocol.Protocol;
using controller_mcp.Features.Tools;
using System.Text.Json;

namespace controller_mcp.Tests
{
    [Collection("TestTarget")]
    public class AccessibilityToolsTests : IDisposable
    {
        private readonly TestTargetHelper _target;

        public AccessibilityToolsTests()
        {
            // We need a visible UI for accessibility inspection
            _target = new TestTargetHelper(hidden: false);
        }

        public void Dispose()
        {
            _target.Dispose();
        }

        [Fact]
        public async Task AccessibilityTools_InspectUiTree_SucceedsOnTestTarget()
        {
            // Give window time to spawn
            await Task.Delay(1000);

            var result = await AccessibilityTools.InspectUiTree("YADMS_TEST_WINDOW");
            
            Assert.True(result.IsError != true, $"Failed to inspect UI: {((TextContentBlock)result.Content?[0])?.Text}");
            Assert.NotNull(result.Content);
            
            string json = ((TextContentBlock)result.Content[0]).Text;
            var doc = JsonDocument.Parse(json);
            
            string rootName = doc.RootElement.GetProperty("Name").GetString();
            Assert.Equal("YADMS_TEST_WINDOW", rootName);
        }
    
        [Fact] public async Task AccessibilityTools_InspectUiTree_FailsGracefullyOnInvalidWindow() { var result = await AccessibilityTools.InspectUiTree("INVALID_WINDOW_XYZ"); Assert.True(result.IsError == true); }
    }
}
