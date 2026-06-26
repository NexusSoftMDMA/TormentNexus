using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.IO;
using System.Security.Cryptography;
using System.Text;
using System.Text.RegularExpressions;
using System.Threading.Tasks;
using System.ComponentModel;
using ModelContextProtocol;
using ModelContextProtocol.Protocol;
using ModelContextProtocol.Server;
using System.Text.Json;

namespace controller_mcp.Features.Tools
{
    public class AudioDevice
    {
        public string Name { get; set; }
        public string AlternativeName { get; set; }
        public string Id { get; set; }
    }

    public static class AudioTools
    {
        private static string ComputeHash(string input)
        {
            using (var sha = SHA256.Create())
            {
                byte[] hash = sha.ComputeHash(Encoding.UTF8.GetBytes(input));
                return BitConverter.ToString(hash).Replace("-", "").Substring(0, 8).ToLower();
            }
        }

        public static async Task<List<AudioDevice>> GetAudioDevicesAsync()
        {
            var devices = new List<AudioDevice>();
            
            string ffmpegPath = FFmpegHelper.GetFFmpegPath();
            var psi = new ProcessStartInfo
            {
                FileName = ffmpegPath,
                Arguments = "-list_devices true -f dshow -i dummy",
                UseShellExecute = false,
                CreateNoWindow = true,
                RedirectStandardError = true,
                RedirectStandardOutput = true
            };

            using (var process = new Process { StartInfo = psi, EnableRaisingEvents = true })
            {
                var tcs = new TaskCompletionSource<bool>();
                process.Exited += (s, e) => tcs.TrySetResult(true);
                process.Start();

                var errorTask = process.StandardError.ReadToEndAsync();
                await tcs.Task;
                string stderr = await errorTask;

                // FFmpeg exits with an error when -i dummy is used, so we ignore the exit code.
                var lines = stderr.Split(new[] { '\r', '\n' }, StringSplitOptions.RemoveEmptyEntries);
                
                string currentDevice = null;
                foreach (var line in lines)
                {
                    if (line.EndsWith("(audio)"))
                    {
                        var match = Regex.Match(line, @"\]\s+""([^""]+)""\s+\(audio\)");
                        if (match.Success)
                        {
                            currentDevice = match.Groups[1].Value;
                        }
                    }
                    else if (currentDevice != null && line.Contains("Alternative name"))
                    {
                        var match = Regex.Match(line, @"Alternative name\s+""([^""]+)""");
                        string altName = match.Success ? match.Groups[1].Value : "";
                        
                        devices.Add(new AudioDevice
                        {
                            Name = currentDevice,
                            AlternativeName = altName,
                            Id = ComputeHash(currentDevice)
                        });
                        
                        currentDevice = null;
                    }
                }
            }

            return devices;
        }

        public static async Task<string> ResolveAudioDeviceNameAsync(string idOrName)
        {
            if (string.IsNullOrEmpty(idOrName)) return null;

            var devices = await GetAudioDevicesAsync();
            foreach (var d in devices)
            {
                if (string.Equals(d.Id, idOrName, StringComparison.OrdinalIgnoreCase) ||
                    string.Equals(d.Name, idOrName, StringComparison.OrdinalIgnoreCase))
                {
                    return d.Name;
                }
            }
            return idOrName; // Fallback to raw string
        }

        [McpServerTool, Description("Lists all available DirectShow audio recording devices on the system, including microphones and virtual audio cables (Voicemeeter).")]
        public static async Task<CallToolResult> ListAudioDevices(IProgress<ProgressNotificationValue> progress = null)
        {
            try
            {
                var devices = await GetAudioDevicesAsync();
                string json = JsonSerializer.Serialize(devices, new JsonSerializerOptions { WriteIndented = true });
                
                return new CallToolResult
                {
                    Content = new List<ContentBlock>
                    {
                        new TextContentBlock { Text = json }
                    }
                };
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = ex.Message } } };
            }
        }

        [McpServerTool, Description("Records audio-only to an MP3 file using the specified audio device (ID or Exact Name). Max 180s.")]
        public static async Task<CallToolResult> RecordAudio(
            string audio_device_id_or_name, 
            int duration_s = 5, 
            string save_directory = null, 
            string filename = null, 
            IProgress<ProgressNotificationValue> progress = null)
        {
            try
            {
                if (string.IsNullOrEmpty(save_directory) || string.IsNullOrEmpty(filename))
                {
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = "save_directory and filename are required to record MP3 audio." } } };
                }

                if (!Directory.Exists(save_directory))
                    Directory.CreateDirectory(save_directory);

                if (!filename.EndsWith(".mp3", StringComparison.OrdinalIgnoreCase))
                    filename += ".mp3";

                string safeDir = InputValidator.ValidateFilePath(save_directory, nameof(save_directory));
                string safeFile = InputValidator.ValidateFilePath(filename, nameof(filename));

                string savePath = Path.Combine(safeDir, safeFile);
                string exactName = await ResolveAudioDeviceNameAsync(audio_device_id_or_name);

                string args = $"-y -f dshow -i audio=\"{exactName}\" -t {duration_s} -c:a libmp3lame -q:a 2 \"{savePath}\"";

                if (progress != null)
                    progress.Report(new ProgressNotificationValue { Progress = 10, Message = $"Capturing audio from {exactName}..." });

                await FFmpegHelper.RunFFmpegAsync(args);

                if (progress != null)
                    progress.Report(new ProgressNotificationValue { Progress = 100, Message = "Complete" });

                return new CallToolResult
                {
                    Content = new List<ContentBlock>
                    {
                        new TextContentBlock { Text = $"Audio successfully saved to: {savePath}" }
                    }
                };
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = ex.Message } } };
            }
        }
    }
}
