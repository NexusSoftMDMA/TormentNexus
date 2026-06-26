#define AppName "YADMS"
#define AppPublisher "Blazium"
#define AppExeName "controller_mcp.exe"

[Setup]
AppId={{9F5AB5A7-0F9F-4D04-8933-2895CD3E646A}
AppName={#AppName}
AppVersion={#AppVersion}
AppPublisher={#AppPublisher}
DefaultDirName={autopf}\Blazium\YADMS
DisableDirPage=no
DefaultGroupName={#AppName}
DisableProgramGroupPage=yes
OutputBaseFilename=YADMS_Installer_{#SkuName}_v{#AppVersion}
OutputDir=installers
Compression=lzma
SolidCompression=yes
WizardStyle=modern
LicenseFile=LICENSE.txt
InfoAfterFile=SOCIALS.txt
UninstallDisplayIcon={app}\{#AppExeName}
AppSupportURL=https://blazium.app/
AppUpdatesURL=https://github.com/blazium-games/

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked
Name: "installservice"; Description: "Install YADMS as a background Daemon (Runs silently on Startup)"; GroupDescription: "Background Service"

[Registry]
Root: HKCU; Subkey: "SOFTWARE\Microsoft\Windows\CurrentVersion\Run"; ValueType: string; ValueName: "ControllerMcpDaemon"; ValueData: """{app}\{#AppExeName}"" --daemon"; Tasks: installservice; Flags: uninsdeletevalue

[Files]
Source: "artifacts\{#SkuName}\{#AppExeName}"; DestDir: "{app}"; Flags: ignoreversion
Source: "artifacts\{#SkuName}\version.txt"; DestDir: "{app}"; Flags: ignoreversion
Source: "artifacts\{#SkuName}\YADMS_Updater.exe"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{autoprograms}\{#AppName}"; Filename: "{app}\{#AppExeName}"
Name: "{autodesktop}\{#AppName}"; Filename: "{app}\{#AppExeName}"; Tasks: desktopicon

[Run]
Filename: "{app}\{#AppExeName}"; Description: "{cm:LaunchProgram,{#StringChange(AppName, '&', '&&')}}"; WorkingDir: "{app}"; Flags: nowait postinstall
Filename: "https://discord.gg/Vbq4zGdaVs"; Description: "Join the Discord Server"; Flags: shellexec postinstall skipifsilent
