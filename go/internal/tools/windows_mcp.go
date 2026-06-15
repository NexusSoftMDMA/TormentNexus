//go:build ignore
// +build ignore

package tools

/**
 * @file windows_mcp.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of Windows MCP tools.
 * Replaces `windows-mcp` (uvx windows-mcp) entry in mcp.json.
 *
 * Provides Windows-specific system operations:
 * - Registry access, service management, process management
 * - Event log queries, system info, clipboard, window management
 * - Works cross-platform but with enhanced Windows support
 *
 * Improvements over original:
 * - No uvx/Python dependency.
 * - Go-native OS API calls.
 * - Graceful degradation on non-Windows platforms.
 */

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func isWindows() bool {
	return runtime.GOOS == "windows"
}

// HandleWindowsMCPGetSystemInfo returns detailed Windows system information.
// Tool: windows_get_system_info
func HandleWindowsMCPGetSystemInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	result := map[string]interface{}{
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
		"cpus":       runtime.NumCPU(),
		"hostname":   func() string { h, _ := os.Hostname(); return h }(),
		"go_version": runtime.Version(),
		"env_vars": map[string]string{
			"USERNAME":        os.Getenv("USERNAME"),
			"USERPROFILE":     os.Getenv("USERPROFILE"),
			"COMPUTERNAME":    os.Getenv("COMPUTERNAME"),
			"PROGRAMFILES":    os.Getenv("PROGRAMFILES"),
			"SYSTEMROOT":      os.Getenv("SYSTEMROOT"),
			"PATH":            truncateEnv(os.Getenv("PATH"), 500),
		},
	}

	if isWindows() {
		// Windows-specific system info via PowerShell
		if output, e := exec.CommandContext(ctx, "powershell", "-Command",
			"Get-CimInstance Win32_OperatingSystem | Select-Object Caption, Version, BuildNumber, OSArchitecture, TotalVisibleMemorySize, FreePhysicalMemory | ConvertTo-Json").Output(); e == nil {
			var winInfo interface{}
			if json.Unmarshal(output, &winInfo) == nil {
				result["windows_info"] = winInfo
			}
		}

		// Disk info
		if output, e := exec.CommandContext(ctx, "powershell", "-Command",
			"Get-CimInstance Win32_LogicalDisk | Select-Object DeviceID, Size, FreeSpace, FileSystem | ConvertTo-Json").Output(); e == nil {
			var diskInfo interface{}
			if json.Unmarshal(output, &diskInfo) == nil {
				result["disk_info"] = diskInfo
			}
		}
	} else {
		// Unix system info
		if output, e := exec.CommandContext(ctx, "uname", "-a").Output(); e == nil {
			result["unix_info"] = strings.TrimSpace(string(output))
		}
		if output, e := exec.CommandContext(ctx, "free", "-h").Output(); e == nil {
			result["memory_info"] = strings.TrimSpace(string(output))
		}
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleWindowsMCPListServices lists Windows services.
// Tool: windows_list_services
func HandleWindowsMCPListServices(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	if !isWindows() {
		// Fallback: list systemd services on Linux
		output, e := exec.CommandContext(ctx, "systemctl", "list-units", "--type=service", "--no-pager").Output()
		if e != nil {
			return ok("Service listing not available on this platform.")
		}
		return ok(string(output))
	}

	state, _ := getString(args, "state")
	psFilter := ""
	if state != "" {
		psFilter = fmt.Sprintf(" | Where-Object Status -eq '%s'", state)
	}

	output, e := exec.CommandContext(ctx, "powershell", "-Command",
		fmt.Sprintf("Get-Service%s | Select-Object Name, DisplayName, Status, StartType | ConvertTo-Json", psFilter)).Output()
	if e != nil {
		return err(fmt.Sprintf("Failed to list services: %v", e))
	}

	return ok(string(output))
}

// HandleWindowsMCPGetService gets details for a specific Windows service.
// Tool: windows_get_service
func HandleWindowsMCPGetService(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ := getString(args, "name", "service_name")
	if name == "" {
		return err("name parameter is required")
	}

	if !isWindows() {
		output, e := exec.CommandContext(ctx, "systemctl", "status", name).Output()
		if e != nil {
			return err(fmt.Sprintf("Service not found: %s", name))
		}
		return ok(string(output))
	}

	output, e := exec.CommandContext(ctx, "powershell", "-Command",
		fmt.Sprintf("Get-Service -Name '%s' | Select-Object Name, DisplayName, Status, StartType, CanPauseAndContinue, CanShutdown, CanStop | ConvertTo-Json", name)).Output()
	if e != nil {
		return err(fmt.Sprintf("Failed to get service: %v", e))
	}

	return ok(string(output))
}

// HandleWindowsMCPListProcesses lists running processes with details.
// Tool: windows_list_processes
func HandleWindowsMCPListProcesses(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	if isWindows() {
		output, e := exec.CommandContext(ctx, "powershell", "-Command",
			"Get-Process | Select-Object Id, ProcessName, CPU, WorkingSet, Path | ConvertTo-Json").Output()
		if e != nil {
			return err(fmt.Sprintf("Failed to list processes: %v", e))
		}
		return ok(string(output))
	}

	output, e := exec.CommandContext(ctx, "ps", "aux").Output()
	if e != nil {
		return err(fmt.Sprintf("Failed to list processes: %v", e))
	}
	return ok(string(output))
}

// HandleWindowsMCPReadRegistry reads a Windows registry value.
// Tool: windows_read_registry
func HandleWindowsMCPReadRegistry(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	if !isWindows() {
		return ok("Registry access is only available on Windows.")
	}

	key, _ := getString(args, "key", "path")
	if key == "" {
		return err("key parameter is required (e.g., 'HKLM:\\Software\\Microsoft\\Windows\\CurrentVersion')")
	}

	name, _ := getString(args, "name", "value_name")

	psCmd := fmt.Sprintf("Get-ItemProperty -Path '%s'", key)
	if name != "" {
		psCmd += fmt.Sprintf(" | Select-Object -ExpandProperty '%s'", name)
	} else {
		psCmd += " | ConvertTo-Json"
	}

	output, e := exec.CommandContext(ctx, "powershell", "-Command", psCmd).Output()
	if e != nil {
		return err(fmt.Sprintf("Failed to read registry: %v", e))
	}

	return ok(string(output))
}

// HandleWindowsMCPOpenApplication opens an application.
// Tool: windows_open_application
func HandleWindowsMCPOpenApplication(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	app, _ := getString(args, "application", "app", "name")
	if app == "" {
		return err("application parameter is required")
	}

	var cmd *exec.Cmd
	switch {
	case isWindows():
		cmd = exec.CommandContext(ctx, "cmd", "/C", "start", "", app)
	case runtime.GOOS == "darwin":
		cmd = exec.CommandContext(ctx, "open", "-a", app)
	default:
		cmd = exec.CommandContext(ctx, "sh", "-c", app)
	}

	if e := cmd.Start(); e != nil {
		return err(fmt.Sprintf("Failed to open application: %v", e))
	}

	return ok(fmt.Sprintf("Application launched: %s", app))
}

// HandleWindowsMCPGetClipboard gets clipboard content.
// Tool: windows_get_clipboard
func HandleWindowsMCPGetClipboard(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	if isWindows() {
		output, e := exec.CommandContext(ctx, "powershell", "-Command", "Get-Clipboard").Output()
		if e != nil {
			return err(fmt.Sprintf("Failed to get clipboard: %v", e))
		}
		return ok(string(output))
	}

	// Try xclip on Linux
	if output, e := exec.CommandContext(ctx, "xclip", "-selection", "clipboard", "-o").Output(); e == nil {
		return ok(string(output))
	}

	// Try pbpaste on macOS
	if output, e := exec.CommandContext(ctx, "pbpaste").Output(); e == nil {
		return ok(string(output))
	}

	return ok("Clipboard access not available on this platform.")
}

// HandleWindowsMCPSetClipboard sets clipboard content.
// Tool: windows_set_clipboard
func HandleWindowsMCPSetClipboard(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	content, _ := getString(args, "content", "text")
	if content == "" {
		return err("content parameter is required")
	}

	if isWindows() {
		cmd := exec.CommandContext(ctx, "powershell", "-Command", "Set-Clipboard")
		cmd.Stdin = strings.NewReader(content)
		if e := cmd.Run(); e != nil {
			return err(fmt.Sprintf("Failed to set clipboard: %v", e))
		}
		return ok("Clipboard content set.")
	}

	return ok("Clipboard write not available on this platform.")
}

// HandleWindowsMCPListDrives lists available drives.
// Tool: windows_list_drives
func HandleWindowsMCPListDrives(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	if isWindows() {
		output, e := exec.CommandContext(ctx, "powershell", "-Command",
			"Get-PSDrive -PSProvider FileSystem | Select-Object Name, Root, Used, Free, Description | ConvertTo-Json").Output()
		if e != nil {
			return err(fmt.Sprintf("Failed to list drives: %v", e))
		}
		return ok(string(output))
	}

	// Unix: list mount points
	output, e := exec.CommandContext(ctx, "df", "-h").Output()
	if e != nil {
		return err(fmt.Sprintf("Failed to list drives: %v", e))
	}
	return ok(string(output))
}

// HandleWindowsMCPGetEventLog queries the Windows Event Log.
// Tool: windows_get_event_log
func HandleWindowsMCPGetEventLog(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	if !isWindows() {
		return ok("Event log access is only available on Windows.")
	}

	logName, _ := getString(args, "log_name", "source")
	if logName == "" {
		logName = "Application"
	}

	maxEntries := getInt(args, "max_entries", "limit")
	if maxEntries <= 0 {
		maxEntries = 20
	}

	psCmd := fmt.Sprintf(
		"Get-EventLog -LogName '%s' -Newest %d | Select-Object TimeGenerated, EntryType, Source, Message | ConvertTo-Json",
		logName, maxEntries)

	output, e := exec.CommandContext(ctx, "powershell", "-Command", psCmd).Output()
	if e != nil {
		return err(fmt.Sprintf("Failed to read event log: %v", e))
	}

	return ok(string(output))
}

func truncateEnv(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// Unused import guard
var _ = filepath.Base
