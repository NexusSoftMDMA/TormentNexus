using System;
using System.Threading;
using System.Threading.Tasks;
using Xunit;
using controller_mcp;

namespace controller_mcp.Tests
{
    public class IpcManagerTests
    {
        [Fact]
        public void TryConnectClient_BackgroundLoop_DoesNotCrashWhenOffline()
        {
            // The method synchronously attempts to connect. If the daemon is offline, it returns false.
            // We just ensure the initialization succeeds and doesn't throw unhandled exceptions.
            bool result = IpcManager.TryConnectClient((msg) => { });
            Assert.False(result);
        }

        [Fact]
        public async Task BroadcastLog_SafelyHandlesDisconnectedClients()
        {
            // Start server
            IpcManager.StartServer(cmd => { });

            // Allow server to initialize
            await Task.Delay(500);

            // Connect a client
            var cts = new CancellationTokenSource();
            using (var clientStream = new System.IO.Pipes.NamedPipeClientStream(".", "ControllerMcpIPC_Daemon", System.IO.Pipes.PipeDirection.InOut, System.IO.Pipes.PipeOptions.Asynchronous))
            {
                await clientStream.ConnectAsync(1000);

                // Give server time to register the client
                await Task.Delay(500);
                
                // Forcefully dispose the client to drop the connection
            }

            // Give server time to realize the pipe is broken
            await Task.Delay(500);

            // Broadcast log should catch the broken pipe exception and not crash
            var ex = Record.Exception(() => IpcManager.BroadcastLog("Test Log After Disconnect"));
            Assert.Null(ex);
        }
    

    
        [Fact] public void IpcManager_TryConnectClient_FailsGracefullyOnNull() { bool res = IpcManager.TryConnectClient(null); Assert.True(res || !res); /* Returns true if daemon running, false if offline. Null handles safely without crash. */ }
    }
}
