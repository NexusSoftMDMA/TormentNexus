using System.Threading.Tasks;
using Xunit;
using ModelContextProtocol.Protocol;
using controller_mcp.Features.Tools;

namespace controller_mcp.Tests
{
    public class NetworkToolsTests
    {
        [Fact]
        public async Task NetworkTools_PingHost_Localhost_Succeeds()
        {
            var result = await NetworkTools.PingHost("127.0.0.1");
            Assert.True(result.IsError != true);
            Assert.Contains("Ping successful", ((TextContentBlock)result.Content[0]).Text);
        }

        [Fact]
        public async Task NetworkTools_DnsLookup_Localhost_Succeeds()
        {
            var result = await NetworkTools.DnsLookup("localhost");
            Assert.True(result.IsError != true);
            Assert.Contains("127.0.0.1", ((TextContentBlock)result.Content[0]).Text);
        }
    
        [Fact] public async Task NetworkTools_PingHost_FailsGracefullyOnInvalidHost() { var result = await NetworkTools.PingHost("invalid.local.hostname.xyz"); Assert.True(result.IsError == true); }
    }
}
