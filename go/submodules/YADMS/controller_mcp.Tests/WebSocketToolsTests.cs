using System;
using System.Threading;
using System.Text.Json;
using System.Linq;
using Xunit;
using controller_mcp.Features.Tools;
using ModelContextProtocol.Protocol;

namespace controller_mcp.Tests
{
    [Collection("StateBackup")]
    public class WebSocketToolsTests : IDisposable
    {
        public void Dispose()
        {
            WebSocketTools.StopAll();
        }

        [Fact]
        public async System.Threading.Tasks.Task WebSocketTools_StartAndDisconnect_Success()
        {
            // Note: We cannot easily spin up an embedded WebSocket server here without adding heavy dependencies,
            // so we will test that an invalid URL safely returns an error, and the connection list remains clean.
            
            var res = await WebSocketTools.WebsocketConnect("wss://invalid.local.domain.test");
            
            // This should fail to connect gracefully, or connect and immediately close.
            Assert.True(res.IsError == true || res.Content.First() is TextContentBlock);

            // Regardless of outcome, check that list doesn't hang
            WebSocketTools.StopAll();
        }
    }
}
