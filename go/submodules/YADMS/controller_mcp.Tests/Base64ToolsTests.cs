using System;
using System.IO;
using Xunit;
using controller_mcp.Features.Tools;
using ModelContextProtocol.Protocol;

namespace controller_mcp.Tests
{
    public class Base64ToolsTests : IDisposable
    {
        private string _tempDir;

        public Base64ToolsTests()
        {
            _tempDir = Path.Combine(Path.GetTempPath(), "Base64Tools_Tests_" + Guid.NewGuid().ToString());
            Directory.CreateDirectory(_tempDir);
        }

        public void Dispose()
        {
            if (Directory.Exists(_tempDir))
                Directory.Delete(_tempDir, true);
        }

        [Fact]
        public void EncodeTextBase64_Success()
        {
            string original = "Hello World!";
            var res = Base64Tools.EncodeTextBase64(original);

            Assert.True(res.IsError != true);
            string base64 = ((TextContentBlock)res.Content[0]).Text;
            Assert.Equal("SGVsbG8gV29ybGQh", base64);
        }

        [Fact]
        public void DecodeTextBase64_Success()
        {
            string base64 = "SGVsbG8gV29ybGQh";
            var res = Base64Tools.DecodeTextBase64(base64);

            Assert.True(res.IsError != true);
            string decoded = ((TextContentBlock)res.Content[0]).Text;
            Assert.Equal("Hello World!", decoded);
        }

        [Fact]
        public void EncodeDecodeFile_RoundTrip_Success()
        {
            string filePath = Path.Combine(_tempDir, "test.bin");
            byte[] originalBytes = new byte[] { 0xDE, 0xAD, 0xBE, 0xEF, 0x00, 0xFF };
            File.WriteAllBytes(filePath, originalBytes);

            var encodeRes = Base64Tools.EncodeFileBase64(filePath);
            Assert.True(encodeRes.IsError != true);
            string base64 = ((TextContentBlock)encodeRes.Content[0]).Text;

            string outPath = Path.Combine(_tempDir, "out.bin");
            var decodeRes = Base64Tools.DecodeFileBase64(base64, outPath);
            Assert.True(decodeRes.IsError != true);

            byte[] decodedBytes = File.ReadAllBytes(outPath);
            Assert.Equal(originalBytes, decodedBytes);
        }

        [Fact]
        public void DecodeTextBase64_InvalidString_ReturnsError()
        {
            var res = Base64Tools.DecodeTextBase64("NOT_VALID_BASE64!!!");
            Assert.True(res.IsError == true);
            Assert.Contains("Decode failed", ((TextContentBlock)res.Content[0]).Text);
        }
        [Fact]
        public void EncodeFileBase64_DirectoryTraversal_ReturnsError()
        {
            var result = Base64Tools.EncodeFileBase64("../../Windows/System32/config/SAM");
            Assert.True(result.IsError);
            Assert.Contains("directory traversal sequences", ((TextContentBlock)result.Content[0]).Text);
        }

        [Fact]
        public void DecodeFileBase64_DirectoryTraversal_ReturnsError()
        {
            var result = Base64Tools.DecodeFileBase64("dGVzdA==", "../../malware.exe");
            Assert.True(result.IsError);
            Assert.Contains("directory traversal sequences", ((TextContentBlock)result.Content[0]).Text);
        }
    }
}
