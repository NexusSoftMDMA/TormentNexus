using Xunit;
using ModelContextProtocol.Protocol;
using controller_mcp.Features.Tools;

namespace controller_mcp.Tests
{
    public class PcapToolsTests
    {
        [Fact]
        public void PcapTools_StartPacketCapture_HandlesMissingNpcapGracefully()
        {
            // If Npcap is missing, it should return an error, not throw an unhandled exception
            var result = PcapTools.StartPacketCapture(0);
            
            // We expect an error if Npcap is missing, or success if it's there. 
            // The important part is it shouldn't crash the runner.
            if (result.IsError == true)
            {
                string errorText = ((TextContentBlock)result.Content[0]).Text.ToLower();
                Assert.True(errorText.Contains("capture") || errorText.Contains("wpcap") || errorText.Contains("npcap"));
            }
            else
            {
                Assert.Contains("capturing", ((TextContentBlock)result.Content[0]).Text);
                
                // Cleanup if it actually started
                string json = ((TextContentBlock)result.Content[0]).Text;
                var doc = System.Text.Json.JsonDocument.Parse(json);
                string id = doc.RootElement.GetProperty("capture_id").GetString();
                PcapTools.StopPacketCapture(id);
            }
        }
    
        [Fact] public void PcapTools_StartPacketCapture_FailsGracefullyOnInvalidDevice() { var result = PcapTools.StartPacketCapture(9999); Assert.True(result.IsError == true); }
    }
}
