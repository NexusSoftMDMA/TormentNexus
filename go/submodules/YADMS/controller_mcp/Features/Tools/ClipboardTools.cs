using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Drawing;
using System.IO;
using System.Threading;
using System.Threading.Tasks;
using System.Windows.Forms;
using ModelContextProtocol;
using ModelContextProtocol.Protocol;
using ModelContextProtocol.Server;

namespace controller_mcp.Features.Tools
{
    public static class ClipboardTools
    {
        private static T RunOnSTAThread<T>(Func<T> action)
        {
            T result = default;
            Exception ex = null;
            Thread thread = new Thread(() =>
            {
                try
                {
                    result = action();
                }
                catch (Exception e)
                {
                    ex = e;
                }
            });
            thread.SetApartmentState(ApartmentState.STA);
            thread.Start();
            thread.Join();

            if (ex != null) throw ex;
            return result;
        }

        private static void RunOnSTAThread(Action action)
        {
            Exception ex = null;
            Thread thread = new Thread(() =>
            {
                try
                {
                    action();
                }
                catch (Exception e)
                {
                    ex = e;
                }
            });
            thread.SetApartmentState(ApartmentState.STA);
            thread.Start();
            thread.Join();

            if (ex != null) throw ex;
        }

        [McpServerTool, Description("Reads the current contents of the system clipboard. Supports text, file paths, and images (returns base64).")]
        public static async Task<CallToolResult> GetClipboard()
        {
            try
            {
                return await Task.Run(() =>
                {
                    return RunOnSTAThread(() =>
                    {
                        if (Clipboard.ContainsText())
                        {
                            return new CallToolResult
                            {
                                Content = new List<ContentBlock> { new TextContentBlock { Text = $"{{\"type\":\"text\",\"content\":\"{Clipboard.GetText().Replace("\\", "\\\\").Replace("\"", "\\\"").Replace("\n", "\\n").Replace("\r", "\\r")}\"}}" } }
                            };
                        }
                        else if (Clipboard.ContainsFileDropList())
                        {
                            var files = Clipboard.GetFileDropList();
                            string[] arr = new string[files.Count];
                            files.CopyTo(arr, 0);
                            string jsonArr = "[\"" + string.Join("\",\"", arr).Replace("\\", "\\\\") + "\"]";
                            return new CallToolResult
                            {
                                Content = new List<ContentBlock> { new TextContentBlock { Text = $"{{\"type\":\"files\",\"content\":{jsonArr}}}" } }
                            };
                        }
                        else if (Clipboard.ContainsImage())
                        {
                            using (Image img = Clipboard.GetImage())
                            {
                                if (img != null)
                                {
                                    using (MemoryStream ms = new MemoryStream())
                                    {
                                        img.Save(ms, System.Drawing.Imaging.ImageFormat.Png);
                                        byte[] byteImage = ms.ToArray();
                                        string base64 = Convert.ToBase64String(byteImage);
                                        return new CallToolResult
                                        {
                                            Content = new List<ContentBlock> { new TextContentBlock { Text = $"{{\"type\":\"image\",\"content\":\"{base64}\"}}" } }
                                        };
                                    }
                                }
                            }
                        }

                        return new CallToolResult
                        {
                            Content = new List<ContentBlock> { new TextContentBlock { Text = "{\"type\":\"empty\",\"content\":\"\"}" } }
                        };
                    });
                });
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = ex.Message } } };
            }
        }

        [McpServerTool, Description("Writes text or base64 image data to the system clipboard.")]
        public static async Task<CallToolResult> SetClipboard(string content, string type = "text")
        {
            try
            {
                return await Task.Run(() =>
                {
                    RunOnSTAThread(() =>
                    {
                        Clipboard.Clear();
                        if (type.Equals("image", StringComparison.OrdinalIgnoreCase))
                        {
                            byte[] bytes = Convert.FromBase64String(content);
                            using (MemoryStream ms = new MemoryStream(bytes))
                            {
                                using (Image img = Image.FromStream(ms))
                                {
                                    Clipboard.SetImage(img);
                                }
                            }
                        }
                        else
                        {
                            Clipboard.SetText(content);
                        }
                    });

                    return new CallToolResult
                    {
                        Content = new List<ContentBlock> { new TextContentBlock { Text = "Clipboard successfully updated." } }
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
