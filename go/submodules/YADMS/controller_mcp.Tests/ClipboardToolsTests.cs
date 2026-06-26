using System;
using System.Threading.Tasks;
using Xunit;
using ModelContextProtocol.Protocol;
using controller_mcp.Features.Tools;

namespace controller_mcp.Tests
{
    // Ensure clipboard tests don't run in parallel with other STA thread intensive tests if we add them
    [Collection("Sequential")]
    public class ClipboardToolsTests
    {
        [Fact]
        public async Task ClipboardTools_SetAndGetText_Succeeds()
        {
            string testText = $"YADMS_TEST_CLIPBOARD_{Guid.NewGuid()}";

            // Write to clipboard
            var setResult = await ClipboardTools.SetClipboard(testText, "text");
            Assert.True(setResult.IsError != true);

            // Read from clipboard
            var getResult = await ClipboardTools.GetClipboard();
            Assert.True(getResult.IsError != true);
            
            string json = ((TextContentBlock)getResult.Content[0]).Text;
            Assert.Contains(testText, json);
            Assert.Contains("\"type\":\"text\"", json);
        }
    
        [Fact] public async Task ClipboardTools_SetClipboard_HandlesExceptions() { var result = await ClipboardTools.SetClipboard(null); Assert.True(result.IsError == true); }
    }
}
