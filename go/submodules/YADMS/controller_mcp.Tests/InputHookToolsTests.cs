using System.Threading.Tasks;
using Xunit;
using ModelContextProtocol.Protocol;
using controller_mcp.Features.Tools;

namespace controller_mcp.Tests
{
    public class InputHookToolsTests
    {
        [Fact]
        public void InputHookTools_FreezeAndUnfreeze_Succeeds()
        {
            // Freeze for 1 second
            var freezeResult = InputHookTools.FreezeHardwareInput(1);
            Assert.True(freezeResult.IsError != true);
            Assert.Contains("successfully frozen", ((TextContentBlock)freezeResult.Content[0]).Text);

            // Manual unfreeze
            var unfreezeResult = InputHookTools.UnfreezeHardwareInput();
            Assert.True(unfreezeResult.IsError != true);
            Assert.Contains("unfrozen", ((TextContentBlock)unfreezeResult.Content[0]).Text);
        }
    
        [Fact] public void InputHookTools_Freeze_FailsGracefullyOnInvalidDuration() { var result = InputHookTools.FreezeHardwareInput(999); Assert.True(result.IsError == true); }
    }
}
