using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.IO;
using System.Text;
using ModelContextProtocol.Protocol;
using ModelContextProtocol.Server;

namespace controller_mcp.Features.Tools
{
    public static class Base64Tools
    {
        [McpServerTool, Description("Encodes raw UTF-8 text into a Base64 string.")]
        public static CallToolResult EncodeTextBase64(string text)
        {
            try
            {
                if (string.IsNullOrEmpty(text))
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = "Input text cannot be empty." } } };

                string base64 = Convert.ToBase64String(Encoding.UTF8.GetBytes(text));
                return new CallToolResult
                {
                    Content = new List<ContentBlock> { new TextContentBlock { Text = base64 } }
                };
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Encode failed: {ex.Message}" } } };
            }
        }

        [McpServerTool, Description("Decodes a Base64 string back into raw UTF-8 text.")]
        public static CallToolResult DecodeTextBase64(string base64)
        {
            try
            {
                if (string.IsNullOrEmpty(base64))
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = "Input base64 cannot be empty." } } };

                byte[] bytes = Convert.FromBase64String(base64);
                string decoded = Encoding.UTF8.GetString(bytes);
                return new CallToolResult
                {
                    Content = new List<ContentBlock> { new TextContentBlock { Text = decoded } }
                };
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Decode failed: {ex.Message}" } } };
            }
        }

        [McpServerTool, Description("Reads a file from disk and encodes its raw bytes into a Base64 string.")]
        public static CallToolResult EncodeFileBase64(string filepath)
        {
            try
            {
                string safePath = InputValidator.ValidateFilePath(filepath, nameof(filepath));
                
                if (!File.Exists(safePath))
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"File not found: {safePath}" } } };

                byte[] bytes = File.ReadAllBytes(safePath);
                string base64 = Convert.ToBase64String(bytes);

                return new CallToolResult
                {
                    Content = new List<ContentBlock> { new TextContentBlock { Text = base64 } }
                };
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"File read failed: {ex.Message}" } } };
            }
        }

        [McpServerTool, Description("Decodes a Base64 string into raw binary and saves it to a file on disk.")]
        public static CallToolResult DecodeFileBase64(string base64, string save_path)
        {
            try
            {
                if (string.IsNullOrEmpty(base64) || string.IsNullOrEmpty(save_path))
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = "Both base64 and save_path are required." } } };

                string safePath = InputValidator.ValidateFilePath(save_path, nameof(save_path));

                byte[] bytes = Convert.FromBase64String(base64);
                
                string directory = Path.GetDirectoryName(safePath);
                if (!string.IsNullOrEmpty(directory) && !Directory.Exists(directory))
                {
                    Directory.CreateDirectory(directory);
                }

                File.WriteAllBytes(safePath, bytes);

                return new CallToolResult
                {
                    Content = new List<ContentBlock> { new TextContentBlock { Text = $"Successfully decoded and saved to: {safePath}" } }
                };
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"File write failed: {ex.Message}" } } };
            }
        }
    }
}
