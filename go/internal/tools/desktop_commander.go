//go:build ignore
// +build ignore

package tools

/**
 * @file desktop_commander.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of Desktop Commander MCP tools.
 * Replaces `desktop-commander` (npx @wonderwhy-er/desktop-commander@latest) entry in mcp.json.
 *
 * Desktop Commander provides system-level operations:
 * process management, file operations, command execution, and system monitoring.
 *
 * Improvements over original:
 * - No npx/Node dependency.
 * - Go-native process management with proper signal handling.
 * - Supports: execute_command, read_file, write_file, list_directory,
 *   search_files, get_system_info, process_list, process_kill, block_command.
 * - Context-aware with timeout; cross-platform support.
 */

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

// HandleDesktopExecuteCommand executes a shell command and returns output.
// Tool: desktop_execute_command
func HandleDesktopExecuteCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ := getString(args, "command", "cmd")
	if command == "" {
		return err("command parameter is required")
	}

	cwd, _ := getString(args, "cwd", "working_dir")
	if cwd == "" {
		cwd, _ = os.Getwd()
	}

	timeoutMs := getInt(args, "timeout", "maxTime")
	if timeoutMs <= 0 {
		timeoutMs = 30000
	}

	tCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(tCtx, "cmd", "/C", command)
	} else {
		cmd = exec.CommandContext(tCtx, "sh", "-c", command)
	}
	cmd.Dir = cwd

	output, e := cmd.CombinedOutput()
	outputStr := string(output)

	if len(outputStr) > 50000 {
		outputStr = outputStr[:50000] + "\n...[Output truncated]"
	}

	if e != nil {
		if tCtx.Err() == context.DeadlineExceeded {
			return err(fmt.Sprintf("Command timed out after %dms", timeoutMs))
		}
		// Return output even on error — many commands return useful stderr
		return ToolResponse{
			Content: []TextContent{{
				Type: "text",
				Text: fmt.Sprintf("Command: %s\nExit code: non-zero\nOutput:\n%s\nError: %v", command, outputStr, e),
			}},
			IsError: false, // Desktop Commander returns output even on non-zero exit
		}, nil
	}

	return ok(outputStr)
}

// HandleDesktopReadFile reads a file with optional line range.
// Tool: desktop_read_file
func HandleDesktopReadFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ := getString(args, "path", "file_path")
	if path == "" {
		return err("path parameter is required")
	}

	data, e := os.ReadFile(path)
	if e != nil {
		return err(fmt.Sprintf("Failed to read file: %v", e))
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	startLine := getInt(args, "start_line", "startLine")
	endLine := getInt(args, "end_line", "endLine")

	if startLine > 0 || endLine > 0 {
		start := startLine - 1
		if start < 0 {
			start = 0
		}
		end := endLine
		if end <= 0 || end > len(lines) {
			end = len(lines)
		}
		if start > len(lines) {
			return ok("")
		}
		return ok(strings.Join(lines[start:end], "\n"))
	}

	return ok(content)
}

// HandleDesktopReadMultipleFiles reads multiple files at once.
// Tool: desktop_read_multiple_files
func HandleDesktopReadMultipleFiles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	paths, pathsOK := args["paths"].([]interface{})
	if !pathsOK || len(paths) == 0 {
		return err("paths parameter (array of file paths) is required")
	}

	var results []string
	for _, p := range paths {
		pathStr, okStr := p.(string)
		if !okStr {
			continue
		}
		data, e := os.ReadFile(pathStr)
		if e != nil {
			results = append(results, fmt.Sprintf("--- %s ---\n[Error: %v]", pathStr, e))
			continue
		}
		results = append(results, fmt.Sprintf("--- %s ---\n%s", pathStr, string(data)))
	}

	return ok(strings.Join(results, "\n\n"))
}

// HandleDesktopWriteFile writes content to a file.
// Tool: desktop_write_file
func HandleDesktopWriteFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ := getString(args, "path", "file_path")
	content, _ := getString(args, "content", "text")
	if path == "" {
		return err("path parameter is required")
	}

	if errDir := os.MkdirAll(filepath.Dir(path), 0755); errDir != nil {
		return err(fmt.Sprintf("Failed to create directory: %v", errDir))
	}

	if e := os.WriteFile(path, []byte(content), 0644); e != nil {
		return err(fmt.Sprintf("Failed to write file: %v", e))
	}

	return ok(fmt.Sprintf("Successfully wrote to %s", path))
}

// HandleDesktopCreateDirectory creates a directory recursively.
// Tool: desktop_create_directory
func HandleDesktopCreateDirectory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ := getString(args, "path")
	if path == "" {
		return err("path parameter is required")
	}

	if e := os.MkdirAll(path, 0755); e != nil {
		return err(fmt.Sprintf("Failed to create directory: %v", e))
	}

	return ok(fmt.Sprintf("Directory created: %s", path))
}

// HandleDesktopListDirectory lists directory contents with details.
// Tool: desktop_list_directory
func HandleDesktopListDirectory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ := getString(args, "path", "dir_path")
	if path == "" {
		path = "."
	}

	entries, e := os.ReadDir(path)
	if e != nil {
		return err(fmt.Sprintf("Failed to list directory: %v", e))
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() != entries[j].IsDir() {
			return entries[i].IsDir()
		}
		return entries[i].Name() < entries[j].Name()
	})

	var results []string
	for _, entry := range entries {
		info, _ := entry.Info()
		name := entry.Name()
		suffix := ""
		if entry.IsDir() {
			suffix = "/"
		}

		sizeStr := ""
		if info != nil && !entry.IsDir() {
			sizeStr = fmt.Sprintf(" (%d bytes)", info.Size())
		}

		results = append(results, name+suffix+sizeStr)
	}

	return ok(strings.Join(results, "\n"))
}

// HandleDesktopDirectoryTree returns a recursive tree of a directory.
// Tool: desktop_directory_tree
func HandleDesktopDirectoryTree(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ := getString(args, "path")
	if path == "" {
		path = "."
	}

	var lines []string
	errWalk := filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		rel, errRel := filepath.Rel(path, p)
		if errRel != nil {
			return errRel
		}
		if rel == "." {
			return nil
		}

		depth := len(strings.Split(rel, string(filepath.Separator)))
		indent := strings.Repeat("  ", depth-1)
		name := d.Name()
		if d.IsDir() {
			name += "/"
		}
		lines = append(lines, indent+name)
		return nil
	})

	if errWalk != nil {
		return err(fmt.Sprintf("Failed to walk directory: %v", errWalk))
	}

	return ok(strings.Join(lines, "\n"))
}

// HandleDesktopSearchFiles searches for files matching a pattern.
// Tool: desktop_search_files
func HandleDesktopSearchFiles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ := getString(args, "path")
	if path == "" {
		path = "."
	}

	pattern, _ := getString(args, "pattern")
	if pattern == "" {
		pattern = "*"
	}

	excludePatterns := []string{}
	if excludeVal, exists := args["excludePatterns"].([]interface{}); exists {
		for _, item := range excludeVal {
			if s, okS := item.(string); okS {
				excludePatterns = append(excludePatterns, s)
			}
		}
	}

	var matches []string
	filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(path, p)

		// Check exclude patterns
		for _, ep := range excludePatterns {
			if matched, _ := filepath.Match(ep, d.Name()); matched {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		matched, _ := filepath.Match(pattern, d.Name())
		if matched && !d.IsDir() {
			matches = append(matches, rel)
		}
		return nil
	})

	if len(matches) == 0 {
		return ok("No files matched the pattern.")
	}

	return ok(strings.Join(matches, "\n"))
}

// HandleDesktopMoveFile moves or renames a file or directory.
// Tool: desktop_move_file
func HandleDesktopMoveFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	source, _ := getString(args, "source")
	destination, _ := getString(args, "destination")
	if source == "" || destination == "" {
		return err("source and destination parameters are required")
	}

	if e := os.MkdirAll(filepath.Dir(destination), 0755); e != nil {
		return err(fmt.Sprintf("Failed to create destination directory: %v", e))
	}

	if e := os.Rename(source, destination); e != nil {
		return err(fmt.Sprintf("Failed to move: %v", e))
	}

	return ok(fmt.Sprintf("Moved %s to %s", source, destination))
}

// HandleDesktopGetFileInfo returns metadata about a file.
// Tool: desktop_get_file_info
func HandleDesktopGetFileInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ := getString(args, "path")
	if path == "" {
		return err("path parameter is required")
	}

	info, e := os.Stat(path)
	if e != nil {
		return err(fmt.Sprintf("Failed to get file info: %v", e))
	}

	result := fmt.Sprintf("Name: %s\nSize: %d bytes\nIsDir: %v\nMode: %s\nModified: %s",
		info.Name(), info.Size(), info.IsDir(), info.Mode().String(), info.ModTime().Format(time.RFC3339))

	return ok(result)
}

// HandleDesktopListProcesses lists running processes.
// Tool: desktop_list_processes
func HandleDesktopListProcesses(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// Cross-platform process listing
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "tasklist", "/FO", "CSV")
	} else {
		cmd = exec.CommandContext(ctx, "ps", "aux")
	}

	output, e := cmd.CombinedOutput()
	if e != nil {
		return err(fmt.Sprintf("Failed to list processes: %v", e))
	}

	result := string(output)
	if len(result) > 20000 {
		result = result[:20000] + "\n...[Output truncated]"
	}

	return ok(result)
}

// HandleDesktopKillProcess kills a process by PID.
// Tool: desktop_kill_process
func HandleDesktopKillProcess(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pid := getInt(args, "pid")
	if pid <= 0 {
		return err("pid parameter is required")
	}

	sig, _ := getString(args, "signal")
	if sig == "" {
		sig = "SIGTERM"
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "taskkill", "/PID", fmt.Sprintf("%d", pid), "/F")
	} else {
		cmd = exec.CommandContext(ctx, "kill", "-"+sig, fmt.Sprintf("%d", pid))
	}

	output, e := cmd.CombinedOutput()
	if e != nil {
		return err(fmt.Sprintf("Failed to kill process %d: %v\n%s", pid, e, string(output)))
	}

	return ok(fmt.Sprintf("Process %d killed (signal: %s)", pid, sig))
}

// HandleDesktopGetSystemInfo returns system information.
// Tool: desktop_get_system_info
func HandleDesktopGetSystemInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	hostname, _ := os.Hostname()
	result := fmt.Sprintf("OS: %s\nArchitecture: %s\nHostname: %s\nCPUs: %d\nGo Version: %s\nCompiler: %s",
		runtime.GOOS, runtime.GOARCH, hostname, runtime.NumCPU(), runtime.Version(), runtime.Compiler)

	// Try to get memory info
	if runtime.GOOS != "windows" {
		if output, e := exec.CommandContext(ctx, "free", "-h").Output(); e == nil {
			result += "\n\nMemory:\n" + string(output)
		}
		if output, e := exec.CommandContext(ctx, "df", "-h", "/").Output(); e == nil {
			result += "\nDisk:\n" + string(output)
		}
	}

	return ok(result)
}

// HandleDesktopExecuteScript executes a script (Node.js, Python, etc).
// Tool: desktop_execute_script
func HandleDesktopExecuteScript(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ := getString(args, "code", "script")
	if code == "" {
		return err("code parameter is required")
	}

	language, _ := getString(args, "language", "lang", "type")
	if language == "" {
		language = "node"
	}

	timeoutMs := getInt(args, "timeout")
	if timeoutMs <= 0 {
		timeoutMs = 30000
	}

	tCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	// Write code to temp file
	ext := ".js"
	runner := "node"
	switch language {
	case "python", "python3":
		ext = ".py"
		runner = "python3"
	case "bash", "sh", "shell":
		ext = ".sh"
		runner = "bash"
	case "ruby", "rb":
		ext = ".rb"
		runner = "ruby"
	case "perl":
		ext = ".pl"
		runner = "perl"
	default:
		ext = ".js"
		runner = "node"
	}

	tmpFile, e := os.CreateTemp("", "desktop-cmd-*"+ext)
	if e != nil {
		return err(fmt.Sprintf("Failed to create temp file: %v", e))
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, e := tmpFile.WriteString(code); e != nil {
		tmpFile.Close()
		return err(fmt.Sprintf("Failed to write script: %v", e))
	}
	tmpFile.Close()

	cmd := exec.CommandContext(tCtx, runner, tmpPath)
	output, e := cmd.CombinedOutput()
	outputStr := string(output)

	if len(outputStr) > 50000 {
		outputStr = outputStr[:50000] + "\n...[Output truncated]"
	}

	if e != nil {
		if tCtx.Err() == context.DeadlineExceeded {
			return err(fmt.Sprintf("Script timed out after %dms", timeoutMs))
		}
		return err(fmt.Sprintf("Script error: %v\nOutput: %s", e, outputStr))
	}

	return ok(outputStr)
}

// HandleDesktopOpenFile opens a file with the system default application.
// Tool: desktop_open_file
func HandleDesktopOpenFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ := getString(args, "path", "file_path")
	if path == "" {
		return err("path parameter is required")
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.CommandContext(ctx, "open", path)
	case "windows":
		cmd = exec.CommandContext(ctx, "cmd", "/C", "start", "", path)
	default:
		cmd = exec.CommandContext(ctx, "xdg-open", path)
	}

	if e := cmd.Start(); e != nil {
		return err(fmt.Sprintf("Failed to open file: %v", e))
	}

	return ok(fmt.Sprintf("Opened %s with default application", path))
}

// HandleDesktopTailFile tails a file (reads the last N lines).
// Tool: desktop_tail_file
func HandleDesktopTailFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ := getString(args, "path", "file_path")
	if path == "" {
		return err("path parameter is required")
	}

	lines := getInt(args, "lines", "n")
	if lines <= 0 {
		lines = 50
	}

	// Use system tail if available
	if _, e := exec.LookPath("tail"); e == nil {
		cmd := exec.CommandContext(ctx, "tail", "-n", fmt.Sprintf("%d", lines), path)
		output, e := cmd.CombinedOutput()
		if e == nil {
			return ok(string(output))
		}
	}

	// Fallback: read file and return last N lines
	data, e := os.ReadFile(path)
	if e != nil {
		return err(fmt.Sprintf("Failed to read file: %v", e))
	}

	allLines := strings.Split(string(data), "\n")
	start := len(allLines) - lines
	if start < 0 {
		start = 0
	}

	return ok(strings.Join(allLines[start:], "\n"))
}

// Helper to get io.Reader from string
func stringReader(s string) io.Reader {
	return strings.NewReader(s)
}
