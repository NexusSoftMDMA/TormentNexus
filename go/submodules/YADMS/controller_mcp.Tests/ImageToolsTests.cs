using System;
using System.Threading.Tasks;
using Xunit;
using controller_mcp.Features.Tools;

namespace controller_mcp.Tests
{
    public class ImageToolsTests
    {
        [Fact]
        public async Task ConvertImage_NonExistentFile_ReturnsError()
        {
            var res = await ImageTools.ConvertImageFormat("C:\\invalid_file_12345.png", "C:\\out.jpg");
            Assert.True(res.IsError == true);
        }
    }
}
