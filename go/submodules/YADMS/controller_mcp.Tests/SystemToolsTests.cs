using System;
using System.Text.Json;
using System.Threading.Tasks;
using Xunit;
using controller_mcp.Features.Tools;
using ModelContextProtocol.Protocol;
using System.Linq;

namespace controller_mcp.Tests
{
    public class SystemToolsTests
    {
        [Fact]
        public void GenerateUuid_ReturnsValidUuid()
        {
            var result = SystemTools.GenerateUuid();
            Assert.True(result.IsError == null || result.IsError == false);
            
            var textBlock = result.Content.First() as TextContentBlock;
            Assert.NotNull(textBlock);

            bool isValid = Guid.TryParse(textBlock.Text, out Guid guid);
            Assert.True(isValid);
            Assert.NotEqual(Guid.Empty, guid);
        }

        [Fact]
        public void GetTimestamps_ReturnsValidJsonStructure()
        {
            var result = SystemTools.GetTimestamps();
            Assert.True(result.IsError == null || result.IsError == false);

            var textBlock = result.Content.First() as TextContentBlock;
            Assert.NotNull(textBlock);

            var json = JsonDocument.Parse(textBlock.Text);
            Assert.True(json.RootElement.TryGetProperty("iso_8601", out _));
            Assert.True(json.RootElement.TryGetProperty("unix_epoch", out _));
            Assert.True(json.RootElement.TryGetProperty("human_readable", out _));
        }
    
        [Fact] public async Task SystemTools_GetTimestamps_HandlesExceptions() { var result = SystemTools.GetTimestamps(); Assert.False(result.IsError == true); }
    }
}
