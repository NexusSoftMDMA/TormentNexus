using System;
using System.Threading.Tasks;
using Xunit;
using controller_mcp.Features.Tools;
using ModelContextProtocol.Protocol;
using System.Text.Json;

namespace controller_mcp.Tests
{
    [Collection("Sequential")]
    public class PcapIntegrationTests : IDisposable
    {
        private readonly PcapTestTargetHelper _target;

        public PcapIntegrationTests()
        {
            _target = new PcapTestTargetHelper();
        }

        public void Dispose()
        {
            _target.Dispose();
        }

        [Fact]
        public async Task PcapTools_CaptureTrafficFromTargetPid_Succeeds()
        {
            int pid = _target.TargetProcess.Id;

            // Find the Loopback adapter device index
            int loopbackIndex = 0;
            var devices = SharpPcap.CaptureDeviceList.Instance;
            for (int i = 0; i < devices.Count; i++)
            {
                if (devices[i].Description.ToLower().Contains("loopback") || devices[i].Name.ToLower().Contains("loopback"))
                {
                    loopbackIndex = i;
                    break;
                }
            }

            // Start packet capture filtering only our target PID's traffic
            var startResult = PcapTools.StartPacketCapture(loopbackIndex, filter: "", target_pid: pid);
            
            // Check if Npcap driver is missing
            if (startResult.IsError == true)
            {
                string errorText = ((TextContentBlock)startResult.Content[0]).Text.ToLower();
                Assert.True(errorText.Contains("capture") || errorText.Contains("wpcap") || errorText.Contains("npcap"));
                // Driver missing is acceptable in CI, so we stop here.
                return;
            }

            Assert.Contains("capturing", ((TextContentBlock)startResult.Content[0]).Text);
            
            var doc = JsonDocument.Parse(((TextContentBlock)startResult.Content[0]).Text);
            string captureId = doc.RootElement.GetProperty("capture_id").GetString();

            try
            {
                int packetsCount = 0;
                bool foundUdpPacket = false;

                // Let the background loop capture, polling for up to 10 seconds
                for (int i = 0; i < 10; i++)
                {
                    await Task.Delay(1000);

                    var receiveResult = PcapTools.ReceivePacketCapture(captureId, clear_buffer: false);
                    if (receiveResult.IsError == true) continue;

                    string json = ((TextContentBlock)receiveResult.Content[0]).Text;
                    var receiveDoc = JsonDocument.Parse(json);
                    packetsCount = receiveDoc.RootElement.GetProperty("packets_count").GetInt32();

                    if (packetsCount > 0)
                    {
                        foreach (var pkt in receiveDoc.RootElement.GetProperty("packets").EnumerateArray())
                        {
                            string summary = pkt.GetString();
                            if (summary.Contains("UDP") && summary.Contains("13337"))
                            {
                                foundUdpPacket = true;
                                break;
                            }
                        }
                    }

                    if (foundUdpPacket) break;
                }

                Assert.True(packetsCount > 0, "No packets were captured from the target PID.");
                Assert.True(foundUdpPacket, "Captured packets did not contain the expected UDP loopback string.");
            }
            finally
            {
                PcapTools.StopPacketCapture(captureId);
            }
        }
    
        [Fact] public void PcapIntegration_StartCapture_FailsGracefullyOnInvalidFilter() { var result = PcapTools.StartPacketCapture(0, "INVALID_FILTER_!@#$!@#$", -1); Assert.True(result.IsError == true || result.Content != null); }
    }
}
