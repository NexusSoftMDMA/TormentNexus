using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.IO;
using System.IO.Compression;
using System.Security.Cryptography;
using System.Threading.Tasks;
using ModelContextProtocol;
using ModelContextProtocol.Protocol;
using ModelContextProtocol.Server;

namespace controller_mcp.Features.Tools
{
    public static class DeveloperTools
    {
        [McpServerTool, Description("Calculates the checksum (MD5 or SHA256) of a file to verify build integrity.")]
        public static async Task<CallToolResult> ComputeFileHash(string file_path, string algorithm = "SHA256")
        {
            try
            {
                string safePath = InputValidator.ValidateFilePath(file_path, nameof(file_path));

                if (!File.Exists(safePath))
                {
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"File not found: {safePath}" } } };
                }

                return await Task.Run(() =>
                {
                    using (var stream = File.OpenRead(safePath))
                    {
                        HashAlgorithm hashAlgo;
                        if (algorithm.Equals("MD5", StringComparison.OrdinalIgnoreCase))
                            hashAlgo = MD5.Create();
                        else
                            hashAlgo = SHA256.Create();

                        using (hashAlgo)
                        {
                            var hash = hashAlgo.ComputeHash(stream);
                            string hex = BitConverter.ToString(hash).Replace("-", "").ToLowerInvariant();

                            return new CallToolResult
                            {
                                Content = new List<ContentBlock> { new TextContentBlock { Text = hex } }
                            };
                        }
                    }
                });
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = ex.Message } } };
            }
        }

        [McpServerTool, Description("Packages up an entire directory into a compressed .zip archive.")]
        public static async Task<CallToolResult> ZipDirectory(string source_dir, string dest_zip)
        {
            try
            {
                string safeSource = InputValidator.ValidateFilePath(source_dir, nameof(source_dir));
                string safeDest = InputValidator.ValidateFilePath(dest_zip, nameof(dest_zip));

                if (!Directory.Exists(safeSource))
                {
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Source directory not found: {safeSource}" } } };
                }

                string dir = Path.GetDirectoryName(safeDest);
                if (!string.IsNullOrEmpty(dir) && !Directory.Exists(dir))
                    Directory.CreateDirectory(dir);

                if (File.Exists(safeDest))
                    File.Delete(safeDest);

                return await Task.Run(() =>
                {
                    ZipFile.CreateFromDirectory(safeSource, safeDest, CompressionLevel.Optimal, false);
                    
                    return new CallToolResult
                    {
                        Content = new List<ContentBlock> { new TextContentBlock { Text = $"Successfully zipped {safeSource} to {safeDest}" } }
                    };
                });
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = ex.Message } } };
            }
        }

        [McpServerTool, Description("Extracts a .zip archive into a target directory.")]
        public static async Task<CallToolResult> UnzipArchive(string source_zip, string dest_dir)
        {
            try
            {
                string safeSource = InputValidator.ValidateFilePath(source_zip, nameof(source_zip));
                string safeDest = InputValidator.ValidateFilePath(dest_dir, nameof(dest_dir));

                if (!File.Exists(safeSource))
                {
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Zip file not found: {safeSource}" } } };
                }

                if (!Directory.Exists(safeDest))
                    Directory.CreateDirectory(safeDest);

                return await Task.Run(() =>
                {
                    ZipFile.ExtractToDirectory(safeSource, safeDest);
                    
                    return new CallToolResult
                    {
                        Content = new List<ContentBlock> { new TextContentBlock { Text = $"Successfully extracted {safeSource} to {safeDest}" } }
                    };
                });
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = ex.Message } } };
            }
        }
    }
}
