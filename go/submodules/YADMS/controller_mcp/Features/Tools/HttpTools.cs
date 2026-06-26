using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.IO;
using System.Net.Http;
using System.Text;
using System.Threading.Tasks;
using ModelContextProtocol;
using ModelContextProtocol.Protocol;
using ModelContextProtocol.Server;

namespace controller_mcp.Features.Tools
{
    public static class HttpTools
    {
        private static readonly HttpClient _httpClient = new HttpClient();

        [McpServerTool, Description("Sends an HTTP request (like curl) to a specified URL and returns the response body and status code. Useful for testing localhost APIs.")]
        public static async Task<CallToolResult> MakeHttpRequest(string method, string url, string headers_json = "{}", string body = "")
        {
            try
            {
                string safeUrl = InputValidator.ValidateUrl(url, nameof(url));
                var request = new HttpRequestMessage(new HttpMethod(method.ToUpper()), safeUrl);

                // Parse headers if any (basic implementation)
                if (!string.IsNullOrWhiteSpace(headers_json) && headers_json != "{}")
                {
                    try
                    {
                        var headers = System.Text.Json.JsonSerializer.Deserialize<Dictionary<string, string>>(headers_json);
                        if (headers != null)
                        {
                            foreach (var kvp in headers)
                            {
                                request.Headers.TryAddWithoutValidation(kvp.Key, kvp.Value);
                            }
                        }
                    }
                    catch { /* ignore header parse failure */ }
                }

                if (!string.IsNullOrEmpty(body) && (method.Equals("POST", StringComparison.OrdinalIgnoreCase) || method.Equals("PUT", StringComparison.OrdinalIgnoreCase) || method.Equals("PATCH", StringComparison.OrdinalIgnoreCase)))
                {
                    request.Content = new StringContent(body, Encoding.UTF8, "application/json");
                }

                var response = await _httpClient.SendAsync(request);
                string responseBody = await response.Content.ReadAsStringAsync();

                return new CallToolResult
                {
                    Content = new List<ContentBlock> 
                    { 
                        new TextContentBlock { Text = $"Status: {(int)response.StatusCode} {response.ReasonPhrase}\n\n{responseBody}" } 
                    }
                };
            }
            catch (Exception ex)
            {
                return new CallToolResult
                {
                    IsError = true,
                    Content = new List<ContentBlock> { new TextContentBlock { Text = $"HTTP Request Failed: {ex.Message}" } }
                };
            }
        }

        [McpServerTool, Description("Downloads a file from a URL and saves it to the local disk.")]
        public static async Task<CallToolResult> DownloadFile(string url, string dest_path, IProgress<ProgressNotificationValue> progress = null)
        {
            try
            {
                string safeUrl = InputValidator.ValidateUrl(url, nameof(url));
                string safeDest = InputValidator.ValidateFilePath(dest_path, nameof(dest_path));

                string dir = Path.GetDirectoryName(safeDest);
                if (!string.IsNullOrEmpty(dir) && !Directory.Exists(dir))
                {
                    Directory.CreateDirectory(dir);
                }

                if (progress != null)
                    progress.Report(new ProgressNotificationValue { Progress = 10, Message = "Connecting..." });

                using (var response = await _httpClient.GetAsync(safeUrl, HttpCompletionOption.ResponseHeadersRead))
                {
                    response.EnsureSuccessStatusCode();

                    using (var stream = await response.Content.ReadAsStreamAsync())
                    using (var fileStream = new FileStream(safeDest, FileMode.Create, FileAccess.Write, FileShare.None, 8192, true))
                    {
                        if (progress != null)
                            progress.Report(new ProgressNotificationValue { Progress = 50, Message = "Downloading..." });

                        await stream.CopyToAsync(fileStream);
                    }
                }

                if (progress != null)
                    progress.Report(new ProgressNotificationValue { Progress = 100, Message = "Complete" });

                return new CallToolResult
                {
                    Content = new List<ContentBlock> { new TextContentBlock { Text = $"Successfully downloaded file to: {safeDest}" } }
                };
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Failed to download file: {ex.Message}" } } };
            }
        }
    }
}
