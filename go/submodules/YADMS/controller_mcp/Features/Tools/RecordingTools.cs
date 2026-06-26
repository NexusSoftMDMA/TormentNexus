using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Diagnostics;
using System.Drawing;
using System.IO;
using System.Runtime.InteropServices;
using System.Text;
using System.Threading;
using System.Threading.Tasks;
using System.Windows.Forms;
using Microsoft.Extensions.Logging;
using ModelContextProtocol;
using ModelContextProtocol.Protocol;
using ModelContextProtocol.Server;

namespace controller_mcp.Features.Tools
{
    [McpServerToolType]
    public static class RecordingTools
    {
        private static void EnforceLimits(ref int duration_s, ref int fps)
        {
            if (duration_s <= 0) duration_s = 3;
            if (duration_s > 180) duration_s = 180;
            if (fps <= 0) fps = 24;
            if (fps > 60) fps = 60;
        }

        private static CallToolResult CreateGifResult(byte[] gifBytes, string save_directory = null, string filename = null)
        {
            if (!string.IsNullOrWhiteSpace(save_directory))
            {
                try
                {
                    if (!Directory.Exists(save_directory)) Directory.CreateDirectory(save_directory);
                    
                    string finalName = string.IsNullOrWhiteSpace(filename) 
                        ? $"recording_{DateTime.Now:yyyyMMdd_HHmmss}.gif" 
                        : filename;
                        
                    if (!finalName.EndsWith(".gif", StringComparison.OrdinalIgnoreCase))
                        finalName += ".gif";
                        
                    string path = Path.Combine(InputValidator.ValidateFilePath(save_directory, nameof(save_directory)), 
                                             InputValidator.ValidateFilePath(finalName, nameof(filename)));
                    File.WriteAllBytes(path, gifBytes);
                }
                catch { }
            }

            string base64 = Convert.ToBase64String(gifBytes);
            byte[] dataBytes = System.Text.Encoding.UTF8.GetBytes(base64);

            return new CallToolResult
            {
                Content = new System.Collections.Generic.List<ContentBlock>
                {
                    new ImageContentBlock
                    {
                        Data = dataBytes,
                        MimeType = "image/gif"
                    }
                }
            };
        }

        private static void LogMcp(string message, string level = "info")
        {
            try
            {
                if (Form1.Server != null)
                {
                    Form1.Server.SendNotificationAsync("notifications/message", new LoggingMessageNotificationParams
                    {
                        Level = level == "info" ? LoggingLevel.Info : (level == "error" ? LoggingLevel.Error : LoggingLevel.Debug),
                        Logger = "controller_mcp",
                        Data = System.Text.Json.JsonDocument.Parse("{\"" + "message" + "\":\"" + message.Replace("\"", "\\\"") + "\"}").RootElement
                    }).GetAwaiter().GetResult();
                }
            }
            catch { }
        }

        private static async Task<CallToolResult> RecordWithFFmpegAsync(
            string gdigrabInput, string additionalArgs, int duration_s, int fps, 
            string save_directory, string filename, string format, string audio_device_id_or_name, 
            Action onStart = null, Action onEnd = null, IProgress<ProgressNotificationValue> progress = null)
        {
            string tempMp4 = Path.Combine(Path.GetTempPath(), Guid.NewGuid().ToString() + ".mp4");
            string tempGif = Path.Combine(Path.GetTempPath(), Guid.NewGuid().ToString() + ".gif");

            try
            {
                onStart?.Invoke();

                LogMcp($"Starting MP4 capture for {duration_s} seconds...", "info");
                if (progress != null)
                {
                    progress.Report(new ProgressNotificationValue { Progress = 10, Message = "Starting capture..." });
                }

                string audioArgs = "";
                string audioCodecArgs = "";
                if (!string.IsNullOrEmpty(audio_device_id_or_name))
                {
                    string exactAudio = await AudioTools.ResolveAudioDeviceNameAsync(audio_device_id_or_name);
                    string safeAudio = InputValidator.EscapeArgument(exactAudio);
                    audioArgs = $"-f dshow -i audio={safeAudio}";
                    audioCodecArgs = "-c:a aac -b:a 128k";
                }

                string safeTempMp4 = InputValidator.EscapeArgument(tempMp4);

                // Step 1: Record to MP4
                string mp4Args = $"-y -f gdigrab -framerate {fps} {additionalArgs} -i {gdigrabInput} {audioArgs} -t {duration_s} -c:v libx264 -preset ultrafast -crf 18 {audioCodecArgs} {safeTempMp4}";
                await FFmpegHelper.RunFFmpegAsync(mp4Args);

                onEnd?.Invoke(); // Window can be minimized as soon as recording finishes

                if (format.Equals("mp4", StringComparison.OrdinalIgnoreCase))
                {
                    if (string.IsNullOrEmpty(save_directory) || string.IsNullOrEmpty(filename))
                    {
                        return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = "save_directory and filename are required to save as MP4." } } };
                    }
                    if (!Directory.Exists(save_directory))
                        Directory.CreateDirectory(save_directory);

                    if (!filename.EndsWith(".mp4", StringComparison.OrdinalIgnoreCase))
                        filename += ".mp4";

                    string savePath = Path.Combine(InputValidator.ValidateFilePath(save_directory, nameof(save_directory)), 
                                                 InputValidator.ValidateFilePath(filename, nameof(filename)));
                    File.Copy(tempMp4, savePath, true);

                    LogMcp($"MP4 saved to {savePath}", "info");
                    if (progress != null)
                        progress.Report(new ProgressNotificationValue { Progress = 100, Message = "Complete" });

                    return new CallToolResult
                    {
                        Content = new System.Collections.Generic.List<ContentBlock>
                        {
                            new TextContentBlock { Text = $"Video successfully saved to: {savePath}" }
                        }
                    };
                }

                LogMcp($"MP4 capture finished. Converting to high quality GIF...", "info");
                if (progress != null)
                {
                    progress.Report(new ProgressNotificationValue { Progress = 70, Message = "Converting to GIF..." });
                }

                // Step 2: Convert MP4 to GIF using high quality palette
                string safeTempGif = InputValidator.EscapeArgument(tempGif);
                string gifArgs = $"-y -i {safeTempMp4} -vf \"fps={fps},split[s0][s1];[s0]palettegen=stats_mode=diff[p];[s1][p]paletteuse=dither=bayer:bayer_scale=5\" -loop 0 {safeTempGif}";
                await FFmpegHelper.RunFFmpegAsync(gifArgs);

                LogMcp($"GIF conversion complete. Preparing result...", "info");
                if (progress != null)
                {
                    progress.Report(new ProgressNotificationValue { Progress = 100, Message = "Complete" });
                }

                byte[] gifBytes = File.ReadAllBytes(tempGif);
                return CreateGifResult(gifBytes, save_directory, filename);
            }
            catch (Exception ex)
            {
                LogMcp($"Recording failed: {ex.Message}", "error");
                throw;
            }
            finally
            {
                if (File.Exists(tempMp4)) File.Delete(tempMp4);
                if (File.Exists(tempGif)) File.Delete(tempGif);
            }
        }

        [McpServerTool, Description("Records an animated GIF or MP4 of all connected screens for a specified duration (max 180s, max 60fps).")]
        public static async Task<CallToolResult> RecordAllScreens(int duration_s = 3, int fps = 24, string save_directory = null, string filename = null, string format = "gif", string audio_device_id_or_name = null, IProgress<ProgressNotificationValue> progress = null)
        {
            try
            {
                if (!string.IsNullOrWhiteSpace(filename) && string.IsNullOrWhiteSpace(save_directory))
                    throw new ArgumentException("save_directory must be provided if filename is set.");
                EnforceLimits(ref duration_s, ref fps);

                Rectangle totalSize = Rectangle.Empty;
                foreach (Screen screen in Screen.AllScreens)
                    totalSize = Rectangle.Union(totalSize, screen.Bounds);

                int width = totalSize.Width;
                int height = totalSize.Height;
                width = width - (width % 2);
                height = height - (height % 2);

                return await RecordWithFFmpegAsync("desktop", $"-offset_x {SystemInformation.VirtualScreen.Left} -offset_y {SystemInformation.VirtualScreen.Top} -video_size {width}x{height}", duration_s, fps, save_directory, filename, format, audio_device_id_or_name, null, null, progress);
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new System.Collections.Generic.List<ContentBlock> { new TextContentBlock { Text = ex.ToString() } } };
            }
        }

        [McpServerTool, Description("Records an animated GIF or MP4 of a specific screen monitor (default 0). Max 180s, max 60fps.")]
        public static async Task<CallToolResult> RecordFullScreen(int screenIndex = 0, int duration_s = 3, int fps = 24, string save_directory = null, string filename = null, string format = "gif", string audio_device_id_or_name = null, IProgress<ProgressNotificationValue> progress = null)
        {
            try
            {
                if (!string.IsNullOrWhiteSpace(filename) && string.IsNullOrWhiteSpace(save_directory))
                    throw new ArgumentException("save_directory must be provided if filename is set.");
                EnforceLimits(ref duration_s, ref fps);

                var screens = Screen.AllScreens;
                if (screenIndex < 0 || screenIndex >= screens.Length)
                    screenIndex = 0;

                var bounds = screens[screenIndex].Bounds;

                int width = bounds.Width;
                int height = bounds.Height;
                width = width - (width % 2);
                height = height - (height % 2);

                return await RecordWithFFmpegAsync("desktop", $"-offset_x {bounds.Left} -offset_y {bounds.Top} -video_size {width}x{height}", duration_s, fps, save_directory, filename, format, audio_device_id_or_name, null, null, progress);
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new System.Collections.Generic.List<ContentBlock> { new TextContentBlock { Text = ex.ToString() } } };
            }
        }

        [McpServerTool, Description("Records an animated GIF or MP4 of a specific window by matching its title/exe name. Max 180s, max 60fps. Set include_borders to true to capture the window decorations.")]
        public static async Task<CallToolResult> RecordWindow(string search_term, int duration_s = 3, int fps = 24, string save_directory = null, string filename = null, bool include_borders = false, string format = "gif", string audio_device_id_or_name = null, IProgress<ProgressNotificationValue> progress = null)
        {
            try
            {
                if (!string.IsNullOrWhiteSpace(filename) && string.IsNullOrWhiteSpace(save_directory))
                    throw new ArgumentException("save_directory must be provided if filename is set.");
                EnforceLimits(ref duration_s, ref fps);

                IntPtr hwnd = IntPtr.Zero;
                string exactTitle = "";

                ScreenshotTools.EnumDelegate filter = delegate (IntPtr hWnd, int lParam)
                {
                    if (!ScreenshotTools.IsWindowVisible(hWnd)) return true;

                    StringBuilder sb = new StringBuilder(255);
                    ScreenshotTools.GetWindowText(hWnd, sb, sb.Capacity + 1);
                    string title = sb.ToString();

                    ScreenshotTools.GetWindowThreadProcessId(hWnd, out uint pid);
                    string processName = "";
                    try
                    {
                        using (var process = Process.GetProcessById((int)pid))
                        {
                            processName = process.ProcessName;
                        }
                    }
                    catch { }

                    if (title.IndexOf(search_term, StringComparison.OrdinalIgnoreCase) >= 0 ||
                        processName.IndexOf(search_term, StringComparison.OrdinalIgnoreCase) >= 0)
                    {
                        hwnd = hWnd;
                        exactTitle = title;
                        return false;
                    }
                    return true;
                };

                ScreenshotTools.EnumWindows(filter, 0);

                if (hwnd == IntPtr.Zero)
                {
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Could not find an open window matching '{search_term}'." } } };
                }

                bool wasMinimized = false;

                Action onStart = () =>
                {
                    if (ScreenshotTools.IsIconic(hwnd))
                    {
                        wasMinimized = true;
                        ScreenshotTools.ShowWindow(hwnd, ScreenshotTools.SW_RESTORE);
                        Thread.Sleep(300);
                    }
                    ScreenshotTools.SetForegroundWindow(hwnd);
                    Thread.Sleep(200);
                };

                Action onEnd = () =>
                {
                    if (wasMinimized)
                    {
                        ScreenshotTools.ShowWindow(hwnd, ScreenshotTools.SW_MINIMIZE);
                    }
                };

                if (include_borders && !string.IsNullOrEmpty(exactTitle))
                {
                    return await RecordWithFFmpegAsync($"title=\"{exactTitle}\"", "", duration_s, fps, save_directory, filename, format, audio_device_id_or_name, onStart, onEnd, progress);
                }
                else
                {
                    onStart(); 
                    
                    if (!ScreenshotTools.GetWindowRect(hwnd, out ScreenshotTools.RECT rect))
                    {
                        return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = "Could not get window rect." } } };
                    }

                    int left = Math.Max(rect.Left, SystemInformation.VirtualScreen.Left);
                    int top = Math.Max(rect.Top, SystemInformation.VirtualScreen.Top);
                    int right = Math.Min(rect.Right, SystemInformation.VirtualScreen.Right);
                    int bottom = Math.Min(rect.Bottom, SystemInformation.VirtualScreen.Bottom);

                    int width = right - left;
                    int height = bottom - top;
                    if (width <= 0 || height <= 0)
                    {
                        return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = "Invalid window dimensions (possibly fully off-screen)." } } };
                    }

                    width = width - (width % 2);
                    height = height - (height % 2);

                    return await RecordWithFFmpegAsync("desktop", $"-offset_x {left} -offset_y {top} -video_size {width}x{height}", duration_s, fps, save_directory, filename, format, audio_device_id_or_name, null, onEnd, progress);
                }
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new System.Collections.Generic.List<ContentBlock> { new TextContentBlock { Text = ex.ToString() } } };
            }
        }
    }
}
