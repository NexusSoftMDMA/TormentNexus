using System;
using System.Threading.Tasks;
using Xunit;
using controller_mcp.Features.Tools;

namespace controller_mcp.Tests
{
    public class VideoEditorToolsTests
    {
        [Fact]
        public async Task ConvertVideo_NonExistentFile_ReturnsError()
        {
            var res = await VideoEditorTools.ConvertVideoFormat("C:\\invalid_file_12345.mp4", "C:\\out.avi");
            Assert.True(res.IsError == true);
        }

        [Fact]
        public async Task ReplaceAudio_NonExistentFile_ReturnsError()
        {
            var res = await VideoEditorTools.ReplaceAudioInVideo("C:\\invalid_video_12345.mp4", "C:\\invalid_audio_12345.mp3", "C:\\out.mp4");
            Assert.True(res.IsError == true);
        }
    }
}
