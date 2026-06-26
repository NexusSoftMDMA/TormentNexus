using System;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.ComponentModel;
using System.Diagnostics;
using System.IO;
using System.Net.Http;
using System.Text.Json;
using System.Threading;
using System.Threading.Tasks;
using SharpPcap;
using PacketDotNet;
using ModelContextProtocol;
using ModelContextProtocol.Protocol;
using ModelContextProtocol.Server;

namespace controller_mcp.Features.Tools
{
    public class PcapSession : IDisposable
    {
        public string Id { get; set; }
        public ILiveDevice Device { get; set; }
        public ConcurrentQueue<string> OutputBuffer { get; set; } = new ConcurrentQueue<string>();
        public int TargetPid { get; set; } = -1;
        public HashSet<ushort> PortCache { get; set; } = new HashSet<ushort>();
        public CancellationTokenSource PollingCts { get; set; }

        public void Dispose()
        {
            try
            {
                PollingCts?.Cancel();
                if (Device != null)
                {
                    Device.StopCapture();
                    Device.Close();
                }
            }
            catch { }
        }
    }

    public static class PcapTools
    {
        private static readonly ConcurrentDictionary<string, PcapSession> _sessions = new ConcurrentDictionary<string, PcapSession>();

        public static void StopAll()
        {
            foreach (var kvp in _sessions)
            {
                try { kvp.Value.Dispose(); } catch { }
            }
            _sessions.Clear();
        }

        public static void RestartPacketCapture(PcapBackup backup)
        {
            try
            {
                var devices = CaptureDeviceList.Instance;
                if (backup.DeviceIndex < 0 || backup.DeviceIndex >= devices.Count)
                {
                    StateBackupManager.RemovePcap(backup.Id);
                    return;
                }

                var device = devices[backup.DeviceIndex];
                var session = new PcapSession { Id = backup.Id, Device = device, TargetPid = backup.TargetPid };

                if (backup.TargetPid > 0)
                {
                    session.PollingCts = new CancellationTokenSource();
                    Task.Run(async () =>
                    {
                        while (!session.PollingCts.IsCancellationRequested)
                        {
                            try
                            {
                                var psi = new ProcessStartInfo("netstat", "-ano") { RedirectStandardOutput = true, UseShellExecute = false, CreateNoWindow = true };
                                string output;
                                using (var proc = Process.Start(psi))
                                {
                                    output = await proc.StandardOutput.ReadToEndAsync();
                                }
                                var newPorts = new HashSet<ushort>();

                                string[] lines = output.Split(new[] { '\r', '\n' }, StringSplitOptions.RemoveEmptyEntries);
                                foreach (var line in lines)
                                {
                                    string[] parts = line.Split(new[] { ' ' }, StringSplitOptions.RemoveEmptyEntries);
                                    if (parts.Length >= 4)
                                    {
                                        string pidStr = parts[parts.Length - 1];
                                        if (int.TryParse(pidStr, out int pid) && pid == backup.TargetPid)
                                        {
                                            string localAddr = parts[1];
                                            int colonIdx = localAddr.LastIndexOf(':');
                                            if (colonIdx > 0 && ushort.TryParse(localAddr.Substring(colonIdx + 1), out ushort port))
                                            {
                                                newPorts.Add(port);
                                            }
                                        }
                                    }
                                }
                                session.PortCache = newPorts;
                            }
                            catch { }
                            await Task.Delay(1000, session.PollingCts.Token);
                        }
                    }, session.PollingCts.Token);
                }

                device.OnPacketArrival += (s, e) =>
                {
                    try
                    {
                        var packet = Packet.ParsePacket(e.GetPacket().LinkLayerType, e.GetPacket().Data);
                        var tcpPacket = packet.Extract<TcpPacket>();
                        var udpPacket = packet.Extract<UdpPacket>();
                        var ipPacket = packet.Extract<IPPacket>();

                        if (ipPacket != null)
                        {
                            ushort srcPort = tcpPacket?.SourcePort ?? udpPacket?.SourcePort ?? 0;
                            ushort dstPort = tcpPacket?.DestinationPort ?? udpPacket?.DestinationPort ?? 0;

                            if (session.TargetPid > 0)
                            {
                                var cache = session.PortCache;
                                if (!cache.Contains(srcPort) && !cache.Contains(dstPort))
                                {
                                    return;
                                }
                            }

                            string protocol = tcpPacket != null ? "TCP" : (udpPacket != null ? "UDP" : "IP");
                            string srcPortStr = srcPort > 0 ? $":{srcPort}" : "";
                            string dstPortStr = dstPort > 0 ? $":{dstPort}" : "";

                            string summary = $"[{e.GetPacket().Timeval.Date:HH:mm:ss.fff}] {protocol} {ipPacket.SourceAddress}{srcPortStr} -> {ipPacket.DestinationAddress}{dstPortStr} (Len: {e.GetPacket().Data.Length})";
                            session.OutputBuffer.Enqueue(summary);
                        }
                    }
                    catch { }
                };

                device.Open(DeviceModes.Promiscuous, 1000);
                if (!string.IsNullOrEmpty(backup.Filter))
                {
                    device.Filter = backup.Filter;
                }

                device.StartCapture();
                _sessions.TryAdd(backup.Id, session);
                AuditLogger.Log(LogLevel.INFO, "PcapTools", $"Restarted active packet capture {backup.Id} from state backup.");
            }
            catch
            {
                StateBackupManager.RemovePcap(backup.Id);
            }
        }

        [McpServerTool, Description("Downloads and installs the Npcap driver required for raw packet sniffing. Must be run as Administrator if UAC prompts.")]
        public static async Task<CallToolResult> InstallNpcap()
        {
            return await Task.Run(async () =>
            {
                try
                {
                    string url = "https://npcap.com/dist/npcap-1.79.exe";
                    string exePath = Path.Combine(Environment.CurrentDirectory, "npcap-installer.exe");

                    AuditLogger.LogSystemEvent("NpcapInstaller", "Downloading Npcap 1.79 from official servers...");

                    using (HttpClient client = new HttpClient())
                    {
                        var bytes = await client.GetByteArrayAsync(url);
                        File.WriteAllBytes(exePath, bytes);
                    }

                    AuditLogger.LogSystemEvent("NpcapInstaller", "Download complete. Launching installer...");

                    var psi = new ProcessStartInfo
                    {
                        FileName = exePath,
                        UseShellExecute = true,
                        Verb = "runas" // Elevate privileges
                    };
                    
                    var proc = Process.Start(psi);
                    if (proc != null)
                    {
                        AuditLogger.LogSystemEvent("NpcapInstaller", "Waiting for user to complete the installation wizard...");
                        proc.WaitForExit();
                        
                        bool isInstalled = File.Exists(Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.System), "Npcap", "wpcap.dll"));
                        if (isInstalled)
                        {
                            AuditLogger.LogSystemEvent("NpcapInstaller", "Success! Npcap driver was successfully installed.");
                        }
                        else
                        {
                            AuditLogger.LogSystemEvent("NpcapInstaller", "Npcap installation aborted or failed. Driver not found.");
                        }
                    }

                    return new CallToolResult { Content = new List<ContentBlock> { new TextContentBlock { Text = "Npcap driver installation completed. Please check logs for status." } } };
                }
                catch (Exception ex)
                {
                    AuditLogger.LogSystemEvent("NpcapInstaller", $"Installation failed: {ex.Message}");
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Failed to install Npcap: {ex.Message}" } } };
                }
            });
        }

        [McpServerTool, Description("Uninstalls the Npcap driver. Must be run as Administrator if UAC prompts.")]
        public static async Task<CallToolResult> UninstallNpcap()
        {
            return await Task.Run(() =>
            {
                try
                {
                    string uninstallerPath = Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.ProgramFiles), "Npcap", "uninstall.exe");
                    if (!File.Exists(uninstallerPath))
                    {
                        uninstallerPath = Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.ProgramFilesX86), "Npcap", "uninstall.exe");
                    }

                    if (!File.Exists(uninstallerPath))
                    {
                        return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = "Uninstaller not found. Npcap may not be installed." } } };
                    }

                    AuditLogger.LogSystemEvent("NpcapInstaller", "Launching Npcap uninstaller...");

                    var psi = new ProcessStartInfo
                    {
                        FileName = uninstallerPath,
                        UseShellExecute = true,
                        Verb = "runas" // Elevate privileges
                    };
                    
                    var proc = Process.Start(psi);
                    if (proc != null)
                    {
                        AuditLogger.LogSystemEvent("NpcapInstaller", "Waiting for user to complete the uninstallation wizard...");
                        proc.WaitForExit();
                        
                        bool isInstalled = File.Exists(Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.System), "Npcap", "wpcap.dll"));
                        if (!isInstalled)
                        {
                            AuditLogger.LogSystemEvent("NpcapInstaller", "Success! Npcap driver was successfully uninstalled.");
                        }
                        else
                        {
                            AuditLogger.LogSystemEvent("NpcapInstaller", "Npcap uninstallation aborted or failed. Driver still found.");
                        }
                    }

                    return new CallToolResult { Content = new List<ContentBlock> { new TextContentBlock { Text = "Npcap driver uninstallation completed. Please check logs for status." } } };
                }
                catch (Exception ex)
                {
                    AuditLogger.LogSystemEvent("NpcapInstaller", $"Uninstallation failed: {ex.Message}");
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Failed to uninstall Npcap: {ex.Message}" } } };
                }
            });
        }

        [McpServerTool, Description("Starts a background network packet capture on a specific device index. You can provide an optional BPF filter (e.g. 'tcp port 80') and an optional target_pid. If target_pid is > 0, it will only capture traffic owned by that Windows Process ID.")]
        public static CallToolResult StartPacketCapture(int device_index, string filter = "", int target_pid = -1)
        {
            try
            {
                var devices = CaptureDeviceList.Instance;
                if (devices.Count == 0)
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = "No capture devices found. Please ensure Npcap is installed." } } };

                if (device_index < 0 || device_index >= devices.Count)
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Invalid device index. Valid range: 0 to {devices.Count - 1}" } } };

                var device = devices[device_index];
                string id = Guid.NewGuid().ToString();

                var session = new PcapSession { Id = id, Device = device, TargetPid = target_pid };

                if (target_pid > 0)
                {
                    session.PollingCts = new CancellationTokenSource();
                    Task.Run(async () =>
                    {
                        while (!session.PollingCts.IsCancellationRequested)
                        {
                            try
                            {
                                var psi = new ProcessStartInfo("netstat", "-ano") { RedirectStandardOutput = true, UseShellExecute = false, CreateNoWindow = true };
                                string output;
                                using (var proc = Process.Start(psi))
                                {
                                    output = await proc.StandardOutput.ReadToEndAsync();
                                }
                                var newPorts = new HashSet<ushort>();

                                string[] lines = output.Split(new[] { '\r', '\n' }, StringSplitOptions.RemoveEmptyEntries);
                                foreach (var line in lines)
                                {
                                    string[] parts = line.Split(new[] { ' ' }, StringSplitOptions.RemoveEmptyEntries);
                                    if (parts.Length >= 4)
                                    {
                                        string pidStr = parts[parts.Length - 1];
                                        if (int.TryParse(pidStr, out int pid) && pid == target_pid)
                                        {
                                            string localAddr = parts[1];
                                            int colonIdx = localAddr.LastIndexOf(':');
                                            if (colonIdx > 0 && ushort.TryParse(localAddr.Substring(colonIdx + 1), out ushort port))
                                            {
                                                newPorts.Add(port);
                                            }
                                        }
                                    }
                                }
                                session.PortCache = newPorts;
                            }
                            catch { }
                            await Task.Delay(1000, session.PollingCts.Token);
                        }
                    }, session.PollingCts.Token);
                }

                device.OnPacketArrival += (s, e) =>
                {
                    try
                    {
                        var packet = Packet.ParsePacket(e.GetPacket().LinkLayerType, e.GetPacket().Data);
                        var tcpPacket = packet.Extract<TcpPacket>();
                        var udpPacket = packet.Extract<UdpPacket>();
                        var ipPacket = packet.Extract<IPPacket>();

                        if (ipPacket != null)
                        {
                            ushort srcPort = tcpPacket?.SourcePort ?? udpPacket?.SourcePort ?? 0;
                            ushort dstPort = tcpPacket?.DestinationPort ?? udpPacket?.DestinationPort ?? 0;

                            if (session.TargetPid > 0)
                            {
                                var cache = session.PortCache; // thread-safe reference copy
                                if (!cache.Contains(srcPort) && !cache.Contains(dstPort))
                                {
                                    return; // Drop packet, not owned by target PID
                                }
                            }

                            string protocol = tcpPacket != null ? "TCP" : (udpPacket != null ? "UDP" : "IP");
                            string srcPortStr = srcPort > 0 ? $":{srcPort}" : "";
                            string dstPortStr = dstPort > 0 ? $":{dstPort}" : "";

                            string summary = $"[{e.GetPacket().Timeval.Date:HH:mm:ss.fff}] {protocol} {ipPacket.SourceAddress}{srcPortStr} -> {ipPacket.DestinationAddress}{dstPortStr} (Len: {e.GetPacket().Data.Length})";
                            session.OutputBuffer.Enqueue(summary);
                        }
                    }
                    catch { }
                };

                device.Open(DeviceModes.Promiscuous, 1000);
                if (!string.IsNullOrEmpty(filter))
                {
                    device.Filter = filter;
                }

                device.StartCapture();
                _sessions.TryAdd(id, session);
                StateBackupManager.AddPcap(id, device_index, filter, target_pid);

                return new CallToolResult
                {
                    Content = new List<ContentBlock> { new TextContentBlock { Text = $"{{\"status\":\"capturing\", \"capture_id\":\"{id}\", \"device\":\"{device.Description}\", \"target_pid\":{(target_pid > 0 ? target_pid.ToString() : "\"none\"")}}}" } }
                };
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Failed to start capture: {ex.Message}. Make sure Npcap is installed." } } };
            }
        }

        [McpServerTool, Description("Retrieves all packet summaries captured since the last time this was called.")]
        public static CallToolResult ReceivePacketCapture(string capture_id, bool clear_buffer = true)
        {
            try
            {
                if (_sessions.TryGetValue(capture_id, out PcapSession session))
                {
                    List<string> retrievedPackets = new List<string>();

                    if (clear_buffer)
                    {
                        while (session.OutputBuffer.TryDequeue(out string pkt))
                        {
                            retrievedPackets.Add(pkt);
                        }
                    }
                    else
                    {
                        retrievedPackets.AddRange(session.OutputBuffer);
                    }

                    string json = JsonSerializer.Serialize(new 
                    {
                        status = "capturing",
                        packets_count = retrievedPackets.Count,
                        packets = retrievedPackets
                    });

                    return new CallToolResult { Content = new List<ContentBlock> { new TextContentBlock { Text = json } } };
                }

                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"No active capture session found with ID '{capture_id}'." } } };
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = ex.Message } } };
            }
        }

        [McpServerTool, Description("Stops a background packet capture.")]
        public static CallToolResult StopPacketCapture(string capture_id)
        {
            try
            {
                if (_sessions.TryRemove(capture_id, out PcapSession session))
                {
                    session.Dispose();
                    StateBackupManager.RemovePcap(capture_id);
                    return new CallToolResult { Content = new List<ContentBlock> { new TextContentBlock { Text = $"Capture {capture_id} stopped successfully." } } };
                }

                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"No active capture session found with ID '{capture_id}'." } } };
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = ex.Message } } };
            }
        }
    }
}
