using System;
using Xunit;
using controller_mcp.Features.Tools;

namespace controller_mcp.Tests
{
    public class RecordingToolsTests
    {
        [Fact]
        public async System.Threading.Tasks.Task RecordFullScreen_InvalidPath_ReturnsError()
        {
            var res = await RecordingTools.RecordFullScreen(0, 3, 24, "C:\\invalid_dir_12345", "test.mp4");
            Assert.True(res.IsError == true);
        }

        [Fact]
        public async System.Threading.Tasks.Task RecordWindow_NonExistentWindow_ReturnsError()
        {
            var res = await RecordingTools.RecordWindow("NON_EXISTENT_WINDOW_12345", 3, 24, "C:\\out", "test.mp4");
            Assert.True(res.IsError == true);
        }
    }
}
