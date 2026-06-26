using System;
using System.Threading.Tasks;
using Xunit;
using controller_mcp.Features.Tools;
using System.IO;
using System.Diagnostics;

namespace controller_mcp.Tests
{
    public class Phase21ToolsTests
    {
        [Fact]
        public async Task AccessibilityTools_InvokeUiElement_InvalidElement_ReturnsError()
        {
            var result = await AccessibilityTools.InvokeUiElement("NonExistentWindow123", "FakeButton");
            Assert.True(result.IsError == true);
        }

        [Fact]
        public void Base64Tools_DecodeFileBase64_InvalidPath_ReturnsError()
        {
            var result = Base64Tools.DecodeFileBase64("SGVsbG8gV29ybGQ=", "Z:\\invalid_drive\\test.txt");
            Assert.True(result.IsError == true);
        }

        [Fact]
        public async Task ClipboardTools_SetClipboard_InvalidImage_ReturnsError()
        {
            var result = await ClipboardTools.SetClipboard("invalid_base64_data!!", "image");
            Assert.True(result.IsError == true);
        }

        [Fact]
        public void CronTools_ScheduleRecurringTask_InvalidInterval_ReturnsError()
        {
            var result = CronTools.ScheduleRecurringTask("command", "echo test", -5);
            Assert.True(result.IsError == true);
        }

#if GAME_HACKING
        [Fact]
        public async Task DecompilerTools_AnalyzeDotNetAssembly_InvalidPath_ReturnsError()
        {
            var result = await DecompilerTools.AnalyzeDotNetAssembly("C:\\invalid_file.dll");
            Assert.True(result.IsError == true);
        }
#endif

        [Fact]
        public async Task DumpTools_AnalyzeDotNetDump_InvalidPath_ReturnsError()
        {
            var result = await DumpTools.AnalyzeDotNetDump("C:\\invalid_dump.dmp");
            Assert.True(result.IsError == true);
        }

        [Fact]
        public void EncryptionTools_AesDecrypt_InvalidPayload_ReturnsError()
        {
            var result = EncryptionTools.AesDecrypt("invalid_payload", "invalid_key");
            Assert.True(result.IsError == true);
        }

#if GAME_HACKING
        [Fact]
        public async Task EngineTools_ScanGameEngine_InvalidPID_ReturnsError()
        {
            var result = await EngineTools.ScanGameEngine(-9999);
            Assert.True(result.IsError == true);
        }
#endif

#if GAME_HACKING
        [Fact]
        public async Task InjectorTools_InjectNativeDll_InvalidPath_ReturnsError()
        {
            var result = await InjectorTools.InjectNativeDll(Process.GetCurrentProcess().Id, "C:\\invalid_dll.dll");
            Assert.True(result.IsError == true);
        }
#endif

        [Fact]
        public void InputHookTools_FreezeHardwareInput_InvalidDuration_ReturnsError()
        {
            var result = InputHookTools.FreezeHardwareInput(999); // max is 10
            Assert.True(result.IsError == true);
        }

#if GAME_HACKING
        [Fact]
        public async Task MemoryTools_ScanLiveMemory_InvalidPID_ReturnsError()
        {
            var result = await MemoryTools.ScanLiveMemoryInt32(-9999, 12345);
            Assert.True(result.IsError == true);
        }
#endif

        [Fact]
        public async Task NetworkTools_PingHost_InvalidHost_ReturnsError()
        {
            var result = await NetworkTools.PingHost("invalid.host.that.does.not.exist.local");
            Assert.True(result.IsError == true);
        }

        [Fact]
        public async Task NotificationTools_ShowNotification_ReturnsSuccess()
        {
            var result = await NotificationTools.ShowNotification("Test", "Test Notification");
            Assert.True(result.IsError != true);
        }

        [Fact]
        public void PcapTools_StartPacketCapture_InvalidDevice_ReturnsError()
        {
            var result = PcapTools.StartPacketCapture(9999);
            Assert.True(result.IsError == true);
        }

#if GAME_HACKING
        [Fact]
        public async Task PeTools_AnalyzeExecutable_InvalidPath_ReturnsError()
        {
            var result = await PeTools.AnalyzeExecutable("C:\\invalid_exe.exe");
            Assert.True(result.IsError == true);
        }
#endif

        [Fact]
        public async Task ResourceTools_GetHardwareMetrics_ReturnsSuccess()
        {
            var result = await ResourceTools.GetHardwareMetrics();
            Assert.True(result.IsError != true);
        }

        [Fact]
        public async Task SandboxTools_LaunchSandboxedProcess_InvalidPath_ReturnsError()
        {
            var result = await SandboxTools.LaunchSandboxedProcess("C:\\invalid_sandbox_exe.exe");
            Assert.True(result.IsError == true);
        }

        [Fact]
        public async Task WindowLayoutTools_SetWindowPosition_InvalidWindow_ReturnsError()
        {
            var result = await WindowLayoutTools.SetWindowPosition("NonExistentWindow123", 0, 0, 100, 100);
            Assert.True(result.IsError == true);
        }

        [Fact]
        public async Task WndProcTools_SendRawWindowMessage_InvalidWindow_ReturnsError()
        {
            var result = await WndProcTools.SendRawWindowMessage("NonExistentWindow123", 0x0100, 0, 0);
            Assert.True(result.IsError == true);
        }
    }
}
