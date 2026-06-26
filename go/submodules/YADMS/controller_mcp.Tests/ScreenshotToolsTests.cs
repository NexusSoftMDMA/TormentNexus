using System;
using Xunit;
using ModelContextProtocol.Protocol;
using controller_mcp.Features.Tools;

namespace controller_mcp.Tests
{
    public class ScreenshotToolsTests
    {
        [Fact]
        public void CaptureWindow_NonExistent_ReturnsError()
        {
            var res = ScreenshotTools.CaptureWindow("NON_EXISTENT_WINDOW_12345");
            Assert.True(res.IsError == true);
        }

        [Fact]
        public void CaptureScreenRegion_InvalidDimensions_ReturnsError()
        {
            var res = ScreenshotTools.CaptureScreenRegion(0, 0, -100, -100);
            Assert.True(res.IsError == true);
        }
        [Fact]
        public void CaptureScreenRegion_DirectoryTraversal_ReturnsError()
        {
            var result = ScreenshotTools.CaptureScreenRegion(0, 0, 100, 100, "../../Windows/System32", "hacked.png");
            Assert.True(result.IsError);
            Assert.Contains("directory traversal sequences", ((ModelContextProtocol.Protocol.TextContentBlock)result.Content[0]).Text);
        }
    }
}
