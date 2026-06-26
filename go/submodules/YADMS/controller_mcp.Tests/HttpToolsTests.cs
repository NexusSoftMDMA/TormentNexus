using System;
using System.Threading.Tasks;
using Xunit;
using controller_mcp.Features.Tools;
using ModelContextProtocol.Protocol;

namespace controller_mcp.Tests
{
    public class HttpToolsTests
    {
        [Fact]
        public async Task MakeHttpRequest_InvalidUrl_ReturnsError()
        {
            var res = await HttpTools.MakeHttpRequest("GET", "http://invalid.local.domain.test");
            Assert.True(res.IsError == true);
        }

        [Fact]
        public async Task DownloadFile_InvalidUrl_ReturnsError()
        {
            var res = await HttpTools.DownloadFile("http://invalid.local.domain.test/file.zip", "C:\\out.zip");
            Assert.True(res.IsError == true);
        }
        [Fact]
        public async Task MakeHttpRequest_InvalidScheme_ReturnsError()
        {
            var result = await HttpTools.MakeHttpRequest("GET", "ftp://localhost/malware");
            Assert.True(result.IsError);
            Assert.Contains("invalid scheme", ((TextContentBlock)result.Content[0]).Text);
        }

        [Fact]
        public async Task DownloadFile_InvalidScheme_ReturnsError()
        {
            var result = await HttpTools.DownloadFile("javascript:alert(1)", "C:\\temp\\file.txt", null);
            Assert.True(result.IsError);
            Assert.Contains("invalid scheme", ((TextContentBlock)result.Content[0]).Text);
        }

        [Fact]
        public async Task DownloadFile_DirectoryTraversalDest_ReturnsError()
        {
            var result = await HttpTools.DownloadFile("https://example.com/file", "../../system32/cmd.exe", null);
            Assert.True(result.IsError);
            Assert.Contains("directory traversal sequences", ((TextContentBlock)result.Content[0]).Text);
        }
    }
}
