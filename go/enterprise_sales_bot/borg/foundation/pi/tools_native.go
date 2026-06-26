package pi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sergi/go-diff/diffmatchpatch"
)

type ToolHandler func(ctx context.Context, cwd string, input json.RawMessage) (*ToolResult, error)

func DefaultToolHandlers() map[string]ToolHandler {
	return map[string]ToolHandler{
		"read":  executeReadTool,
		"write": executeWriteTool,
		"edit":  executeEditTool,
		"bash":  executeBashTool,
		"grep":  executeGrepTool,
		"find":  executeFindTool,
		"ls":    executeLsTool,
	}
}

func executeReadTool(ctx context.Context, cwd string, raw json.RawMessage) (*ToolResult, error) {
	var input ReadToolInput
	if err := json.Unmarshal(raw, &input); err != nil {
		return nil, fmt.Errorf("invalid read input: %w", err)
	}
	if strings.TrimSpace(input.Path) == "" {
		return nil, fmt.Errorf("path is required")
	}
	absolutePath, err := resolvePath(cwd, input.Path)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(absolutePath)
	if err != nil {
		return nil, err
	}
	mimeType := mime.TypeByExtension(strings.ToLower(filepath.Ext(absolutePath)))
	if strings.HasPrefix(mimeType, "image/") {
		return &ToolResult{
			ToolName: "read",
			Content: []any{
				TextContent{Type: "text", Text: fmt.Sprintf("Read image file [%s]", mimeType)},
				ImageContent{Type: "image", Data: base64.StdEncoding.EncodeToString(data), MimeType: mimeType},
			},
		}, nil
	}

	lines := strings.Split(string(data), "\n")
	start := 0
	if input.Offset > 0 {
		start = input.Offset - 1
	}
	if start >= len(lines) {
		return nil, fmt.Errorf("offset %d is beyond end of file (%d lines total)", input.Offset, len(lines))
	}
	selected := lines[start:]
	if input.Limit > 0 && input.Limit < len(selected) {
		selected = selected[:input.Limit]
	}
	selectedText := strings.Join(selected, "\n")
	truncation, text := truncateHead(selectedText)
	output := text
	if truncation.FirstLineExceeds {
		lineNo := start + 1
		output = fmt.Sprintf("[Line %d exceeds %d byte limit. Use bash to inspect it safely.]", lineNo, DefaultMaxBytes)
	} else if truncation.Truncated {
		endLine := start + truncation.OutputLines
		output += fmt.Sprintf("\n\n[Showing lines %d-%d of %d. Use offset=%d to continue.]", start+1, endLine, len(lines), endLine+1)
	} else if input.Limit > 0 && start+len(selected) < len(lines) {
		output += fmt.Sprintf("\n\n[%d more lines in file. Use offset=%d to continue.]", len(lines)-(start+len(selected)), start+len(selected)+1)
	}
	var details *ReadToolDetails
	if truncation.Truncated || truncation.FirstLineExceeds {
		nextOffset := 0
		if truncation.OutputLines > 0 {
			nextOffset = start + truncation.OutputLines + 1
		}
		truncation.ContinuationOffset = nextOffset
		details = &ReadToolDetails{Truncation: &truncation}
	}
	result := &ToolResult{ToolName: "read", Content: []any{TextContent{Type: "text", Text: output}}}
	if details != nil {
		result.Details = details
	}
	return result, nil
}

func executeWriteTool(ctx context.Context, cwd string, raw json.RawMessage) (*ToolResult, error) {
	var input WriteToolInput
	if err := json.Unmarshal(raw, &input); err != nil {
		return nil, fmt.Errorf("invalid write input: %w", err)
	}
	if strings.TrimSpace(input.Path) == "" {
		return nil, fmt.Errorf("path is required")
	}
	absolutePath, err := resolvePath(cwd, input.Path)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(absolutePath, []byte(input.Content), 0o644); err != nil {
		return nil, err
	}
	return &ToolResult{
		ToolName: "write",
		Content:  []any{TextContent{Type: "text", Text: fmt.Sprintf("Successfully wrote %d bytes to %s", len(input.Content), input.Path)}},
	}, nil
}

func executeEditTool(ctx context.Context, cwd string, raw json.RawMessage) (*ToolResult, error) {
	var input EditToolInput
	if err := json.Unmarshal(raw, &input); err != nil {
		return nil, fmt.Errorf("invalid edit input: %w", err)
	}
	if strings.TrimSpace(input.Path) == "" {
		return nil, fmt.Errorf("path is required")
	}
	if len(input.Edits) == 0 {
		return nil, fmt.Errorf("edit tool input is invalid. edits must contain at least one replacement")
	}
	absolutePath, err := resolvePath(cwd, input.Path)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(absolutePath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %s", input.Path)
	}
	original := string(data)
	type replacement struct {
		start   int
		end     int
		newText string
	}
	replacements := make([]replacement, 0, len(input.Edits))
	for _, edit := range input.Edits {
		if edit.OldText == "" {
			return nil, fmt.Errorf("oldText must not be empty")
		}
		matches := findAllOccurrences(original, edit.OldText)
		if len(matches) == 0 {
			return nil, fmt.Errorf("oldText not found in %s", input.Path)
		}
		if len(matches) > 1 {
			return nil, fmt.Errorf("oldText must match a unique region in %s", input.Path)
		}
		replacements = append(replacements, replacement{start: matches[0], end: matches[0] + len(edit.OldText), newText: edit.NewText})
	}
	sort.Slice(replacements, func(i, j int) bool { return replacements[i].start < replacements[j].start })
	for i := 1; i < len(replacements); i++ {
		if replacements[i].start < replacements[i-1].end {
			return nil, fmt.Errorf("edits overlap or are nested in %s", input.Path)
		}
	}
	var out strings.Builder
	cursor := 0
	firstChangedLine := 1
	if len(replacements) > 0 {
		firstChangedLine = 1 + strings.Count(original[:replacements[0].start], "\n")
	}
	for _, replacement := range replacements {
		out.WriteString(original[cursor:replacement.start])
		out.WriteString(replacement.newText)
		cursor = replacement.end
	}
	out.WriteString(original[cursor:])
	updated := out.String()
	if err := os.WriteFile(absolutePath, []byte(updated), 0o644); err != nil {
		return nil, err
	}
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(original, updated, false)
	return &ToolResult{
		ToolName: "edit",
		Content:  []any{TextContent{Type: "text", Text: fmt.Sprintf("Successfully replaced %d block(s) in %s.", len(input.Edits), input.Path)}},
		Details: &EditToolDetails{
			Diff:             dmp.DiffPrettyText(diffs),
			FirstChangedLine: firstChangedLine,
		},
	}, nil
}

func executeBashTool(ctx context.Context, cwd string, raw json.RawMessage) (*ToolResult, error) {
	var input BashToolInput
	if err := json.Unmarshal(raw, &input); err != nil {
		return nil, fmt.Errorf("invalid bash input: %w", err)
	}
	if strings.TrimSpace(input.Command) == "" {
		return nil, fmt.Errorf("command is required")
	}
	commandCtx := ctx
	var cancel context.CancelFunc
	if input.Timeout > 0 {
		commandCtx, cancel = context.WithTimeout(ctx, time.Duration(input.Timeout*float64(time.Second)))
		defer cancel()
	}
	cmd := shellCommand(commandCtx, input.Command)
	cmd.Dir = cwd
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	combined := stdout.String() + stderr.String()
	truncation, rendered := truncateTail(combined)
	var fullOutputPath string
	var details *BashToolDetails
	if truncation.Truncated {
		fullOutputPath = filepath.Join(os.TempDir(), fmt.Sprintf("pi-bash-%s.log", uuid.NewString()))
		if writeErr := os.WriteFile(fullOutputPath, []byte(combined), 0o600); writeErr == nil {
			truncation.FullOutputPath = fullOutputPath
			rendered += fmt.Sprintf("\n\n[Showing lines %d-%d of %d. Full output: %s]", truncation.TotalLines-truncation.OutputLines+1, truncation.TotalLines, truncation.TotalLines, fullOutputPath)
		}
		details = &BashToolDetails{Truncation: &truncation, FullOutputPath: fullOutputPath}
	}
	if rendered == "" {
		rendered = "(no output)"
	}
	if err != nil {
		if commandCtx.Err() == context.DeadlineExceeded {
			rendered += fmt.Sprintf("\n\nCommand timed out after %g seconds", input.Timeout)
		} else if exitErr, ok := err.(*exec.ExitError); ok {
			rendered += fmt.Sprintf("\n\nCommand exited with code %d", exitErr.ExitCode())
		}
		result := &ToolResult{ToolName: "bash", Content: []any{TextContent{Type: "text", Text: rendered}}, IsError: true}
		if details != nil {
			result.Details = details
		}
		return result, fmt.Errorf("%s", rendered)
	}
	result := &ToolResult{ToolName: "bash", Content: []any{TextContent{Type: "text", Text: rendered}}}
	if details != nil {
		result.Details = details
	}
	return result, nil
}

func resolvePath(cwd, toolPath string) (string, error) {
	if strings.TrimSpace(toolPath) == "" {
		return "", fmt.Errorf("path is required")
	}
	if filepath.IsAbs(toolPath) {
		return filepath.Clean(toolPath), nil
	}
	return filepath.Join(cwd, filepath.Clean(toolPath)), nil
}

func findAllOccurrences(haystack, needle string) []int {
	var matches []int
	for start := 0; start <= len(haystack)-len(needle); {
		idx := strings.Index(haystack[start:], needle)
		if idx < 0 {
			break
		}
		absolute := start + idx
		matches = append(matches, absolute)
		start = absolute + len(needle)
	}
	return matches
}

func truncateHead(text string) (TruncationDetails, string) {
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return TruncationDetails{}, ""
	}
	if len(lines[0]) > DefaultMaxBytes {
		return TruncationDetails{Truncated: true, FirstLineExceeds: true, MaxBytes: DefaultMaxBytes, TotalLines: len(lines)}, ""
	}
	var out []string
	bytesUsed := 0
	for i, line := range lines {
		lineBytes := len(line)
		if i < len(lines)-1 {
			lineBytes++
		}
		if len(out) >= DefaultMaxLines {
			return TruncationDetails{Truncated: true, TruncatedBy: "lines", TotalLines: len(lines), OutputLines: len(out), OutputBytes: bytesUsed, MaxLines: DefaultMaxLines, MaxBytes: DefaultMaxBytes}, strings.Join(out, "\n")
		}
		if bytesUsed+lineBytes > DefaultMaxBytes {
			return TruncationDetails{Truncated: true, TruncatedBy: "bytes", TotalLines: len(lines), OutputLines: len(out), OutputBytes: bytesUsed, MaxLines: DefaultMaxLines, MaxBytes: DefaultMaxBytes}, strings.Join(out, "\n")
		}
		out = append(out, line)
		bytesUsed += lineBytes
	}
	return TruncationDetails{TotalLines: len(lines), OutputLines: len(out), OutputBytes: bytesUsed, MaxLines: DefaultMaxLines, MaxBytes: DefaultMaxBytes}, strings.Join(out, "\n")
}

func truncateTail(text string) (TruncationDetails, string) {
	lines := strings.Split(text, "\n")
	if len(lines) <= DefaultMaxLines && len(text) <= DefaultMaxBytes {
		return TruncationDetails{TotalLines: len(lines), OutputLines: len(lines), OutputBytes: len(text), MaxLines: DefaultMaxLines, MaxBytes: DefaultMaxBytes}, text
	}
	selected := make([]string, 0, min(DefaultMaxLines, len(lines)))
	bytesUsed := 0
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		lineBytes := len(line)
		if len(selected) > 0 {
			lineBytes++
		}
		if len(selected) >= DefaultMaxLines {
			break
		}
		if bytesUsed+lineBytes > DefaultMaxBytes {
			break
		}
		selected = append(selected, line)
		bytesUsed += lineBytes
	}
	for i, j := 0, len(selected)-1; i < j; i, j = i+1, j-1 {
		selected[i], selected[j] = selected[j], selected[i]
	}
	truncatedBy := "lines"
	if len(text) > DefaultMaxBytes {
		truncatedBy = "bytes"
	}
	return TruncationDetails{Truncated: true, TruncatedBy: truncatedBy, TotalLines: len(lines), OutputLines: len(selected), OutputBytes: bytesUsed, MaxLines: DefaultMaxLines, MaxBytes: DefaultMaxBytes}, strings.Join(selected, "\n")
}

func shellCommand(ctx context.Context, command string) *exec.Cmd {
	if isWindows() {
		return exec.CommandContext(ctx, "cmd", "/C", command)
	}
	return exec.CommandContext(ctx, "sh", "-lc", command)
}
