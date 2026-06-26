using System;
using System.Text.Json;
using Xunit;
using controller_mcp.Features.Tools;

namespace controller_mcp.Tests
{
    public class InputToolsTests
    {
        [Fact]
        public void KeyboardType_EmptyString_ReturnsError()
        {
            var res = InputTools.KeyboardType("");
            Assert.True(res.IsError == true);
        }

        [Fact]
        public void ExecuteMacro_InvalidType_ReturnsError()
        {
            string json = "[{\"type\": \"INVALID_ACTION\"}]";
            var element = JsonDocument.Parse(json).RootElement;

            var res = InputTools.ExecuteMacro(element);
            Assert.True(res.IsError == true);
        }

        [Fact]
        public async System.Threading.Tasks.Task ShadowTypeFile_NonExistentFile_ReturnsError()
        {
            var res = await InputTools.ShadowTypeFile("C:\\invalid_file_12345.txt");
            Assert.True(res.IsError == true);
        }
    }
}
