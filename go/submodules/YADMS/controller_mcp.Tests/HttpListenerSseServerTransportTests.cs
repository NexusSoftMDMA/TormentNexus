using System;
using System.Net.Sockets;
using System.Text;
using System.Threading;
using System.Threading.Tasks;
using Xunit;
using controller_mcp;

namespace controller_mcp.Tests
{
    public class HttpListenerSseServerTransportTests
    {
        [Fact]
        public async Task HandleRequestAsync_IncompletePost_DoesNotCrash()
        {
            // Start the transport on a random available port
            int port = 51234;
            var transport = new HttpListenerSseServerTransport(port);
            
            bool errorLogged = false;
            transport.OnLog = (msg) => 
            {
                if (msg.Contains("[HTTP IN PAYLOAD ERROR]"))
                {
                    errorLogged = true;
                }
            };

            var cts = new CancellationTokenSource();
            await transport.StartAsync(cts.Token);

            // Give the server a moment to start listening
            await Task.Delay(500);

            // Connect raw TCP client to simulate incomplete HTTP POST
            using (var tcpClient = new TcpClient())
            {
                await tcpClient.ConnectAsync("127.0.0.1", port);
                var stream = tcpClient.GetStream();

                // Send incomplete HTTP POST request
                string payload = "POST /mcp/messages HTTP/1.1\r\n" +
                                 "Host: localhost:" + port + "\r\n" +
                                 "Content-Length: 1000\r\n" +
                                 "Content-Type: application/json\r\n\r\n" +
                                 "{\"partial_json\": \"...";

                byte[] buffer = Encoding.UTF8.GetBytes(payload);
                await stream.WriteAsync(buffer, 0, buffer.Length);
                await stream.FlushAsync();

                // Forcefully drop the socket while the server is awaiting ReadToEndAsync
                tcpClient.Client.Close();
            }

            // Wait for the server to process the drop
            await Task.Delay(1000);

            // Stop transport
            cts.Cancel();
            await transport.DisposeAsync();

            // The exception should have been caught and logged, preventing a crash
            Assert.True(errorLogged, "The server should have caught the stream failure and logged it.");
        }
    
        [Fact] public async Task SseTransport_StartAsync_HandlesExceptions() { var transport = new HttpListenerSseServerTransport(-1); await Assert.ThrowsAnyAsync<Exception>(async () => await transport.StartAsync(new System.Threading.CancellationToken())); }
    }
}
