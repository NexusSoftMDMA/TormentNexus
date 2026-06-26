using System.Threading.Tasks;
using Xunit;
using ModelContextProtocol.Protocol;
using controller_mcp.Features.Tools;

namespace controller_mcp.Tests
{
    public class NotificationToolsTests
    {
        [Fact]
        public async Task NotificationTools_ShowNotification_Succeeds()
        {
            // Note: This will actually popup a notification on the test host machine
            var result = await NotificationTools.ShowNotification("YADMS Automated Test", "This is a unit test validation popup.");
            Assert.True(result.IsError != true);
            Assert.Contains("notification sent", ((TextContentBlock)result.Content[0]).Text);
        }
    
        [Fact] public async Task NotificationTools_ShowNotification_FailsGracefullyOnExtremeInput() { var result = await NotificationTools.ShowNotification(new string('a', 50000), "Test"); Assert.False(result.IsError == true); }
    }
}
