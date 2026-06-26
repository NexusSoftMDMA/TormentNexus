using System;
using System.Threading.Tasks;
using Xunit;
using ModelContextProtocol.Protocol;
using controller_mcp.Features.Tools;

namespace controller_mcp.Tests
{
    public class DeveloperToolsTests
    {
        [Fact]
        public async Task ComputeFileHash_NonExistentFile_ReturnsError()
        {
            var res = await DeveloperTools.ComputeFileHash(System.IO.Path.Combine(System.IO.Path.GetTempPath(), System.Guid.NewGuid().ToString() + ".txt"));
            Assert.True(res.IsError == true);
        }

        [Fact]
        public async Task ZipDirectory_NonExistentDir_ReturnsError()
        {
            var res = await DeveloperTools.ZipDirectory(System.IO.Path.Combine(System.IO.Path.GetTempPath(), System.Guid.NewGuid().ToString()), "C:\\out.zip");
            Assert.True(res.IsError == true);
        }

        [Fact]
        public async Task UnzipArchive_NonExistentZip_ReturnsError()
        {
            var res = await DeveloperTools.UnzipArchive(System.IO.Path.Combine(System.IO.Path.GetTempPath(), System.Guid.NewGuid().ToString() + ".zip"), "C:\\out");
            Assert.True(res.IsError == true);
        }
        [Fact]
        public async Task ZipDirectory_DirectoryTraversal_ReturnsError()
        {
            var result = await DeveloperTools.ZipDirectory("../../Windows/System32", "C:\\temp\\out.zip");
            Assert.True(result.IsError);
            Assert.Contains("directory traversal sequences", ((ModelContextProtocol.Protocol.TextContentBlock)result.Content[0]).Text);
        }
    }
}
