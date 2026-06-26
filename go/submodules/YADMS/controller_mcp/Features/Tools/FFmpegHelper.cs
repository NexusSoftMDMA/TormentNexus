using System;
using System.Collections.Concurrent;
using System.Diagnostics;
using System.IO;
using System.Reflection;

namespace controller_mcp.Features.Tools
{
    public static class FFmpegHelper
    {
        private static string _ffmpegPath;
        private static readonly ConcurrentDictionary<int, Process> _activeProcesses = new ConcurrentDictionary<int, Process>();

        static FFmpegHelper()
        {
            AppDomain.CurrentDomain.ProcessExit += (s, e) => KillAll();
        }

        public static void KillAll()
        {
            foreach (var kvp in _activeProcesses)
            {
                try { if (!kvp.Value.HasExited) kvp.Value.Kill(); } catch { }
            }
            _activeProcesses.Clear();
        }

        public static string GetFFmpegPath()
        {
            if (!string.IsNullOrEmpty(_ffmpegPath) && File.Exists(_ffmpegPath))
                return _ffmpegPath;

#if BUILD_WITH_FFMPEG
            _ffmpegPath = Path.Combine(Path.GetTempPath(), "controller_mcp_ffmpeg.exe");

            if (!File.Exists(_ffmpegPath))
            {
                string resourceName = "controller_mcp.ffmpeg.exe";
                var assembly = Assembly.GetExecutingAssembly();
                
                using (Stream resStream = assembly.GetManifestResourceStream(resourceName))
                {
                    if (resStream == null)
                    {
                        throw new FileNotFoundException($"Embedded resource {resourceName} not found.");
                    }

                    using (FileStream fs = new FileStream(_ffmpegPath, FileMode.Create, FileAccess.Write))
                    {
                        resStream.CopyTo(fs);
                    }
                }
            }

            return _ffmpegPath;
#else
            var settings = AppSettings.Load();
            if (!string.IsNullOrEmpty(settings.FFmpegPath) && File.Exists(settings.FFmpegPath))
            {
                _ffmpegPath = settings.FFmpegPath;
                return _ffmpegPath;
            }
            throw new FileNotFoundException("FFmpeg is not bundled with this build. Please configure the FFmpeg Path in the UI Settings.");
#endif
        }

        public static async System.Threading.Tasks.Task RunFFmpegAsync(string arguments)
        {
            string ffmpegPath = GetFFmpegPath();
            var psi = new ProcessStartInfo
            {
                FileName = ffmpegPath,
                Arguments = arguments,
                UseShellExecute = false,
                CreateNoWindow = true,
                RedirectStandardOutput = true,
                RedirectStandardError = true
            };

            using (var process = new Process { StartInfo = psi, EnableRaisingEvents = true })
            {
                var tcs = new System.Threading.Tasks.TaskCompletionSource<bool>();
                process.Exited += (sender, args) => tcs.TrySetResult(true);

                process.Start();
                _activeProcesses.TryAdd(process.Id, process);

                var errorTask = process.StandardError.ReadToEndAsync();
                var outputTask = process.StandardOutput.ReadToEndAsync();
                
                await tcs.Task;
                _activeProcesses.TryRemove(process.Id, out _);
                string error = await errorTask;
                string output = await outputTask;

                if (process.ExitCode != 0)
                {
                    throw new InvalidOperationException($"FFmpeg failed with exit code {process.ExitCode}\nError: {error}\nArgs: {arguments}");
                }
            }
        }
    }
}
