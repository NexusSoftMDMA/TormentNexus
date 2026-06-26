using Xunit;
using controller_mcp.Features.Tools;
using System.Threading.Tasks;

namespace controller_mcp.Tests
{
    public class DialogToolsTests
    {
        [Fact(Skip = "Blocks the testing thread waiting for human interaction. Skipping for CI environments.")]
        public async Task DialogTools_ShowMessageBox_Succeeds()
        {
            var result = await DialogTools.ShowMessageBox("YADMS Test", "Please click OK");
            Assert.True(result.IsError != true);
        }

        [Fact(Skip = "Blocks the testing thread waiting for human interaction. Skipping for CI environments.")]
        public async Task DialogTools_ShowInputPrompt_Succeeds()
        {
            var result = await DialogTools.ShowInputPrompt("YADMS Test", "Type something");
            Assert.True(result.IsError != true);
        }
    
        [Fact(Skip="Interactive")] public async Task DialogTools_ShowMessageBox_Negative() { var result = await DialogTools.ShowMessageBox("", ""); Assert.True(result.IsError == true); }
    }
}
