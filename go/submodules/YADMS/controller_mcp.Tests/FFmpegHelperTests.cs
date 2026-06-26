using System;
using System.IO;
using System.Threading.Tasks;
using Xunit;
using controller_mcp.Features.Tools;

namespace controller_mcp.Tests
{
    public class FFmpegHelperTests
    {
        [Fact]
        public void FFmpegHelper_GetFFmpegPath_ReturnsValidPath()
        {
            try
            {
                string path = FFmpegHelper.GetFFmpegPath();
                Assert.True(File.Exists(path), "FFmpeg executable not found at the returned path.");
            }
            catch (FileNotFoundException)
            {
                // If not bundled and not configured, it correctly throws FileNotFoundException
            }
        }
    
        [Fact] public async Task FFmpegHelper_RunFFmpegAsync_ThrowsOnInvalidArgs() { await Assert.ThrowsAnyAsync<System.Exception>(async () => await FFmpegHelper.RunFFmpegAsync("-invalid_arg_xyz")); }
    }
}
