using System;
using System.Threading.Tasks;
using Xunit;
using ModelContextProtocol.Protocol;
using controller_mcp.Features.Tools;

namespace controller_mcp.Tests
{
    public class AudioToolsTests
    {
        [Fact]
        public async Task RecordAudio_InvalidPath_ReturnsError()
        {
            var res = await AudioTools.RecordAudio("dummy_device", 1, "C:\\invalid_dir_12345", "test.mp3");
            Assert.True(res.IsError == true);
        }
        [Fact]
        public async Task RecordAudio_DirectoryTraversal_ReturnsError()
        {
            var result = await AudioTools.RecordAudio("test", 5, "../../Windows/System32", "malware.mp3", null);
            Assert.True(result.IsError);
            Assert.Contains("directory traversal sequences", ((ModelContextProtocol.Protocol.TextContentBlock)result.Content[0]).Text);
        }
    
        
    }
}
