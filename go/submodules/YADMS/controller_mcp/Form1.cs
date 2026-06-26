using System;
using System.Drawing;
using System.Text.Json;
using System.Threading;
using System.Threading.Tasks;
using System.Windows.Forms;
using Microsoft.Extensions.DependencyInjection;
using ModelContextProtocol.Server;
using ModelContextProtocol.Protocol;
using System.ComponentModel;
using System.IO;
using System.IO.Compression;

namespace controller_mcp
{
    public partial class Form1 : Form
    {
        public static McpServer Server { get; private set; }
        public static Form1 Instance { get; private set; }

        private HttpListenerSseServerTransport _transport;
        private McpServer _server;
        private CancellationTokenSource _cts;

        private bool _isDaemon;
        private bool _isMirror;
        private NotifyIcon _notifyIcon;
        private System.Windows.Forms.Timer _uiTimer;
        private System.Windows.Forms.Timer _updateTimer;

        public Form1(bool isDaemon = false)
        {
            InitializeComponent();
            controller_mcp.Features.Tools.AuditLogger.OnLogWritten += AuditLogger_OnLogWritten;
            SetupTabs();
            Instance = this;
            _isDaemon = isDaemon;

            SetupNotifyIcon();
            this.FormClosing += Form1_FormClosing;

            if (_isDaemon)
            {
                // Headless Daemon
                this.WindowState = FormWindowState.Minimized;
                this.ShowInTaskbar = false;
                this.Opacity = 0;
                
                IpcManager.StartServer(OnIpcCommand);
                btnStart_Click(null, null); // Auto-start the server!
            }
            else
            {
                if (IpcManager.TryConnectClient(Log))
                {
                    _isMirror = true;
                    btnStart.Text = "Mirror Mode (Active)";
                    btnStart.Enabled = false;
                    btnStop.Enabled = true;
                    txtPort.Enabled = false;
                    lblStatus.Text = "Connected to Daemon";
                    lblStatus.ForeColor = Color.Blue;
                }
            }
        }

        private void Form1_FormClosing(object sender, FormClosingEventArgs e)
        {
            // Unhook static events to prevent ObjectDisposedException
            controller_mcp.Features.Tools.AuditLogger.OnLogWritten -= AuditLogger_OnLogWritten;

            // Stop UI timers
            _uiTimer?.Stop();
            _uiTimer?.Dispose();
            _updateTimer?.Stop();
            _updateTimer?.Dispose();

            // Cleanly shutdown the background server and all active tools
            GracefulShutdown();
        }

        private void SetupTabs()
        {
            TabControl tabControl = new TabControl();
            tabControl.Bounds = this.ClientRectangle;
            tabControl.Anchor = AnchorStyles.Top | AnchorStyles.Bottom | AnchorStyles.Left | AnchorStyles.Right;

            TabPage tabLogs = new TabPage("Logs & Controls");
            TabPage tabSettings = new TabPage("Settings");

            tabLogs.Controls.Add(this.btnStart);
            tabLogs.Controls.Add(this.btnStop);
            tabLogs.Controls.Add(this.lblStatus);
            tabLogs.Controls.Add(this.txtPort);
            tabLogs.Controls.Add(this.lblPort);
            tabLogs.Controls.Add(this.rtbLog);
            tabLogs.Controls.Add(this.txtConfig);
            tabLogs.Controls.Add(this.lblConfig);
            tabLogs.Controls.Add(this.btnCopyConfig);

            Label lblDaemon = new Label { Text = "Daemon Installed: " + (controller_mcp.Features.Tools.DaemonTools.IsServiceInstalled() ? "Yes" : "No"), Location = new Point(500, 17), AutoSize = true, Font = new Font(this.Font, FontStyle.Bold), ForeColor = controller_mcp.Features.Tools.DaemonTools.IsServiceInstalled() ? Color.Green : Color.Gray };
            tabLogs.Controls.Add(lblDaemon);

            TabPage tabAnalytics = new TabPage("Analytics");

            BuildSettingsTab(tabSettings);
            BuildAnalyticsTab(tabAnalytics);

            tabControl.TabPages.Add(tabLogs);
            tabControl.TabPages.Add(tabSettings);
            tabControl.TabPages.Add(tabAnalytics);
            this.Controls.Add(tabControl);
        }

        private void BuildSettingsTab(TabPage tab)
        {
            var settings = AppSettings.Load();

            Label lblLog = new Label { Text = "Log Directory:", Location = new Point(20, 20), AutoSize = true };
            TextBox txtLog = new TextBox { Text = settings.LogDirectory, Location = new Point(20, 40), Width = 300 };
            Button btnBrowseLog = new Button { Text = "Browse", Location = new Point(330, 38) };
            btnBrowseLog.Click += (s, e) => {
                using (var fbd = new FolderBrowserDialog()) {
                    if (fbd.ShowDialog() == DialogResult.OK) txtLog.Text = fbd.SelectedPath;
                }
            };

            tab.Controls.Add(lblLog);
            tab.Controls.Add(txtLog);
            tab.Controls.Add(btnBrowseLog);

#if !BUILD_WITH_FFMPEG
            Label lblFfmpeg = new Label { Text = "FFmpeg Path:", Location = new Point(20, 80), AutoSize = true };
            TextBox txtFfmpeg = new TextBox { Text = settings.FFmpegPath, Location = new Point(20, 100), Width = 300 };
            Button btnBrowseFfmpeg = new Button { Text = "Browse", Location = new Point(330, 98) };
            Button btnInstallFfmpeg = new Button { Text = "Install FFmpeg", Location = new Point(410, 98), Width = 100 };

            btnBrowseFfmpeg.Click += (s, e) => {
                using (var ofd = new OpenFileDialog { Filter = "Executable|*.exe" }) {
                    if (ofd.ShowDialog() == DialogResult.OK) txtFfmpeg.Text = ofd.FileName;
                }
            };

            btnInstallFfmpeg.Click += async (s, e) => {
                try {
                    btnInstallFfmpeg.Enabled = false;
                    btnInstallFfmpeg.Text = "Downloading...";
                    string targetDir = Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "ffmpeg");
                    if (!Directory.Exists(targetDir)) Directory.CreateDirectory(targetDir);
                    
                    controller_mcp.Features.Tools.AuditLogger.Log(controller_mcp.Features.Tools.LogLevel.INFO, "System", "Starting FFmpeg download...");

                    System.Net.ServicePointManager.SecurityProtocol = System.Net.SecurityProtocolType.Tls12;
                    string url = "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-win64-gpl.zip";
                    string zipPath = Path.Combine(targetDir, "ffmpeg.zip");
                    
                    using (var client = new System.Net.WebClient()) {
                        await client.DownloadFileTaskAsync(url, zipPath);
                    }
                    
                    controller_mcp.Features.Tools.AuditLogger.Log(controller_mcp.Features.Tools.LogLevel.INFO, "System", "Extracting FFmpeg...");

                    using (var archive = System.IO.Compression.ZipFile.OpenRead(zipPath))
                    {
                        foreach (var entry in archive.Entries)
                        {
                            string destinationPath = Path.GetFullPath(Path.Combine(targetDir, entry.FullName));
                            if (destinationPath.StartsWith(targetDir, StringComparison.OrdinalIgnoreCase))
                            {
                                if (string.IsNullOrEmpty(entry.Name))
                                {
                                    Directory.CreateDirectory(destinationPath);
                                }
                                else
                                {
                                    Directory.CreateDirectory(Path.GetDirectoryName(destinationPath));
                                    entry.ExtractToFile(destinationPath, overwrite: true);
                                }
                            }
                        }
                    }
                    File.Delete(zipPath);
                    
                    string[] files = Directory.GetFiles(targetDir, "ffmpeg.exe", SearchOption.AllDirectories);
                    if (files.Length > 0) {
                        txtFfmpeg.Text = files[0];
                        var appSettings = AppSettings.Load();
                        appSettings.FFmpegPath = files[0];
                        appSettings.Save();
                        controller_mcp.Features.Tools.AuditLogger.Log(controller_mcp.Features.Tools.LogLevel.INFO, "System", "FFmpeg Installed Successfully and saved to configuration!");
                        MessageBox.Show("FFmpeg Installed Successfully!");
                    }
                } catch (Exception ex) {
                    controller_mcp.Features.Tools.AuditLogger.Log(controller_mcp.Features.Tools.LogLevel.ERROR, "System", $"FFmpeg Install failed: {ex.Message}");
                    MessageBox.Show("Install failed: " + ex.Message);
                } finally {
                    btnInstallFfmpeg.Enabled = true;
                    btnInstallFfmpeg.Text = "Install FFmpeg";
                }
            };

            tab.Controls.Add(lblFfmpeg);
            tab.Controls.Add(txtFfmpeg);
            tab.Controls.Add(btnBrowseFfmpeg);
            tab.Controls.Add(btnInstallFfmpeg);
#endif

            CheckBox chkDebug = new CheckBox { Text = "Enable Debug Logging", Location = new Point(20, 140), AutoSize = true, Checked = settings.EnableDebugLogging };
            tab.Controls.Add(chkDebug);

            Button btnSave = new Button { Text = "Save Settings", Location = new Point(20, 170), Width = 120 };
            btnSave.Click += (s, e) => {
                settings.LogDirectory = txtLog.Text;
                settings.EnableDebugLogging = chkDebug.Checked;
#if !BUILD_WITH_FFMPEG
                settings.FFmpegPath = txtFfmpeg.Text;
                
                if (txtFfmpeg.Text.Trim().ToLower() == "exit") {
                     // Pass validation per user request
                } else if (!string.IsNullOrWhiteSpace(txtFfmpeg.Text) && !File.Exists(txtFfmpeg.Text)) {
                     MessageBox.Show("Error: FFmpeg path does not exist!");
                     return;
                }
#endif
                settings.Save();
                controller_mcp.Features.Tools.AuditLogger.Reconfigure(settings.LogDirectory);
                controller_mcp.Features.Tools.AuditLogger.Log(controller_mcp.Features.Tools.LogLevel.INFO, "System", "System configuration explicitly saved by user.");
                MessageBox.Show("Settings Saved!");
            };

            tab.Controls.Add(btnSave);

            // AUTO UPDATER UI
            Label lblCurrentVer = new Label { Text = $"Current Version: {VersionInfo.CurrentVersion}", Location = new Point(360, 140), AutoSize = true, Font = new Font(this.Font, FontStyle.Bold) };
            Label lblLatestVer = new Label { Text = "Latest Version: Checking...", Location = new Point(360, 160), AutoSize = true };
            Label lblLastChecked = new Label { Text = $"Last Checked: {(settings.LastUpdateCheck > DateTime.MinValue ? settings.LastUpdateCheck.ToString("g") : "Never")}", Location = new Point(360, 180), AutoSize = true };
            
            CheckBox chkAutoUpdate = new CheckBox { Text = "Enable Auto-Updater (Checks every 30 mins)", Location = new Point(360, 200), AutoSize = true, Checked = settings.EnableAutoUpdate };
            chkAutoUpdate.CheckedChanged += (s, e) => {
                settings.EnableAutoUpdate = chkAutoUpdate.Checked;
                settings.Save();
            };

            Button btnCheckUpdate = new Button { Text = "Check For Updates", Location = new Point(360, 225), Width = 150 };

            Action performUpdateCheck = async () => {
                btnCheckUpdate.Enabled = false;
                btnCheckUpdate.Text = "Checking...";
                lblLatestVer.Text = "Latest Version: Checking...";
                try {
                    var (available, latestVersion, downloadUrl, errorReason) = await controller_mcp.Features.Tools.UpdateManager.CheckForUpdatesAsync();
                    
                    settings.LastUpdateCheck = DateTime.Now;
                    settings.Save();
                    lblLastChecked.Text = $"Last Checked: {settings.LastUpdateCheck:g}";
                    
                    if (!string.IsNullOrEmpty(errorReason)) {
                        lblLatestVer.Text = $"Latest Version: {errorReason}";
                    } else if (available) {
                        lblLatestVer.Text = $"Latest Version: {latestVersion ?? "Unknown"}";
                        if (!string.IsNullOrEmpty(downloadUrl)) {
                            controller_mcp.Features.Tools.AuditLogger.LogSystemEvent("Updater", $"Update {latestVersion} found. Initiating silent download: {downloadUrl}");
                            lblLatestVer.Text += " (Downloading...)";
                            bool success = await controller_mcp.Features.Tools.UpdateManager.DownloadAndInstallUpdateAsync(downloadUrl, latestVersion);
                            if (!success) lblLatestVer.Text = $"Latest Version: {latestVersion} (Download Failed)";
                        } else {
                            lblLatestVer.Text = $"Latest Version: {latestVersion} (No asset for {controller_mcp.Features.Tools.UpdateManager.GetCurrentSku()})";
                        }
                    } else {
                        lblLatestVer.Text = $"Latest Version: {VersionInfo.CurrentVersion} (Up to date)";
                    }
                } catch (Exception ex) {
                    lblLatestVer.Text = "Latest Version: Error";
                    controller_mcp.Features.Tools.AuditLogger.Log(controller_mcp.Features.Tools.LogLevel.ERROR, "System", $"Update check failed: {ex.Message}");
                } finally {
                    btnCheckUpdate.Text = "Check For Updates";
                    btnCheckUpdate.Enabled = true;
                }
            };

            btnCheckUpdate.Click += (s, e) => performUpdateCheck();

            tab.Controls.Add(lblCurrentVer);
            tab.Controls.Add(lblLatestVer);
            tab.Controls.Add(lblLastChecked);
            tab.Controls.Add(chkAutoUpdate);
            tab.Controls.Add(btnCheckUpdate);

            _updateTimer = new System.Windows.Forms.Timer { Interval = 30 * 60 * 1000 }; // 30 minutes
            _updateTimer.Tick += (s, e) => {
                if (settings.EnableAutoUpdate) {
                    performUpdateCheck();
                }
            };
            _updateTimer.Start();
            
            // Do an initial check 5 seconds after startup if enabled
            _ = Task.Run(async () => {
                await Task.Delay(5000);
                if (settings.EnableAutoUpdate && InvokeRequired) {
                    Invoke(new Action(() => performUpdateCheck()));
                }
            });

            Button btnExportKey = new Button { Text = "Export Master Key", Location = new Point(20, 210), Width = 150 };
            btnExportKey.Click += (s, e) => {
                if (settings.MasterKeyExported)
                {
                    MessageBox.Show("You have already exported the Master Key. For security reasons, it can only be exported once.", "Already Exported", MessageBoxButtons.OK, MessageBoxIcon.Warning);
                    return;
                }

                var res = MessageBox.Show("WARNING: This key grants full access to your saved passwords and state backups. Save it securely. It can only be exported once. Proceed?", "Export Master Key", MessageBoxButtons.YesNo, MessageBoxIcon.Warning);
                if (res == DialogResult.Yes)
                {
                    try
                    {
                        string b64 = controller_mcp.Features.Tools.StateBackupManager.ExportMasterKey();
                        settings.MasterKeyExported = true;
                        settings.Save();
                        MessageBox.Show($"MASTER KEY (Save this now!):\n\n{b64}", "Master Key Exported", MessageBoxButtons.OK, MessageBoxIcon.Information);
                        btnExportKey.Enabled = false;
                    }
                    catch (Exception ex)
                    {
                        MessageBox.Show($"Failed: {ex.Message}", "Error", MessageBoxButtons.OK, MessageBoxIcon.Error);
                    }
                }
            };
            tab.Controls.Add(btnExportKey);

            Button btnImportKey = new Button { Text = "Import Master Key", Location = new Point(190, 210), Width = 150 };
            btnImportKey.Click += (s, e) => {
                string input = "";
                using (Form prompt = new Form() { Width = 500, Height = 150, FormBorderStyle = FormBorderStyle.FixedDialog, Text = "Import Master Key", StartPosition = FormStartPosition.CenterScreen })
                {
                    Label textLabel = new Label() { Left = 20, Top = 20, Width = 450, Text = "Enter the Base64 Master Key to import and bind to this machine:" };
                    TextBox textBox = new TextBox() { Left = 20, Top = 50, Width = 450 };
                    Button confirmation = new Button() { Text = "Ok", Left = 350, Width = 100, Top = 80, DialogResult = DialogResult.OK };
                    prompt.Controls.Add(textBox);
                    prompt.Controls.Add(confirmation);
                    prompt.Controls.Add(textLabel);
                    prompt.AcceptButton = confirmation;

                    if (prompt.ShowDialog() == DialogResult.OK)
                    {
                        input = textBox.Text;
                    }
                }

                if (!string.IsNullOrWhiteSpace(input))
                {
                    try
                    {
                        controller_mcp.Features.Tools.StateBackupManager.ImportMasterKey(input);
                        MessageBox.Show("Master Key imported and sealed to this machine successfully. You can now start the daemon.", "Success", MessageBoxButtons.OK, MessageBoxIcon.Information);
                    }
                    catch (Exception ex)
                    {
                        MessageBox.Show($"Import failed: {ex.Message}", "Error", MessageBoxButtons.OK, MessageBoxIcon.Error);
                    }
                }
            };
            tab.Controls.Add(btnImportKey);

            Button btnInstallDaemon = new Button { Text = "Install Background Daemon", Location = new Point(20, 250), Width = 160 };
            btnInstallDaemon.Click += (s, e) => {
                var res = controller_mcp.Features.Tools.DaemonTools.InstallAsService();
                if (res.IsError == true) MessageBox.Show("Failed to install Daemon. Please run as Administrator.", "Error", MessageBoxButtons.OK, MessageBoxIcon.Error);
                else MessageBox.Show("Daemon installed successfully. It will automatically start when you log in.", "Success", MessageBoxButtons.OK, MessageBoxIcon.Information);
            };
            tab.Controls.Add(btnInstallDaemon);

            Button btnRemoveDaemon = new Button { Text = "Remove Background Daemon", Location = new Point(190, 250), Width = 160 };
            btnRemoveDaemon.Click += (s, e) => {
                var res = controller_mcp.Features.Tools.DaemonTools.RemoveAsService();
                if (res.IsError == true) MessageBox.Show("Failed to remove Daemon. Please run as Administrator.", "Error", MessageBoxButtons.OK, MessageBoxIcon.Error);
                else MessageBox.Show("Daemon removed from background startup.", "Success", MessageBoxButtons.OK, MessageBoxIcon.Information);
            };
            tab.Controls.Add(btnRemoveDaemon);

            Label lblNpcapStatus = new Label { Location = new Point(20, 290), AutoSize = true, Font = new Font(this.Font, FontStyle.Bold) };
            Button btnInstallNpcap = new Button { Text = "Install Npcap", Location = new Point(200, 286), Width = 100 };
            
            Action updateNpcapStatus = () => {
                bool isInstalled = System.IO.File.Exists(System.IO.Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.System), "Npcap", "wpcap.dll"));
                if (isInstalled) {
                    lblNpcapStatus.Text = "Npcap Driver: Installed";
                    lblNpcapStatus.ForeColor = Color.Green;
                    btnInstallNpcap.Text = "Uninstall Npcap";
                    btnInstallNpcap.Enabled = true;
                } else {
                    lblNpcapStatus.Text = "Npcap Driver: Missing";
                    lblNpcapStatus.ForeColor = Color.Red;
                    btnInstallNpcap.Text = "Install Npcap";
                    btnInstallNpcap.Enabled = true;
                }
            };
            
            updateNpcapStatus();

            btnInstallNpcap.Click += async (s, e) => {
                bool isInstalling = btnInstallNpcap.Text.StartsWith("Install");
                btnInstallNpcap.Enabled = false;
                btnInstallNpcap.Text = isInstalling ? "Installing..." : "Uninstalling...";
                try {
                    if (isInstalling) {
                        await controller_mcp.Features.Tools.PcapTools.InstallNpcap();
                    } else {
                        await controller_mcp.Features.Tools.PcapTools.UninstallNpcap();
                    }
                    
                    bool isInstalled = System.IO.File.Exists(System.IO.Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.System), "Npcap", "wpcap.dll"));
                    
                    if (isInstalling && isInstalled) {
                        MessageBox.Show("Npcap installation completed successfully!", "Success", MessageBoxButtons.OK, MessageBoxIcon.Information);
                    } else if (isInstalling && !isInstalled) {
                        MessageBox.Show("Npcap installation was aborted or failed.", "Installation Failed", MessageBoxButtons.OK, MessageBoxIcon.Warning);
                    } else if (!isInstalling && !isInstalled) {
                        MessageBox.Show("Npcap uninstallation completed successfully!", "Success", MessageBoxButtons.OK, MessageBoxIcon.Information);
                    } else if (!isInstalling && isInstalled) {
                        MessageBox.Show("Npcap uninstallation was aborted or failed.", "Uninstallation Failed", MessageBoxButtons.OK, MessageBoxIcon.Warning);
                    }
                } catch (Exception ex) {
                    MessageBox.Show($"Failed: {ex.Message}", "Error", MessageBoxButtons.OK, MessageBoxIcon.Error);
                } finally {
                    updateNpcapStatus();
                }
            };

            tab.Controls.Add(lblNpcapStatus);
            tab.Controls.Add(btnInstallNpcap);
        }

        private void BuildAnalyticsTab(TabPage tab)
        {
            Label lblUptime = new Label { Text = "Total Uptime: 00:00:00", Location = new Point(20, 20), AutoSize = true, Font = new Font(this.Font, FontStyle.Bold) };
            Label lblRequests = new Label { Text = "Total Requests Processed: 0", Location = new Point(20, 50), AutoSize = true };
            Label lblErrors = new Label { Text = "Total Errors: 0", Location = new Point(20, 80), AutoSize = true };
            Label lblBytes = new Label { Text = "Bytes Sent: 0 | Received: 0", Location = new Point(20, 110), AutoSize = true };

            Label lblGrid = new Label { Text = "Tool Invocations:", Location = new Point(20, 150), AutoSize = true };
            ListView lvTools = new ListView { Location = new Point(20, 170), Width = 400, Height = 200, View = View.Details, FullRowSelect = true, GridLines = true };
            lvTools.Columns.Add("Tool Name", 250);
            lvTools.Columns.Add("Invocation Count", 120);

            tab.Controls.Add(lblUptime);
            tab.Controls.Add(lblRequests);
            tab.Controls.Add(lblErrors);
            tab.Controls.Add(lblBytes);
            tab.Controls.Add(lblGrid);
            tab.Controls.Add(lvTools);

            _uiTimer = new System.Windows.Forms.Timer { Interval = 1000 };
            _uiTimer.Tick += (s, e) => {
                var analytics = controller_mcp.Features.Tools.AnalyticsManager.Current;
                var uptime = DateTime.UtcNow - analytics.StartTime;
                lblUptime.Text = $"Total Uptime: {(int)uptime.TotalHours:D2}:{uptime.Minutes:D2}:{uptime.Seconds:D2}";
                lblRequests.Text = $"Total Requests Processed: {analytics.TotalRequestsProcessed}";
                lblErrors.Text = $"Total Errors: {analytics.TotalErrors}";
                lblBytes.Text = $"Bytes Sent: {analytics.TotalBytesSent:N0} | Received: {analytics.TotalBytesReceived:N0}";

                lvTools.BeginUpdate();
                lvTools.Items.Clear();
                foreach (var kvp in analytics.ToolInvocations)
                {
                    lvTools.Items.Add(new ListViewItem(new[] { kvp.Key, kvp.Value.ToString() }));
                }
                lvTools.EndUpdate();
            };
            _uiTimer.Start();
        }

        private void AuditLogger_OnLogWritten(string logMessage)
        {
            if (InvokeRequired)
            {
                Invoke(new Action(() => AuditLogger_OnLogWritten(logMessage)));
                return;
            }
            rtbLog.AppendText(logMessage + Environment.NewLine);

            if (_isDaemon)
                IpcManager.BroadcastLog(logMessage);
            
            if (rtbLog.Lines.Length > 5000)
            {
                rtbLog.Select(0, rtbLog.GetFirstCharIndexFromLine(rtbLog.Lines.Length - 5000));
                rtbLog.SelectedText = "";
            }
            
            rtbLog.ScrollToCaret();
        }

        private void SetupNotifyIcon()
        {
            _notifyIcon = new NotifyIcon();
            _notifyIcon.Icon = SystemIcons.Application;
            _notifyIcon.Visible = true;
            _notifyIcon.Text = "Controller MCP";
            
            var menu = new ContextMenu();
            menu.MenuItems.Add("Show", (s, e) => { this.Show(); this.WindowState = FormWindowState.Normal; this.Opacity = 1; this.ShowInTaskbar = true; });
            menu.MenuItems.Add("Exit", (s, e) => { Application.Exit(); });
            
            _notifyIcon.ContextMenu = menu;
            _notifyIcon.DoubleClick += (s, e) => { this.Show(); this.WindowState = FormWindowState.Normal; this.Opacity = 1; this.ShowInTaskbar = true; };

            this.FormClosing += (s, e) =>
            {
                if (e.CloseReason == CloseReason.UserClosing)
                {
                    if (_isMirror)
                    {
                        var result = MessageBox.Show("Kill the background Daemon too?", "Exit", MessageBoxButtons.YesNoCancel);
                        if (result == DialogResult.Cancel) { e.Cancel = true; return; }
                        if (result == DialogResult.Yes) { IpcManager.SendCommand("EXIT"); Application.Exit(); }
                        if (result == DialogResult.No) { Application.Exit(); }
                    }
                    else
                    {
                        e.Cancel = true;
                        this.Hide();
                        _notifyIcon.ShowBalloonTip(3000, "Minimized", "The server is still running in the background.", ToolTipIcon.Info);
                    }
                }
            };

            AppDomain.CurrentDomain.ProcessExit += (s, e) => GracefulShutdown();
        }

        protected override void OnFormClosed(FormClosedEventArgs e)
        {
            GracefulShutdown();
            base.OnFormClosed(e);
        }

        private void OnIpcCommand(string cmd)
        {
            if (cmd == "EXIT")
            {
                if (InvokeRequired) { Invoke(new Action(Application.Exit)); }
                else { Application.Exit(); }
            }
        }

        public void Log(string message)
        {
            if (InvokeRequired)
            {
                Invoke(new Action<string>(Log), message);
                return;
            }

            controller_mcp.Features.Tools.AuditLogger.LogSystemEvent("DebugConsole", message);
        }

        public void ShowBalloonTip(string title, string message)
        {
            if (InvokeRequired) { Invoke(new Action<string, string>(ShowBalloonTip), title, message); return; }
            _notifyIcon?.ShowBalloonTip(5000, title, message, ToolTipIcon.Info);
        }

        private async void btnStart_Click(object sender, EventArgs e)
        {
            if (!int.TryParse(txtPort.Text, out int port) || port < 1 || port > 65535)
            {
                MessageBox.Show("Please enter a valid port number.", "Invalid Port", MessageBoxButtons.OK, MessageBoxIcon.Warning);
                return;
            }

            btnStart.Enabled = false;
            txtPort.Enabled = false;

            try
            {
                _transport = new HttpListenerSseServerTransport(port);
                _transport.OnLog = Log;

                // Set up the DI container
                var services = new ServiceCollection();
                services.AddLogging();
                services.AddSingleton<ITransport>(_transport);
                services.AddMcpServer(options => 
                {
                    options.Capabilities = new ServerCapabilities
                    {
                        Logging = new LoggingCapability()
                    };
                })
                .WithToolsFromAssembly();

                var provider = services.BuildServiceProvider();
                
                var options = provider.GetRequiredService<Microsoft.Extensions.Options.IOptions<McpServerOptions>>().Value;
                var loggerFactory = provider.GetService<Microsoft.Extensions.Logging.ILoggerFactory>();
                
                _server = McpServer.Create(_transport, options, loggerFactory, provider);
                Server = _server;

                _cts = new CancellationTokenSource();

                // Start the transport listening
                await _transport.StartAsync(_cts.Token);
                
                // Connect the MCP server to the transport
                _ = Task.Run(() => _server.RunAsync(_cts.Token), _cts.Token);

                lblStatus.Text = "Online";
                lblStatus.ForeColor = Color.Green;
                btnStop.Enabled = true;

                // Update Config Snippet
                var config = new
                {
                    mcpServers = new
                    {
                        controller = new
                        {
                            command = "curl", // Not strictly a command since it's SSE, some AI clients might require specific SSE structure. 
                                              // We'll just provide the raw SSE url as an example
                            url = $"http://localhost:{port}/mcp/sse"
                        }
                    }
                };

                // Claude desktop config usually takes 'command', but for SSE:
                // We'll provide a generic JSON config that AI clients typically understand.
                txtConfig.Text = "{\r\n  \"mcpServers\": {\r\n    \"controller_mcp\": {\r\n      \"url\": \"http://localhost:" + port + "/mcp/sse\"\r\n    }\r\n  }\r\n}";

                Log("MCP Server started successfully.");
                controller_mcp.Features.Tools.StateBackupManager.RestoreState();
            }
            catch (Exception ex)
            {
                Log($"Failed to start MCP Server: {ex.Message}");
                GracefulShutdown();
            }
        }

        private void btnStop_Click(object sender, EventArgs e)
        {
            if (_isMirror)
            {
                var result = MessageBox.Show("Are you sure you want to kill the Background Daemon?", "Kill Daemon", MessageBoxButtons.YesNo, MessageBoxIcon.Warning);
                if (result == DialogResult.Yes)
                {
                    IpcManager.SendCommand("EXIT");
                    Log("Sent EXIT command to Daemon.");
                    _isMirror = false;
                    btnStart.Text = "Start";
                    btnStart.Enabled = true;
                    btnStop.Enabled = false;
                    txtPort.Enabled = true;
                    lblStatus.Text = "Offline";
                    lblStatus.ForeColor = Color.Red;
                }
            }
            else
            {
                GracefulShutdown();
            }
        }

        public void GracefulShutdown()
        {
            try
            {
                _cts?.Cancel();
                _transport?.DisposeAsync().AsTask().Wait();
                
                if (_server is IAsyncDisposable asyncSrv)
                    asyncSrv.DisposeAsync().AsTask().Wait();
                else if (_server is IDisposable syncSrv)
                    syncSrv.Dispose();

                controller_mcp.Features.Tools.FFmpegHelper.KillAll();
                controller_mcp.Features.Tools.CronTools.StopAll();
                controller_mcp.Features.Tools.PcapTools.StopAll();
                controller_mcp.Features.Tools.StatefulRecordingTools.StopAll();
                controller_mcp.Features.Tools.TerminalTools.StopAll();
                controller_mcp.Features.Tools.WatcherTools.StopAll();
                controller_mcp.Features.Tools.SshTools.StopAll();
                controller_mcp.Features.Tools.WebSocketTools.StopAll();
            }
            catch (Exception ex)
            {
                Log($"Error during shutdown: {ex.Message}");
            }
            finally
            {
                _server = null;
                _transport = null;
                _cts = null;

                if (!this.IsDisposed && !this.Disposing)
                {
                    if (InvokeRequired)
                    {
                        try {
                            Invoke(new Action(() => 
                            {
                                lblStatus.Text = "Offline";
                                lblStatus.ForeColor = Color.Red;
                                btnStart.Enabled = true;
                                btnStop.Enabled = false;
                                txtPort.Enabled = true;
                                txtConfig.Text = string.Empty;
                                Log("MCP Server stopped.");
                            }));
                        } catch { }
                    }
                    else
                    {
                        try {
                            lblStatus.Text = "Offline";
                            lblStatus.ForeColor = Color.Red;
                            btnStart.Enabled = true;
                            btnStop.Enabled = false;
                            txtPort.Enabled = true;
                            txtConfig.Text = string.Empty;
                            Log("MCP Server stopped.");
                        } catch { }
                    }
                }
            }
        }

        private void btnCopyConfig_Click(object sender, EventArgs e)
        {
            if (!string.IsNullOrWhiteSpace(txtConfig.Text))
            {
                Clipboard.SetText(txtConfig.Text);
                MessageBox.Show("Config copied to clipboard!", "Copied", MessageBoxButtons.OK, MessageBoxIcon.Information);
            }
        }
    }

}
