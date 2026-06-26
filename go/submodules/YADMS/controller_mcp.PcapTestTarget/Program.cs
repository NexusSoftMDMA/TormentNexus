using System;
using System.Net;
using System.Net.Sockets;
using System.Text;
using System.Threading;

namespace controller_mcp.PcapTestTarget
{
    class Program
    {
        static void Main(string[] args)
        {
            Console.WriteLine("PcapTestTarget Starting...");
            int port = 13337;

            try
            {
                // Create a UDP client bound to a specific port so netstat registers it to our PID
                using (UdpClient udpClient = new UdpClient(new IPEndPoint(IPAddress.Loopback, port)))
                {
                    Console.WriteLine($"Bound to 127.0.0.1:{port}");
                    
                    IPEndPoint targetEndpoint = new IPEndPoint(IPAddress.Loopback, port);
                    byte[] payload = Encoding.UTF8.GetBytes("YADMS_PCAP_TEST_PACKET");

                    // Loop infinitely, sending a packet every 500ms
                    while (true)
                    {
                        udpClient.Send(payload, payload.Length, targetEndpoint);
                        Thread.Sleep(500);
                    }
                }
            }
            catch (Exception ex)
            {
                Console.WriteLine("Error: " + ex.Message);
            }
        }
    }
}
