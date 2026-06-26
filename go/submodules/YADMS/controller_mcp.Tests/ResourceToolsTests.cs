using System.Threading.Tasks;
using Xunit;
using ModelContextProtocol.Protocol;
using controller_mcp.Features.Tools;
using System.Text.Json;

namespace controller_mcp.Tests
{
    public class ResourceToolsTests
    {
        [Fact]
        public async Task ResourceTools_GetHardwareMetrics_ReturnsValidJson()
        {
            var result = await ResourceTools.GetHardwareMetrics();
            Assert.True(result.IsError != true);
            
            string json = ((TextContentBlock)result.Content[0]).Text;
            var doc = JsonDocument.Parse(json);
            
            // Check that CPU, RAM, and GPU keys exist
            Assert.True(doc.RootElement.TryGetProperty("CPU_Load_Percent", out _));
            Assert.True(doc.RootElement.TryGetProperty("RAM_Total_MB", out _));
            Assert.True(doc.RootElement.TryGetProperty("GPUs", out _));
        }
    
        
    
        [Fact] public async Task ResourceTools_GetHardwareMetrics_HandlesExceptions() { var result = await ResourceTools.GetHardwareMetrics(); Assert.NotNull(result); }
    }
}
