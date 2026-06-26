using System;
using System.Diagnostics;
using System.IO;
using System.Net.Http;
using System.Text.RegularExpressions;
using System.Threading.Tasks;
using System.Windows.Forms;

namespace controller_mcp.Features.Tools
{
    public static class UpdateManager
    {
        private const string GITHUB_API_URL = "https://api.github.com/repos/blazium-games/YADMS/releases/latest";
        private static readonly HttpClient _httpClient;

        static UpdateManager()
        {
            _httpClient = new HttpClient();
            _httpClient.Timeout = TimeSpan.FromSeconds(15);
            _httpClient.DefaultRequestHeaders.Add("User-Agent", "YADMS-AutoUpdater");
        }

        public static string GetCurrentSku()
        {
            bool ffmpeg = false;
            bool hack = false;
#if BUILD_WITH_FFMPEG
            ffmpeg = true;
#endif
#if GAME_HACKING
            hack = true;
#endif
            if (ffmpeg && hack) return "Bundled_Hacked";
            if (ffmpeg && !hack) return "Bundled_Standard";
            if (!ffmpeg && hack) return "Lite_Hacked";
            return "Lite_Standard";
        }

        public static (bool UpdateAvailable, string LatestVersion, string DownloadUrl) ParseReleaseResponse(string json, string currentVersion, string currentSku)
        {
            // Parse tag name (e.g. "v1.0.142" or "1.0.142")
            var tagMatch = Regex.Match(json, @"""tag_name"":\s*""v?([^""]+)""");
            if (!tagMatch.Success) return (false, null, null);

            string latestVersionStr = tagMatch.Groups[1].Value;
            
            if (!Version.TryParse(latestVersionStr, out Version latestVer)) return (false, null, null);
            if (!Version.TryParse(currentVersion, out Version currentVer)) return (false, null, null);

            if (latestVer <= currentVer)
            {
                return (false, latestVersionStr, null);
            }

            // We expect an asset like YADMS_Installer_Lite_Standard_v1.0.142.exe
            string expectedPattern = $@"""browser_download_url"":\s*""([^""]+YADMS_Installer_{currentSku}_v{latestVersionStr}\.exe)""";
            var urlMatch = Regex.Match(json, expectedPattern, RegexOptions.IgnoreCase);

            if (urlMatch.Success)
            {
                return (true, latestVersionStr, urlMatch.Groups[1].Value);
            }

            // Fallback: If for some reason the version has no 'v' in the filename, try a looser match
            string loosePattern = $@"""browser_download_url"":\s*""([^""]+{currentSku}[^""]+\.exe)""";
            var looseMatch = Regex.Match(json, loosePattern, RegexOptions.IgnoreCase);
            
            if (looseMatch.Success)
            {
                return (true, latestVersionStr, looseMatch.Groups[1].Value);
            }

            return (true, latestVersionStr, null); // Update available, but no matching SKU asset found
        }

        public static async Task<(bool UpdateAvailable, string LatestVersion, string DownloadUrl, string ErrorReason)> CheckForUpdatesAsync()
        {
            try
            {
                string json;
                try
                {
                    json = await _httpClient.GetStringAsync(GITHUB_API_URL);
                }
                catch (HttpRequestException httpEx)
                {
                    if (httpEx.Message.Contains("404"))
                        return (false, null, null, "No Releases Found (404)");
                    if (httpEx.Message.Contains("403"))
                        return (false, null, null, "API Rate Limited (403)");
                    
                    return (false, null, null, $"HTTP Error: {httpEx.Message}");
                }

                var result = ParseReleaseResponse(json, VersionInfo.CurrentVersion, GetCurrentSku());
                return (result.UpdateAvailable, result.LatestVersion, result.DownloadUrl, null);
            }
            catch (Exception ex)
            {
                AuditLogger.Log(LogLevel.ERROR, "System", $"Failed to check for updates: {ex.Message}");
                return (false, null, null, ex.Message);
            }
        }

        public static async Task<bool> DownloadAndInstallUpdateAsync(string downloadUrl, string latestVersion)
        {
            string tempDir = Path.GetTempPath();
            string fileName = $"YADMS_Update_v{latestVersion}.exe";
            string tempFilePath = Path.Combine(tempDir, fileName);

            try
            {
                var request = new HttpRequestMessage(HttpMethod.Get, downloadUrl);
                long existingLength = 0;
                
                if (File.Exists(tempFilePath))
                {
                    existingLength = new FileInfo(tempFilePath).Length;
                    request.Headers.Range = new System.Net.Http.Headers.RangeHeaderValue(existingLength, null);
                    AuditLogger.Log(LogLevel.INFO, "Updater", $"Resuming partial download from {existingLength} bytes.");
                }

                using (var response = await _httpClient.SendAsync(request, HttpCompletionOption.ResponseHeadersRead))
                {
                    if (response.StatusCode == System.Net.HttpStatusCode.PartialContent)
                    {
                        // Server supports resuming, append to file
                        using (var fs = new FileStream(tempFilePath, FileMode.Append, FileAccess.Write, FileShare.None))
                        {
                            await response.Content.CopyToAsync(fs);
                        }
                    }
                    else
                    {
                        // Server didn't support range or file didn't exist, download from scratch
                        response.EnsureSuccessStatusCode();
                        using (var fs = new FileStream(tempFilePath, FileMode.Create, FileAccess.Write, FileShare.None))
                        {
                            await response.Content.CopyToAsync(fs);
                        }
                    }
                }

                // Instead of shutting down here, we clone the updater to Temp and launch it.
                // The updater will IPC shutdown us, wait for exit, and launch installer.
                string updaterSource = Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "YADMS_Updater.exe");
                if (!File.Exists(updaterSource))
                {
                    AuditLogger.Log(LogLevel.ERROR, "Updater", "YADMS_Updater.exe not found alongside daemon.");
                    if (File.Exists(tempFilePath)) File.Delete(tempFilePath);
                    return false;
                }

                string updaterTempPath = Path.Combine(tempDir, $"updater_{Guid.NewGuid():N}.exe");
                File.Copy(updaterSource, updaterTempPath, true);

                ProcessStartInfo psi = new ProcessStartInfo
                {
                    FileName = updaterTempPath,
                    Arguments = $"/install \"{tempFilePath}\" /pid {Process.GetCurrentProcess().Id}",
                    CreateNoWindow = true,
                    UseShellExecute = false
                };
                
                Process.Start(psi);
                
                // We do NOT exit. We wait for the IPC EXIT command from updater.
                return true;
            }
            catch (Exception ex)
            {
                AuditLogger.Log(LogLevel.ERROR, "System", $"Failed to download or install update: {ex.Message}");
                // We purposefully do NOT delete the temp file here to preserve the partial download for resuming.
                return false;
            }
        }
    }
}
