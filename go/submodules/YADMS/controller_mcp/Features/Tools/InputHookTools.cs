using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Diagnostics;
using System.Runtime.InteropServices;
using System.Threading;
using System.Threading.Tasks;
using ModelContextProtocol;
using ModelContextProtocol.Protocol;
using ModelContextProtocol.Server;

namespace controller_mcp.Features.Tools
{
    public static class InputHookTools
    {
        private const int WH_KEYBOARD_LL = 13;
        private const int WH_MOUSE_LL = 14;

        [DllImport("user32.dll", CharSet = CharSet.Auto, SetLastError = true)]
        private static extern IntPtr SetWindowsHookEx(int idHook, LowLevelProc lpfn, IntPtr hMod, uint dwThreadId);

        [DllImport("user32.dll", CharSet = CharSet.Auto, SetLastError = true)]
        [return: MarshalAs(UnmanagedType.Bool)]
        private static extern bool UnhookWindowsHookEx(IntPtr hhk);

        [DllImport("user32.dll", CharSet = CharSet.Auto, SetLastError = true)]
        private static extern IntPtr CallNextHookEx(IntPtr hhk, int nCode, IntPtr wParam, IntPtr lParam);

        [DllImport("kernel32.dll", CharSet = CharSet.Auto, SetLastError = true)]
        private static extern IntPtr GetModuleHandle(string lpModuleName);

        private delegate IntPtr LowLevelProc(int nCode, IntPtr wParam, IntPtr lParam);

        private static IntPtr _mouseHookID = IntPtr.Zero;
        private static IntPtr _keyboardHookID = IntPtr.Zero;
        private static LowLevelProc _mouseProc = HookCallback;
        private static LowLevelProc _keyboardProc = HookCallback;

        private static bool _isFrozen = false;
        private static CancellationTokenSource _freezeCts;

        static InputHookTools()
        {
            // Failsafe 1: Automatically unhook if the server crashes or exits
            AppDomain.CurrentDomain.ProcessExit += (s, e) => UnfreezeHardware();
        }

        private static IntPtr SetHook(int hookId, LowLevelProc proc)
        {
            using (Process curProcess = Process.GetCurrentProcess())
            using (ProcessModule curModule = curProcess.MainModule)
            {
                return SetWindowsHookEx(hookId, proc, GetModuleHandle(curModule.ModuleName), 0);
            }
        }

        private static IntPtr HookCallback(int nCode, IntPtr wParam, IntPtr lParam)
        {
            if (nCode >= 0 && _isFrozen)
            {
                // Return 1 to swallow the event (freeze the input)
                return (IntPtr)1;
            }
            return CallNextHookEx(IntPtr.Zero, nCode, wParam, lParam);
        }

        private static void UnfreezeHardware()
        {
            _isFrozen = false;
            if (_mouseHookID != IntPtr.Zero)
            {
                UnhookWindowsHookEx(_mouseHookID);
                _mouseHookID = IntPtr.Zero;
            }
            if (_keyboardHookID != IntPtr.Zero)
            {
                UnhookWindowsHookEx(_keyboardHookID);
                _keyboardHookID = IntPtr.Zero;
            }
        }

        [McpServerTool, Description("Temporarily freezes the physical mouse and keyboard of the host machine by intercepting all hardware interrupts. Includes a hard-coded maximum failsafe of 10 seconds to prevent permanent lockouts.")]
        public static CallToolResult FreezeHardwareInput(int duration_seconds)
        {
            try
            {
                if (_isFrozen)
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = "Hardware is already frozen." } } };

                // Failsafe 2: Cap the duration strictly to 10 seconds
                if (duration_seconds <= 0 || duration_seconds > 10)
                    return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = "Duration must be between 1 and 10 seconds for safety reasons." } } };

                _isFrozen = true;
                _mouseHookID = SetHook(WH_MOUSE_LL, _mouseProc);
                _keyboardHookID = SetHook(WH_KEYBOARD_LL, _keyboardProc);

                // Failsafe 3: Background watchdog thread to ensure it unfreezes even if the main thread hangs
                _freezeCts?.Cancel();
                _freezeCts = new CancellationTokenSource();
                Task.Run(async () =>
                {
                    try
                    {
                        await Task.Delay(duration_seconds * 1000, _freezeCts.Token);
                    }
                    catch { }
                    finally
                    {
                        UnfreezeHardware();
                    }
                });

                return new CallToolResult { Content = new List<ContentBlock> { new TextContentBlock { Text = $"Hardware successfully frozen for {duration_seconds} seconds. All physical inputs will be discarded." } } };
            }
            catch (Exception ex)
            {
                UnfreezeHardware();
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Failed to freeze hardware: {ex.Message}" } } };
            }
        }

        [McpServerTool, Description("Manually unfreezes the hardware input if it was frozen. This is usually handled automatically by the timeout watchdog.")]
        public static CallToolResult UnfreezeHardwareInput()
        {
            try
            {
                _freezeCts?.Cancel();
                UnfreezeHardware();
                return new CallToolResult { Content = new List<ContentBlock> { new TextContentBlock { Text = "Hardware input successfully unfrozen." } } };
            }
            catch (Exception ex)
            {
                return new CallToolResult { IsError = true, Content = new List<ContentBlock> { new TextContentBlock { Text = $"Failed to unfreeze hardware: {ex.Message}" } } };
            }
        }
    }
}
