using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.IO;
using System.Threading.Tasks;
using ModelContextProtocol;
using ModelContextProtocol.Protocol;
using ModelContextProtocol.Server;

namespace controller_mcp.Features.Tools
{
    public static class ImageTools
    {
        private static CallToolResult EnsureFileExists(string path, string paramName)
        {
            if (!File.Exists(path))
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"The specified {paramName} was not found: {path}" } } };
            }
            return null;
        }

        private static void EnsureDirectoryExists(string filePath)
        {
            string dir = Path.GetDirectoryName(filePath);
            if (!string.IsNullOrEmpty(dir) && !Directory.Exists(dir))
            {
                Directory.CreateDirectory(dir);
            }
        }

        [McpServerTool, Description("Converts an image from one format to another (e.g., .png to .jpg, .webp, .gif). FFmpeg automatically handles the encoding based on the output extension.")]
        public static async Task<CallToolResult> ConvertImageFormat(string input_file, string output_file, IProgress<ProgressNotificationValue> progress = null)
        {
            string safeInput = InputValidator.ValidateFilePath(input_file, nameof(input_file));
            string safeOutput = InputValidator.ValidateFilePath(output_file, nameof(output_file));

            var check = EnsureFileExists(safeInput, nameof(input_file));
            if (check != null) return check;
            EnsureDirectoryExists(safeOutput);

            if (progress != null)
                progress.Report(new ProgressNotificationValue { Progress = 10, Message = "Starting image conversion..." });

            string args = $"-y -i \"{safeInput}\" \"{safeOutput}\"";
            await FFmpegHelper.RunFFmpegAsync(args);

            if (progress != null)
                progress.Report(new ProgressNotificationValue { Progress = 100, Message = "Complete" });

            return new CallToolResult
            {
                Content = new List<ContentBlock> { new TextContentBlock { Text = $"Successfully converted image format. Saved to: {output_file}" } }
            };
        }
    }
}
