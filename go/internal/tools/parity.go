
package tools

/**
 * @file parity.go
 * @module go/internal/tools
 *
 * WHAT: Go-native implementation of tool parity aliases (Claude Code, Codex, Gemini).
 * Ensures the Go sidecar can execute tools with identical signatures to major AI harnesses.
 *
 * WHY: Total Autonomy — if the Node control plane is down, the Go sidecar must be 
 * capable of fulfilling tool calls with the exact schemas models expect.
 */

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ToolResponse matches the MCP CallToolResult structure.
type ToolResponse struct {
	Content []TextContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func ok(text string) (ToolResponse, error) {
	return ToolResponse{
		Content: []TextContent{{Type: "text", Text: text}},
	}, nil
}

func err(text string) (ToolResponse, error) {
	return ToolResponse{
		Content: []TextContent{{Type: "text", Text: text}},
		IsError: true,
	}, nil
}

// ---------------------------------------------------------------------------
// Core Handler Implementations
// ---------------------------------------------------------------------------

func HandleRead(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ := getString(args, "file_path", "path", "AbsolutePath", "filePath")
	if path == "" {
		return err("file_path is required")
	}

	data, e := os.ReadFile(path)
	if e != nil {
		return err(fmt.Sprintf("Error reading file: %v", e))
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	startLine := getInt(args, "start_line", "startLine", "offset", "line_start")
	endLine := getInt(args, "end_line", "endLine", "limit", "line_end")

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
		sliced := lines[start:end]

		if getBool(args, "show_line_numbers", "showLineNumbers", "lineNumbers") {
			for i, line := range sliced {
				sliced[i] = fmt.Sprintf("%d: %s", start+i+1, line)
			}
		}
		return ok(strings.Join(sliced, "\n"))
	}

	if getBool(args, "show_line_numbers", "showLineNumbers") {
		for i, line := range lines {
			lines[i] = fmt.Sprintf("%d: %s", i+1, line)
		}
		return ok(strings.Join(lines, "\n"))
	}

	return ok(content)
}

func HandleGrep(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pattern, _ := getString(args, "pattern", "query", "regex")
	path, _ := getString(args, "dir_path", "path", "directory")
	if pattern == "" {
		return err("pattern is required")
	}
	if path == "" {
		path = "."
	}

	includePattern := getStringValue(args, "include_pattern")
	
	cmdArgs := []string{"--line-number", "--with-filename", "--no-heading"}
	if includePattern != "" {
		cmdArgs = append(cmdArgs, "-g", includePattern)
	}
	cmdArgs = append(cmdArgs, "--", pattern, path)

	cmd := exec.CommandContext(ctx, "rg", cmdArgs...)
	output, e := cmd.CombinedOutput()
	outputStr := string(output)

	if e != nil {
		// rg returns exit code 1 if no matches found
		if exitErr, okMatch := e.(*exec.ExitError); okMatch && exitErr.ExitCode() == 1 {
			return ok("No matches found.")
		}
		return err(fmt.Sprintf("Grep Error: %v\n%s", e, outputStr))
	}

	return ok(strings.TrimSpace(outputStr))
}

func HandleGlob(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pattern, _ := getString(args, "pattern", "glob")
	path, _ := getString(args, "path", "dir_path", "directory")
	if pattern == "" {
		return err("pattern is required")
	}
	if path == "" {
		path = "."
	}

	// Use ripgrep --files with glob if possible for performance
	cmd := exec.CommandContext(ctx, "rg", "--files", "-g", pattern, path)
	output, e := cmd.CombinedOutput()
	outputStr := string(output)

	if e == nil {
		return ok(strings.TrimSpace(outputStr))
	}

	// Fallback to manual walk if rg fails or not found
	var matches []string
	filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		
		rel, _ := filepath.Rel(path, p)
		matched, _ := filepath.Match(pattern, rel)
		if matched {
			matches = append(matches, rel)
		}
		return nil
	})

	if len(matches) == 0 {
		return ok("No files matched.")
	}

	return ok(strings.Join(matches, "\n"))
}

func HandleApplyPatch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ := getString(args, "file_path", "path")
	patchContent, _ := getString(args, "patch", "diff")

	if path == "" {
		return err("file_path is required")
	}
	if patchContent == "" {
		return err("patch is required")
	}

	tmpPatch := filepath.Join(os.TempDir(), fmt.Sprintf("patch-%d.diff", time.Now().UnixNano()))
	if e := os.WriteFile(tmpPatch, []byte(patchContent), 0644); e != nil {
		return err(fmt.Sprintf("Error writing temp patch: %v", e))
	}
	defer os.Remove(tmpPatch)

	cmd := exec.CommandContext(ctx, "patch", "-u", path, "-i", tmpPatch)
	output, e := cmd.CombinedOutput()
	if e != nil {
		return err(fmt.Sprintf("Patch Error: %v\n%s", e, string(output)))
	}

	return ok(fmt.Sprintf("Successfully applied patch to %s", path))
}

func HandleMultiEdit(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ := getString(args, "file_path", "path", "filePath")
	if path == "" {
		return err("file_path is required")
	}

	var editsRaw []interface{}
	if v, okEdits := args["edits"].([]interface{}); okEdits {
		editsRaw = v
	} else if v, okOps := args["operations"].([]interface{}); okOps {
		editsRaw = v
	}
	if len(editsRaw) == 0 {
		return err("edits array is required")
	}

	data, e := os.ReadFile(path)
	if e != nil {
		return err(fmt.Sprintf("Error reading file: %v", e))
	}

	content := string(data)
	results := []string{}

	for i, editRaw := range editsRaw {
		edit, ok := editRaw.(map[string]interface{})
		if !ok {
			continue
		}

		oldText, _ := getString(edit, "old_string", "oldText", "old_text", "find")
		newText, _ := getString(edit, "new_string", "newText", "new_text", "replace")

		if oldText == "" {
			return err(fmt.Sprintf("Edit %d: old_string is required", i+1))
		}

		if !strings.Contains(content, oldText) {
			return err(fmt.Sprintf("Edit %d: old_string not found in %s", i+1, path))
		}

		if getBool(edit, "replace_all", "replaceAll") {
			content = strings.ReplaceAll(content, oldText, newText)
		} else {
			content = strings.Replace(content, oldText, newText, 1)
		}
		results = append(results, fmt.Sprintf("Edit %d applied", i+1))
	}

	if e := os.WriteFile(path, []byte(content), 0644); e != nil {
		return err(fmt.Sprintf("Error writing file: %v", e))
	}

	return ok(fmt.Sprintf("Successfully applied %d edits to %s\n%s", len(results), path, strings.Join(results, "\n")))
}

func HandleWrite(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ := getString(args, "file_path", "path", "filePath", "AbsolutePath")
	content, _ := getString(args, "content", "text", "body")
	if path == "" {
		return err("file_path is required")
	}

	if errDir := os.MkdirAll(filepath.Dir(path), 0755); errDir != nil {
		return err(fmt.Sprintf("Error creating directory: %v", errDir))
	}

	if e := os.WriteFile(path, []byte(content), 0644); e != nil {
		return err(fmt.Sprintf("Error writing file: %v", e))
	}

	return ok(fmt.Sprintf("Successfully wrote to %s", path))
}

func HandleEdit(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ := getString(args, "file_path", "path", "filePath", "AbsolutePath")
	oldText, _ := getString(args, "old_string", "oldText", "old_text", "find")
	newText, _ := getString(args, "new_string", "newText", "new_text", "replace")

	if path == "" {
		return err("file_path is required")
	}
	if oldText == "" {
		return err("old_string is required")
	}

	data, e := os.ReadFile(path)
	if e != nil {
		return err(fmt.Sprintf("Error reading file: %v", e))
	}

	content := string(data)
	if !strings.Contains(content, oldText) {
		return err(fmt.Sprintf("old_string not found in %s. Ensure exact match including whitespace and indentation.", path))
	}

	matchCount := strings.Count(content, oldText)
	if matchCount > 1 && !getBool(args, "allow_multiple", "replace_all") {
		return err(fmt.Sprintf("old_string found %d times in %s. Provide more context to uniquely identify the location, or set allow_multiple=true.", matchCount, path))
	}

	var newContent string
	if getBool(args, "allow_multiple", "replace_all") {
		newContent = strings.ReplaceAll(content, oldText, newText)
	} else {
		newContent = strings.Replace(content, oldText, newText, 1)
	}

	if e := os.WriteFile(path, []byte(newContent), 0644); e != nil {
		return err(fmt.Sprintf("Error writing file: %v", e))
	}

	return ok(fmt.Sprintf("Successfully edited %s", path))
}

func HandleBash(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ := getString(args, "command", "cmd", "script")
	cwd, _ := getString(args, "cwd", "workingDir", "working_dir", "workdir")
	if command == "" {
		return err("command is required")
	}

	if cwd == "" {
		var errCwd error
		cwd, errCwd = os.Getwd()
		if errCwd != nil {
			cwd = "."
		}
	}

	timeoutMs := getInt(args, "timeout", "maxTime")
	if timeoutMs <= 0 {
		timeoutMs = 120000
	}

	tCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	// Use shell for command execution
	cmd := exec.CommandContext(tCtx, "sh", "-c", command)
	if os.Getenv("OS") == "Windows_NT" {
		cmd = exec.CommandContext(tCtx, "cmd", "/C", command)
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
		return err(fmt.Sprintf("%s\nExecution error: %v", outputStr, e))
	}

	return ok(strings.TrimSpace(outputStr))
}

func HandleLS(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ := getString(args, "path", "dir_path", "directory", "dir")
	if path == "" {
		path = "."
	}

	entries, e := os.ReadDir(path)
	if e != nil {
		return err(fmt.Sprintf("Error listing directory: %v", e))
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() != entries[j].IsDir() {
			return entries[i].IsDir()
		}
		return entries[i].Name() < entries[j].Name()
	})

	results := make([]map[string]string, 0, len(entries))
	plain := []string{}
	for _, entry := range entries {
		name := entry.Name()
		t := "file"
		if entry.IsDir() {
			name += "/"
			t = "directory"
		}
		results = append(results, map[string]string{
			"name": name,
			"type": t,
		})
		plain = append(plain, name)
	}

	if getStringValue(args, "format") == "plain" || getBool(args, "simple") {
		return ok(strings.Join(plain, "\n"))
	}

	jsonData, _ := json.MarshalIndent(results, "", "  ")
	return ok(string(jsonData))
}

func HandleWebFetch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ := getString(args, "url", "uri")
	if url == "" {
		return err("url is required")
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("Error creating request: %v", e))
	}
	req.Header.Set("User-Agent", "TormentNexus/ToolParity-Go")

	resp, e := client.Do(req)
	if e != nil {
		return err(fmt.Sprintf("Fetch error: %v", e))
	}
	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("Error reading response: %v", e))
	}

	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)))
	}

	content := string(body)
	if len(content) > 100000 {
		content = content[:100000] + "\n...[Response truncated]"
	}

	return ok(content)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func getString(args map[string]interface{}, keys ...string) (string, bool) {
	for _, k := range keys {
		if v, ok := args[k].(string); ok {
			return v, true
		}
	}
	return "", false
}

func getStringValue(args map[string]interface{}, key string) string {
	v, _ := args[key].(string)
	return v
}

func getInt(args map[string]interface{}, keys ...string) int {
	for _, k := range keys {
		if v, ok := args[k].(float64); ok {
			return int(v)
		}
		if v, ok := args[k].(int); ok {
			return v
		}
	}
	return 0
}

func getBool(args map[string]interface{}, keys ...string) bool {
	for _, k := range keys {
		if v, ok := args[k].(bool); ok {
			return v
		}
	}
	return false
}
