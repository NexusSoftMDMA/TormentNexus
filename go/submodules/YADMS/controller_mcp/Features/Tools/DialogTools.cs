using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Drawing;
using System.Threading;
using System.Threading.Tasks;
using System.Windows.Forms;
using ModelContextProtocol;
using ModelContextProtocol.Protocol;
using ModelContextProtocol.Server;

namespace controller_mcp.Features.Tools
{
    public static class DialogTools
    {
        [McpServerTool, Description("Spawns a native Windows popup input box on the screen, forcing the user to type an answer. The server halts until the user clicks OK or Cancel, and returns the typed string.")]
        public static async Task<CallToolResult> ShowInputPrompt(string title, string message)
        {
            try
            {
                return await Task.Run(() =>
                {
                    string resultStr = null;
                    Exception ex = null;
                    
                    Thread thread = new Thread(() =>
                    {
                        try
                        {
                            using (Form prompt = new Form())
                            {
                                prompt.Width = 500;
                                prompt.Height = 200;
                                prompt.FormBorderStyle = FormBorderStyle.FixedDialog;
                                prompt.Text = title;
                                prompt.StartPosition = FormStartPosition.CenterScreen;
                                prompt.TopMost = true;

                                Label textLabel = new Label() { Left = 20, Top = 20, Width = 440, Height = 40, Text = message };
                                TextBox textBox = new TextBox() { Left = 20, Top = 70, Width = 440 };
                                Button confirmation = new Button() { Text = "OK", Left = 360, Width = 100, Top = 120, DialogResult = DialogResult.OK };
                                Button cancel = new Button() { Text = "Cancel", Left = 240, Width = 100, Top = 120, DialogResult = DialogResult.Cancel };

                                confirmation.Click += (sender, e) => { prompt.Close(); };
                                cancel.Click += (sender, e) => { prompt.Close(); };

                                prompt.Controls.Add(textBox);
                                prompt.Controls.Add(confirmation);
                                prompt.Controls.Add(cancel);
                                prompt.Controls.Add(textLabel);
                                prompt.AcceptButton = confirmation;
                                prompt.CancelButton = cancel;

                                DialogResult dialogResult = prompt.ShowDialog();
                                
                                if (dialogResult == DialogResult.OK)
                                {
                                    resultStr = textBox.Text;
                                }
                                else
                                {
                                    resultStr = "[User Cancelled]";
                                }
                            }
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

                    return new CallToolResult
                    {
                        Content = new List<ContentBlock> { new TextContentBlock { Text = resultStr } }
                    };
                });
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = ex.Message } } };
            }
        }

        [McpServerTool, Description("Spawns a native Windows MessageBox to ask the user a Yes/No/Cancel question. Returns their choice.")]
        public static async Task<CallToolResult> ShowMessageBox(string title, string message)
        {
            try
            {
                return await Task.Run(() =>
                {
                    string resultStr = null;
                    Exception ex = null;

                    Thread thread = new Thread(() =>
                    {
                        try
                        {
                            DialogResult result = MessageBox.Show(message, title, MessageBoxButtons.YesNoCancel, MessageBoxIcon.Question, MessageBoxDefaultButton.Button1, MessageBoxOptions.DefaultDesktopOnly);
                            resultStr = result.ToString();
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

                    return new CallToolResult
                    {
                        Content = new List<ContentBlock> { new TextContentBlock { Text = resultStr } }
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
