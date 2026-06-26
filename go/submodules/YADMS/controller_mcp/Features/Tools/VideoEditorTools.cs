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
    public static class VideoEditorTools
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

        [McpServerTool, Description("Converts a video from one format/container to another (e.g., .avi to .mp4). FFmpeg will automatically handle codec translation based on the output extension.")]
        public static async Task<CallToolResult> ConvertVideoFormat(string input_file, string output_file, IProgress<ProgressNotificationValue> progress = null)
        {
            string safeInputPath = InputValidator.ValidateFilePath(input_file, nameof(input_file));
            string safeOutputPath = InputValidator.ValidateFilePath(output_file, nameof(output_file));

            var check = EnsureFileExists(safeInputPath, nameof(input_file));
            if (check != null) return check;
            EnsureDirectoryExists(safeOutputPath);

            if (progress != null)
                progress.Report(new ProgressNotificationValue { Progress = 10, Message = "Starting format conversion..." });

            string safeInput = InputValidator.EscapeArgument(safeInputPath);
            string safeOutput = InputValidator.EscapeArgument(safeOutputPath);

            string args = $"-y -i {safeInput} {safeOutput}";
            await FFmpegHelper.RunFFmpegAsync(args);

            if (progress != null)
                progress.Report(new ProgressNotificationValue { Progress = 100, Message = "Complete" });

            return new CallToolResult
            {
                Content = new List<ContentBlock> { new TextContentBlock { Text = $"Successfully converted video format. Saved to: {safeOutputPath}" } }
            };
        }

        [McpServerTool, Description("Strips the audio track from a video and saves it as an independent audio file (e.g., .mp3 or .wav).")]
        public static async Task<CallToolResult> ExtractAudioFromVideo(string input_file, string output_audio_file, IProgress<ProgressNotificationValue> progress = null)
        {
            string safeInputPath = InputValidator.ValidateFilePath(input_file, nameof(input_file));
            string safeOutputPath = InputValidator.ValidateFilePath(output_audio_file, nameof(output_audio_file));

            var check = EnsureFileExists(safeInputPath, nameof(input_file));
            if (check != null) return check;
            EnsureDirectoryExists(safeOutputPath);

            if (progress != null)
                progress.Report(new ProgressNotificationValue { Progress = 10, Message = "Extracting audio..." });

            string safeInput = InputValidator.EscapeArgument(safeInputPath);
            string safeOutput = InputValidator.EscapeArgument(safeOutputPath);

            // -vn disables video stream processing
            string args = $"-y -i {safeInput} -vn {safeOutput}";
            await FFmpegHelper.RunFFmpegAsync(args);

            if (progress != null)
                progress.Report(new ProgressNotificationValue { Progress = 100, Message = "Complete" });

            return new CallToolResult
            {
                Content = new List<ContentBlock> { new TextContentBlock { Text = $"Successfully extracted audio. Saved to: {safeOutputPath}" } }
            };
        }

        [McpServerTool, Description("Creates a silenced version of a video by completely removing all audio tracks. Video quality is losslessly copied.")]
        public static async Task<CallToolResult> RemoveAudioFromVideo(string input_file, string output_file, IProgress<ProgressNotificationValue> progress = null)
        {
            string safeInputPath = InputValidator.ValidateFilePath(input_file, nameof(input_file));
            string safeOutputPath = InputValidator.ValidateFilePath(output_file, nameof(output_file));

            var check = EnsureFileExists(safeInputPath, nameof(input_file));
            if (check != null) return check;
            EnsureDirectoryExists(safeOutputPath);

            if (progress != null)
                progress.Report(new ProgressNotificationValue { Progress = 10, Message = "Removing audio tracks..." });

            string safeInput = InputValidator.EscapeArgument(safeInputPath);
            string safeOutput = InputValidator.EscapeArgument(safeOutputPath);

            // -c:v copy copies video stream without re-encoding
            // -an disables audio streams
            string args = $"-y -i {safeInput} -c:v copy -an {safeOutput}";
            await FFmpegHelper.RunFFmpegAsync(args);

            if (progress != null)
                progress.Report(new ProgressNotificationValue { Progress = 100, Message = "Complete" });

            return new CallToolResult
            {
                Content = new List<ContentBlock> { new TextContentBlock { Text = $"Successfully removed audio. Silenced video saved to: {safeOutputPath}" } }
            };
        }

        [McpServerTool, Description("Replaces the original audio of a video with a new audio file. Resulting video will stop when the shortest of the two streams ends.")]
        public static async Task<CallToolResult> ReplaceAudioInVideo(string input_video_file, string input_audio_file, string output_file, IProgress<ProgressNotificationValue> progress = null)
        {
            string safeVideoPath = InputValidator.ValidateFilePath(input_video_file, nameof(input_video_file));
            string safeAudioPath = InputValidator.ValidateFilePath(input_audio_file, nameof(input_audio_file));
            string safeOutputPath = InputValidator.ValidateFilePath(output_file, nameof(output_file));

            var check1 = EnsureFileExists(safeVideoPath, nameof(input_video_file));
            if (check1 != null) return check1;
            var check2 = EnsureFileExists(safeAudioPath, nameof(input_audio_file));
            if (check2 != null) return check2;
            EnsureDirectoryExists(safeOutputPath);

            if (progress != null)
                progress.Report(new ProgressNotificationValue { Progress = 10, Message = "Multiplexing new audio into video..." });

            string safeVideo = InputValidator.EscapeArgument(safeVideoPath);
            string safeAudio = InputValidator.EscapeArgument(safeAudioPath);
            string safeOutput = InputValidator.EscapeArgument(safeOutputPath);

            // -map 0:v:0 grabs first video stream from first input
            // -map 1:a:0 grabs first audio stream from second input
            // -shortest trims output to length of shortest stream
            string args = $"-y -i {safeVideo} -i {safeAudio} -c:v copy -c:a aac -map 0:v:0 -map 1:a:0 -shortest {safeOutput}";
            await FFmpegHelper.RunFFmpegAsync(args);

            if (progress != null)
                progress.Report(new ProgressNotificationValue { Progress = 100, Message = "Complete" });

            return new CallToolResult
            {
                Content = new List<ContentBlock> { new TextContentBlock { Text = $"Successfully replaced audio. Merged video saved to: {safeOutputPath}" } }
            };
        }
    }
}
