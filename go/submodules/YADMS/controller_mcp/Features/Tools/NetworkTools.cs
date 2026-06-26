using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Net;
using System.Net.NetworkInformation;
using System.Text.Json;
using System.Threading.Tasks;
using ModelContextProtocol;
using ModelContextProtocol.Protocol;
using ModelContextProtocol.Server;

namespace controller_mcp.Features.Tools
{
    public static class NetworkTools
    {
        [McpServerTool, Description("Pings a host or IP address to check if it is reachable on the network.")]
        public static async Task<CallToolResult> PingHost(string host)
        {
            return await Task.Run(() =>
            {
                try
                {
                    Ping pingSender = new Ping();
                    PingReply reply = pingSender.Send(host, 4000);

                    if (reply.Status == IPStatus.Success)
                    {
                        return new CallToolResult { Content = new List<ContentBlock> { new TextContentBlock { Text = $"Ping successful. IP: {reply.Address}, RoundTrip time: {reply.RoundtripTime}ms" } } };
                    }
                    else
                    {
                        return new CallToolResult { Content = new List<ContentBlock> { new TextContentBlock { Text = $"Ping failed: {reply.Status}" } } };
                    }
                }
                catch (Exception ex)
                {
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Ping failed: {ex.Message}" } } };
                }
            });
        }

        [McpServerTool, Description("Performs a DNS lookup for a given domain name and returns the resolved IP addresses.")]
        public static async Task<CallToolResult> DnsLookup(string domain)
        {
            return await Task.Run(() =>
            {
                try
                {
                    IPHostEntry hostInfo = Dns.GetHostEntry(domain);
                    var ips = new List<string>();
                    foreach (IPAddress ip in hostInfo.AddressList)
                    {
                        ips.Add(ip.ToString());
                    }

                    string json = JsonSerializer.Serialize(ips, new JsonSerializerOptions { WriteIndented = true });
                    return new CallToolResult { Content = new List<ContentBlock> { new TextContentBlock { Text = json } } };
                }
                catch (Exception ex)
                {
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"DNS Lookup failed: {ex.Message}" } } };
                }
            });
        }
    }
}
