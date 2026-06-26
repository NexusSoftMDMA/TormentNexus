using System;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.ComponentModel;
using System.Diagnostics;
using System.Drawing;
using System.IO;
using System.Runtime.InteropServices;
using System.Text;
using System.Threading.Tasks;
using System.Windows.Forms;
using CSCore.CoreAudioAPI;
using ModelContextProtocol;
using ModelContextProtocol.Protocol;
using ModelContextProtocol.Server;

namespace controller_mcp.Features.Tools
{
    public class ActiveRecording
    {
        public string Id { get; set; }
        public Process Process { get; set; }
        public DateTime StartTime { get; set; }
        public string OutputPath { get; set; }
        public string TargetName { get; set; }

        public int IsolateAudioPid { get; set; } = -1;
        public List<Tuple<SimpleAudioVolume, bool>> MutedSessions { get; set; } = new List<Tuple<SimpleAudioVolume, bool>>();
        public AudioSessionManager2 SessionManager { get; set; }
        public EventHandler<SessionCreatedEventArgs> SessionCreatedHandler { get; set; }

        public void RestoreAudio()
        {
            try
            {
                if (SessionManager != null && SessionCreatedHandler != null)
                {
                    SessionManager.SessionCreated -= SessionCreatedHandler;
                }
                foreach (var state in MutedSessions)
                {
                    try
                    {
                        if (state.Item1 != null)
                            state.Item1.IsMuted = state.Item2;
                    }
                    catch { }
                }
                MutedSessions.Clear();
            }
            catch { }
        }
    }

    public static class StatefulRecordingTools
    {
        private static readonly ConcurrentDictionary<string, ActiveRecording> _activeRecordings = new ConcurrentDictionary<string, ActiveRecording>();

        public static void StopAll()
        {
            foreach (var kvp in _activeRecordings)
            {
                try
                {
                    kvp.Value.RestoreAudio();
                    if (kvp.Value.Process != null && !kvp.Value.Process.HasExited)
                    {
                        kvp.Value.Process.Kill();
                        kvp.Value.Process.Dispose();
                    }
                }
                catch { }
            }
            _activeRecordings.Clear();
        }

        public static void ResumeRecording(RecordingBackup backup)
        {
            try
            {
                var proc = Process.GetProcessById(backup.ProcessId);
                if (!proc.HasExited)
                {
                    var rec = new ActiveRecording
                    {
                        Id = backup.Id,
                        Process = proc,
                        StartTime = backup.StartTime,
                        OutputPath = backup.OutputPath,
                        TargetName = backup.TargetName,
                        IsolateAudioPid = backup.IsolateAudioPid
                    };
                    
                    if (backup.IsolateAudioPid > 0)
                    {
                        EnforceIsolation(rec);
                    }

                    _activeRecordings.TryAdd(backup.Id, rec);
                    AuditLogger.Log(LogLevel.INFO, "StatefulRecording", $"Resumed active recording {backup.Id} (PID: {backup.ProcessId}) from state backup.");
                }
                else
                {
                    StateBackupManager.RemoveRecording(backup.Id);
                }
            }
            catch
            {
                StateBackupManager.RemoveRecording(backup.Id);
            }
        }

        static StatefulRecordingTools()
        {
            AppDomain.CurrentDomain.ProcessExit += (s, e) =>
            {
                foreach (var recording in _activeRecordings.Values)
                {
                    recording.RestoreAudio();
                }
            };
        }

        [DllImport("user32.dll")]
        [return: MarshalAs(UnmanagedType.Bool)]
        private static extern bool GetWindowRect(IntPtr hWnd, out RECT lpRect);

        [DllImport("user32.dll")]
        [return: MarshalAs(UnmanagedType.Bool)]
        private static extern bool IsWindowVisible(IntPtr hWnd);

        [StructLayout(LayoutKind.Sequential)]
        private struct RECT
        {
            public int Left;
            public int Top;
            public int Right;
            public int Bottom;
        }

        private static void EnforceIsolation(ActiveRecording recording)
        {
            try
            {
                var enumerator = new MMDeviceEnumerator();
                var device = enumerator.GetDefaultAudioEndpoint(DataFlow.Render, CSCore.CoreAudioAPI.Role.Multimedia);
                var sessionManager = AudioSessionManager2.FromMMDevice(device);
                recording.SessionManager = sessionManager;

                Action<AudioSessionControl> handleSession = (session) =>
                {
                    try
                    {
                        using (var sessionControl2 = session.QueryInterface<AudioSessionControl2>())
                        {
                            int pid = sessionControl2.ProcessID;
                            if (pid != recording.IsolateAudioPid && pid != Process.GetCurrentProcess().Id && pid != 0)
                            {
                                var volume = session.QueryInterface<SimpleAudioVolume>();
                                bool wasMuted = volume.IsMuted;
                                if (!wasMuted)
                                {
                                    volume.IsMuted = true;
                                    lock (recording.MutedSessions)
                                    {
                                        recording.MutedSessions.Add(new Tuple<SimpleAudioVolume, bool>(volume, wasMuted));
                                    }
                                }
                            }
                        }
                    }
                    catch { }
                };

                using (var sessionEnumerator = sessionManager.GetSessionEnumerator())
                {
                    foreach (var session in sessionEnumerator)
                    {
                        handleSession(session);
                    }
                }

                recording.SessionCreatedHandler = (s, e) => handleSession(e.NewSession);
                sessionManager.SessionCreated += recording.SessionCreatedHandler;
            }
            catch { }
        }

        [McpServerTool, Description("Starts a long-running background video recording. Returns a recording_id immediately. target_type must be 'window', 'monitor', or 'all_screens'. Can optionally completely mute audio, or isolate audio to a specific PID.")]
        public static async Task<CallToolResult> StartRecording(
            string target_type,
            string target_name_or_index,
            string save_directory,
            string filename,
            int fps = 24,
            bool mute_audio = false,
            int isolate_audio_pid = -1,
            string audio_device_id_or_name = null)
        {
            try
            {
                if (string.IsNullOrEmpty(save_directory) || string.IsNullOrEmpty(filename))
                {
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = "save_directory and filename are required." } } };
                }
                if (!Directory.Exists(save_directory))
                    Directory.CreateDirectory(save_directory);

                string finalName = string.IsNullOrWhiteSpace(filename) 
                    ? $"recording_{DateTime.Now:yyyyMMdd_HHmmss}.mp4" 
                    : filename;
                    
                if (!finalName.EndsWith(".mp4", StringComparison.OrdinalIgnoreCase))
                    finalName += ".mp4";

                string outputPath = Path.Combine(InputValidator.ValidateFilePath(save_directory, nameof(save_directory)), 
                                        InputValidator.ValidateFilePath(finalName, nameof(filename)));

                string gdigrabInput = "";
                string additionalArgs = "";

                if (target_type.Equals("all_screens", StringComparison.OrdinalIgnoreCase))
                {
                    Rectangle totalSize = Rectangle.Empty;
                    foreach (Screen screen in Screen.AllScreens)
                        totalSize = Rectangle.Union(totalSize, screen.Bounds);

                    int width = totalSize.Width - (totalSize.Width % 2);
                    int height = totalSize.Height - (totalSize.Height % 2);
                    gdigrabInput = "desktop";
                    additionalArgs = $"-offset_x {SystemInformation.VirtualScreen.Left} -offset_y {SystemInformation.VirtualScreen.Top} -video_size {width}x{height}";
                }
                else if (target_type.Equals("monitor", StringComparison.OrdinalIgnoreCase))
                {
                    if (!int.TryParse(target_name_or_index, out int screenIndex))
                        screenIndex = 0;

                    var screens = Screen.AllScreens;
                    if (screenIndex < 0 || screenIndex >= screens.Length)
                        screenIndex = 0;

                    var bounds = screens[screenIndex].Bounds;
                    int width = bounds.Width - (bounds.Width % 2);
                    int height = bounds.Height - (bounds.Height % 2);
                    gdigrabInput = "desktop";
                    additionalArgs = $"-offset_x {bounds.Left} -offset_y {bounds.Top} -video_size {width}x{height}";
                }
                else if (target_type.Equals("window", StringComparison.OrdinalIgnoreCase))
                {
                    Process targetProcess = null;
                    ScreenshotTools.EnumDelegate filter = delegate (IntPtr hWnd, int lParam)
                    {
                        if (!IsWindowVisible(hWnd)) return true;
                        StringBuilder sb = new StringBuilder(255);
                        ScreenshotTools.GetWindowText(hWnd, sb, sb.Capacity + 1);
                        string title = sb.ToString();

                        ScreenshotTools.GetWindowThreadProcessId(hWnd, out uint pid);
                        try
                        {
                            var proc = Process.GetProcessById((int)pid);
                            if (title.IndexOf(target_name_or_index, StringComparison.OrdinalIgnoreCase) >= 0 ||
                                proc.ProcessName.IndexOf(target_name_or_index, StringComparison.OrdinalIgnoreCase) >= 0)
                            {
                                targetProcess = proc;
                                return false;
                            }
                        }
                        catch { }
                        return true;
                    };

                    ScreenshotTools.EnumWindows(filter, 0);

                    if (targetProcess == null)
                    {
                        return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Could not find an open window matching '{target_name_or_index}'." } } };
                    }
                    if (!GetWindowRect(targetProcess.MainWindowHandle, out RECT rect))
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

                    gdigrabInput = "desktop";
                    additionalArgs = $"-offset_x {left} -offset_y {top} -video_size {width}x{height}";
                }
                else
                {
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = "target_type must be 'window', 'monitor', or 'all_screens'" } } };
                }

                string audioArgs = "";
                string audioCodecArgs = "";
                if (!mute_audio && !string.IsNullOrEmpty(audio_device_id_or_name))
                {
                    string exactAudio = await AudioTools.ResolveAudioDeviceNameAsync(audio_device_id_or_name);
                    string safeAudio = InputValidator.EscapeArgument(exactAudio);
                    audioArgs = $"-f dshow -i audio={safeAudio}";
                    audioCodecArgs = "-c:a aac -b:a 128k";
                }

                string safeOutputPath = InputValidator.EscapeArgument(outputPath);
                string mp4Args = $"-y -f gdigrab -framerate {fps} {additionalArgs} -i {gdigrabInput} {audioArgs} -c:v libx264 -preset ultrafast -crf 18 {audioCodecArgs} {safeOutputPath}";

                string ffmpegPath = FFmpegHelper.GetFFmpegPath();
                var psi = new ProcessStartInfo
                {
                    FileName = ffmpegPath,
                    Arguments = mp4Args,
                    UseShellExecute = false,
                    CreateNoWindow = true,
                    RedirectStandardInput = true,
                    RedirectStandardOutput = true,
                    RedirectStandardError = true
                };

                Process process = new Process { StartInfo = psi, EnableRaisingEvents = true };
                
                string id = Guid.NewGuid().ToString();

                process.OutputDataReceived += (s, e) => { };
                process.ErrorDataReceived += (s, e) => { };

                process.Start();
                process.BeginOutputReadLine();
                process.BeginErrorReadLine();

                var recording = new ActiveRecording
                {
                    Id = id,
                    Process = process,
                    StartTime = DateTime.Now,
                    OutputPath = outputPath,
                    TargetName = target_name_or_index,
                    IsolateAudioPid = isolate_audio_pid
                };

                if (isolate_audio_pid > 0 && !mute_audio)
                {
                    EnforceIsolation(recording);
                }

                _activeRecordings.TryAdd(id, recording);
                StateBackupManager.AddRecording(id, process.Id, recording.StartTime, outputPath, target_name_or_index, isolate_audio_pid);

                return new CallToolResult
                {
                    Content = new List<ContentBlock> { 
                        new TextContentBlock { Text = $"{{\"status\":\"started\", \"recording_id\":\"{id}\", \"output_path\":\"{outputPath.Replace("\\", "\\\\")}\"}}" } 
                    }
                };
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = ex.ToString() } } };
            }
        }

        [McpServerTool, Description("Checks the status of an active long-running recording session.")]
        public static CallToolResult CheckRecordingStatus(string recording_id)
        {
            if (_activeRecordings.TryGetValue(recording_id, out ActiveRecording recording))
            {
                bool isRunning = !recording.Process.HasExited;
                TimeSpan elapsed = DateTime.Now - recording.StartTime;
                
                return new CallToolResult
                {
                    Content = new List<ContentBlock> { 
                        new TextContentBlock { Text = $"{{\"status\":\"{(isRunning ? "running" : "exited")}\", \"recording_id\":\"{recording_id}\", \"elapsed_seconds\":{elapsed.TotalSeconds:F1}, \"output_path\":\"{recording.OutputPath.Replace("\\", "\\\\")}\"}}" } 
                    }
                };
            }
            else
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"No active recording found with ID '{recording_id}'." } } };
            }
        }

        [McpServerTool, Description("Gracefully stops a long-running recording session and finalizes the MP4 file. Also restores any muted audio states if isolated audio was used.")]
        public static async Task<CallToolResult> StopRecording(string recording_id)
        {
            if (_activeRecordings.TryGetValue(recording_id, out ActiveRecording recording))
            {
                recording.RestoreAudio();

                if (!recording.Process.HasExited)
                {
                    try
                    {
                        recording.Process.StandardInput.WriteLine("q");
                        await Task.Run(() => recording.Process.WaitForExit(10000));

                        if (!recording.Process.HasExited)
                        {
                            recording.Process.Kill();
                        }
                    }
                    catch { }
                }

                _activeRecordings.TryRemove(recording_id, out _);
                StateBackupManager.RemoveRecording(recording_id);

                return new CallToolResult
                {
                    Content = new List<ContentBlock> { 
                        new TextContentBlock { Text = $"Recording stopped successfully. File finalized at: {recording.OutputPath}" } 
                    }
                };
            }
            else
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"No active recording found with ID '{recording_id}'." } } };
            }
        }
    }
}
