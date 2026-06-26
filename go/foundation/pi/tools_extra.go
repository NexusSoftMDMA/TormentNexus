package pi

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	defaultGrepLimit  = 100
	defaultFindLimit  = 1000
	defaultLsLimit    = 500
	grepMaxLineLength = 500
)

func executeGrepTool(ctx context.Context, cwd string, raw json.RawMessage) (*ToolResult, error) {
	var input GrepToolInput
	if err := json.Unmarshal(raw, &input); err != nil {
		return &ToolResult{ToolName: "grep", Content: []any{TextContent{Type: "text", Text: fmt.Sprintf("invalid grep input: %v", err)}}, IsError: true}, nil
	}
	if strings.TrimSpace(input.Pattern) == "" {
		return &ToolResult{ToolName: "grep", Content: []any{TextContent{Type: "text", Text: "pattern is required"}}, IsError: true}, nil
	}

	searchPath := cwd
	if input.Path != "" {
		searchPath = resolveOrDefault(cwd, input.Path)
	}

	effectiveLimit := input.Limit
	if effectiveLimit <= 0 {
		effectiveLimit = defaultGrepLimit
	}

	// Try ripgrep first
	rgPath, err := exec.LookPath("rg")
	if err == nil {
		return executeGrepWithRipgrep(ctx, cwd, searchPath, rgPath, input, effectiveLimit)
	}

	// Fall back to native Go implementation
	return executeGrepNative(ctx, cwd, searchPath, input, effectiveLimit)
}

func executeGrepWithRipgrep(ctx context.Context, cwd, searchPath, rgPath string, input GrepToolInput, effectiveLimit int) (*ToolResult, error) {
	args := []string{"--json", "--line-number", "--color=never", "--hidden"}
	if input.IgnoreCase {
		args = append(args, "--ignore-case")
	}
	if input.Literal {
		args = append(args, "--fixed-strings")
	}
	if input.Glob != "" {
		args = append(args, "--glob", input.Glob)
	}
	args = append(args, "--max-count", fmt.Sprintf("%d", effectiveLimit))
	if input.Context > 0 {
		args = append(args, "-C", fmt.Sprintf("%d", input.Context))
	}
	args = append(args, input.Pattern, searchPath)

	cmd := exec.CommandContext(ctx, rgPath, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return &ToolResult{ToolName: "grep", Content: []any{TextContent{Type: "text", Text: fmt.Sprintf("failed to run ripgrep: %v", err)}}, IsError: true}, nil
	}
	stderrBuf := &bytes.Buffer{}
	cmd.Stderr = stderrBuf

	if err := cmd.Start(); err != nil {
		return &ToolResult{ToolName: "grep", Content: []any{TextContent{Type: "text", Text: fmt.Sprintf("failed to run ripgrep: %v", err)}}, IsError: true}, nil
	}

	searchInfo, _ := os.Stat(searchPath)
	isDirSearch := searchInfo != nil && searchInfo.IsDir()

	var outputLines []string
	matchCount := 0
	scanner := newScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, `"type"`) {
			continue
		}
		var event map[string]interface{}
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}
		if event["type"] == "match" {
			matchCount++
			data, _ := event["data"].(map[string]interface{})
			if data == nil {
				continue
			}
			pathVal, _ := data["path"].(map[string]interface{})
			filePath, _ := pathVal["text"].(string)
			lineNum, _ := data["line_number"].(float64)
			if filePath == "" || lineNum == 0 {
				continue
			}

			relPath := filepath.ToSlash(relPathOrBase(filePath, searchPath))
			lines, _ := readFileLines(filePath)
			if len(lines) == 0 {
				outputLines = append(outputLines, fmt.Sprintf("%s:%d: (unable to read file)", relPath, int(lineNum)))
				continue
			}

			start := int(lineNum)
			if input.Context > 0 {
				for i := max(1, start-input.Context); i <= min(len(lines), start+input.Context); i++ {
					txt := strings.ReplaceAll(lines[i-1], "\r", "")
					tc, _ := truncateLineStr(txt)
					if i == start {
						outputLines = append(outputLines, fmt.Sprintf("%s:%d: %s", relPath, i, tc))
					} else {
						outputLines = append(outputLines, fmt.Sprintf("%s-%d- %s", relPath, i, tc))
					}
				}
			} else {
				txt := strings.ReplaceAll(lines[start-1], "\r", "")
				tc, wasTruncated := truncateLineStr(txt)
				if !isDirSearch {
					outputLines = append(outputLines, fmt.Sprintf("%s:%d: %s", filepath.Base(filePath), int(lineNum), tc))
				} else {
					outputLines = append(outputLines, fmt.Sprintf("%s:%d: %s", relPath, int(lineNum), tc))
				}
				if wasTruncated {
					// line was truncated, noted implicitly
				}
			}

			if matchCount >= effectiveLimit {
				break
			}
		}
	}
	cmd.Wait()

	if matchCount == 0 {
		return &ToolResult{ToolName: "grep", Content: []any{TextContent{Type: "text", Text: "No matches found"}}}, nil
	}

	rawOutput := strings.Join(outputLines, "\n")
	truncation, output := truncateHead(rawOutput)

	var details *GrepToolDetails
	notices := []string{}
	if matchCount >= effectiveLimit {
		notices = append(notices, fmt.Sprintf("%d matches limit reached. Use limit=%d for more, or refine pattern", effectiveLimit, effectiveLimit*2))
		details = &GrepToolDetails{MatchLimitReached: effectiveLimit}
	}
	if truncation.Truncated {
		notices = append(notices, fmt.Sprintf("%d byte limit reached", DefaultMaxBytes))
		if details == nil {
			details = &GrepToolDetails{}
		}
		details.Truncation = &truncation
	}
	if len(notices) > 0 {
		output += fmt.Sprintf("\n\n[%s]", strings.Join(notices, ". "))
	}

	return &ToolResult{ToolName: "grep", Content: []any{TextContent{Type: "text", Text: output}}, Details: details}, nil
}

func executeGrepNative(ctx context.Context, cwd, searchPath string, input GrepToolInput, effectiveLimit int) (*ToolResult, error) {
	var pattern *regexp.Regexp
	if input.Literal {
		pattern = regexp.MustCompile(regexp.QuoteMeta(input.Pattern))
	} else {
		pat := input.Pattern
		if input.IgnoreCase {
			pat = "(?i)" + pat
		}
		var err error
		pattern, err = regexp.Compile(pat)
		if err != nil {
			return &ToolResult{ToolName: "grep", Content: []any{TextContent{Type: "text", Text: fmt.Sprintf("invalid regex: %v", err)}}, IsError: true}, nil
		}
	}

	var files []string
	info, err := os.Stat(searchPath)
	if err != nil {
		return &ToolResult{ToolName: "grep", Content: []any{TextContent{Type: "text", Text: fmt.Sprintf("path not found: %s", searchPath)}}, IsError: true}, nil
	}

	isDirSearch := info.IsDir()
	if isDirSearch {
		err := filepath.WalkDir(searchPath, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() && (d.Name() == ".git" || d.Name() == "node_modules") {
				return filepath.SkipDir
			}
			if d.IsDir() {
				return nil
			}
			if input.Glob != "" {
				matched, _ := filepath.Match(input.Glob, filepath.Base(path))
				if !matched {
					return nil
				}
			}
			files = append(files, path)
			return nil
		})
		if err != nil {
			return &ToolResult{ToolName: "grep", Content: []any{TextContent{Type: "text", Text: fmt.Sprintf("walk error: %v", err)}}, IsError: true}, nil
		}
	} else {
		files = append(files, searchPath)
	}

	var outputLines []string
	matchCount := 0

	for _, file := range files {
		if matchCount >= effectiveLimit {
			break
		}
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
		for i, line := range lines {
			if matchCount >= effectiveLimit {
				break
			}
			if pattern.MatchString(line) {
				matchCount++
				relPath := filepath.ToSlash(relPathOrBase(file, searchPath))
				if !isDirSearch {
					relPath = filepath.Base(file)
				}
				tc, _ := truncateLineStr(line)

				if input.Context > 0 {
					start := max(1, i+1-input.Context)
					end := min(len(lines), i+1+input.Context)
					for j := start; j <= end; j++ {
						t, _ := truncateLineStr(strings.ReplaceAll(lines[j-1], "\r", ""))
						if j == i+1 {
							outputLines = append(outputLines, fmt.Sprintf("%s:%d: %s", relPath, j, t))
						} else {
							outputLines = append(outputLines, fmt.Sprintf("%s-%d- %s", relPath, j, t))
						}
					}
				} else {
					outputLines = append(outputLines, fmt.Sprintf("%s:%d: %s", relPath, i+1, tc))
				}
			}
		}
	}

	if matchCount == 0 {
		return &ToolResult{ToolName: "grep", Content: []any{TextContent{Type: "text", Text: "No matches found"}}}, nil
	}

	rawOutput := strings.Join(outputLines, "\n")
	truncation, output := truncateHead(rawOutput)

	var details *GrepToolDetails
	notices := []string{}
	if matchCount >= effectiveLimit {
		notices = append(notices, fmt.Sprintf("%d matches limit reached. Use limit=%d for more, or refine pattern", effectiveLimit, effectiveLimit*2))
		details = &GrepToolDetails{MatchLimitReached: effectiveLimit}
	}
	if truncation.Truncated {
		notices = append(notices, fmt.Sprintf("%d byte limit reached", DefaultMaxBytes))
		if details == nil {
			details = &GrepToolDetails{}
		}
		details.Truncation = &truncation
	}
	if len(notices) > 0 {
		output += fmt.Sprintf("\n\n[%s]", strings.Join(notices, ". "))
	}

	return &ToolResult{ToolName: "grep", Content: []any{TextContent{Type: "text", Text: output}}, Details: details}, nil
}

func executeFindTool(ctx context.Context, cwd string, raw json.RawMessage) (*ToolResult, error) {
	var input FindToolInput
	if err := json.Unmarshal(raw, &input); err != nil {
		return &ToolResult{ToolName: "find", Content: []any{TextContent{Type: "text", Text: fmt.Sprintf("invalid find input: %v", err)}}, IsError: true}, nil
	}
	if strings.TrimSpace(input.Pattern) == "" {
		return &ToolResult{ToolName: "find", Content: []any{TextContent{Type: "text", Text: "pattern is required"}}, IsError: true}, nil
	}

	searchPath := cwd
	if input.Path != "" {
		searchPath = resolveOrDefault(cwd, input.Path)
	}

	effectiveLimit := input.Limit
	if effectiveLimit <= 0 {
		effectiveLimit = defaultFindLimit
	}

	// Try fd first
	fdPath, err := exec.LookPath("fd")
	if err == nil {
		return executeFindWithFd(ctx, searchPath, fdPath, input.Pattern, effectiveLimit)
	}

	return executeFindNative(ctx, searchPath, input.Pattern, effectiveLimit)
}

func executeFindWithFd(ctx context.Context, searchPath, fdPath, pattern string, effectiveLimit int) (*ToolResult, error) {
	args := []string{"--glob", "--color=never", "--hidden", "--max-results", fmt.Sprintf("%d", effectiveLimit), pattern, searchPath}
	cmd := exec.CommandContext(ctx, fdPath, args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return &ToolResult{ToolName: "find", Content: []any{TextContent{Type: "text", Text: "No files found matching pattern"}}}, nil
		}
		return &ToolResult{ToolName: "find", Content: []any{TextContent{Type: "text", Text: fmt.Sprintf("failed to run fd: %v", err)}}, IsError: true}, nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var results []string
	resultLimitReached := false
	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		if len(results) >= effectiveLimit {
			resultLimitReached = true
			break
		}
		relPath := relPathOrBase(line, searchPath)
		results = append(results, filepath.ToSlash(relPath))
	}

	return formatFindResults(results, resultLimitReached || len(results) >= effectiveLimit, effectiveLimit)
}

func executeFindNative(ctx context.Context, searchPath, pattern string, effectiveLimit int) (*ToolResult, error) {
	var results []string

	err := filepath.WalkDir(searchPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && (d.Name() == ".git" || d.Name() == "node_modules") {
			return filepath.SkipDir
		}
		if d.IsDir() {
			return nil
		}
		if len(results) >= effectiveLimit {
			return nil
		}

		matched, _ := filepath.Match(pattern, filepath.Base(path))
		if !matched {
			relPath, _ := filepath.Rel(searchPath, path)
			matched, _ = filepath.Match(pattern, relPath)
		}
		if matched {
			relPath := filepath.ToSlash(relPathOrBase(path, searchPath))
			results = append(results, relPath)
		}
		return nil
	})
	if err != nil {
		return &ToolResult{ToolName: "find", Content: []any{TextContent{Type: "text", Text: fmt.Sprintf("walk error: %v", err)}}, IsError: true}, nil
	}

	resultLimitReached := len(results) >= effectiveLimit
	return formatFindResults(results, resultLimitReached || len(results) >= effectiveLimit, effectiveLimit)
}

func formatFindResults(results []string, resultLimitReached bool, effectiveLimit int) (*ToolResult, error) {
	if len(results) == 0 {
		return &ToolResult{ToolName: "find", Content: []any{TextContent{Type: "text", Text: "No files found matching pattern"}}}, nil
	}

	rawOutput := strings.Join(results, "\n")
	truncation, outputText := truncateHead(rawOutput)

	var details *FindToolDetails
	notices := []string{}
	if resultLimitReached {
		notices = append(notices, fmt.Sprintf("%d results limit reached. Use limit=%d for more, or refine pattern", effectiveLimit, effectiveLimit*2))
		details = &FindToolDetails{ResultLimitReached: effectiveLimit}
	}
	if truncation.Truncated {
		notices = append(notices, fmt.Sprintf("%d byte limit reached", DefaultMaxBytes))
		if details == nil {
			details = &FindToolDetails{}
		}
		details.Truncation = &truncation
	}
	if len(notices) > 0 {
		outputText += fmt.Sprintf("\n\n[%s]", strings.Join(notices, ". "))
	}

	return &ToolResult{ToolName: "find", Content: []any{TextContent{Type: "text", Text: outputText}}, Details: details}, nil
}

func executeLsTool(ctx context.Context, cwd string, raw json.RawMessage) (*ToolResult, error) {
	var input LsToolInput
	if err := json.Unmarshal(raw, &input); err != nil {
		return &ToolResult{ToolName: "ls", Content: []any{TextContent{Type: "text", Text: fmt.Sprintf("invalid ls input: %v", err)}}, IsError: true}, nil
	}

	dirPath := cwd
	if input.Path != "" {
		dirPath = resolveOrDefault(cwd, input.Path)
	}

	effectiveLimit := input.Limit
	if effectiveLimit <= 0 {
		effectiveLimit = defaultLsLimit
	}

	info, err := os.Stat(dirPath)
	if err != nil {
		return &ToolResult{ToolName: "ls", Content: []any{TextContent{Type: "text", Text: fmt.Sprintf("path not found: %s", dirPath)}}, IsError: true}, nil
	}
	if !info.IsDir() {
		return &ToolResult{ToolName: "ls", Content: []any{TextContent{Type: "text", Text: fmt.Sprintf("not a directory: %s", dirPath)}}, IsError: true}, nil
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return &ToolResult{ToolName: "ls", Content: []any{TextContent{Type: "text", Text: fmt.Sprintf("cannot read directory: %v", err)}}, IsError: true}, nil
	}

	var results []string
	entryLimitReached := false
	for _, entry := range entries {
		if len(results) >= effectiveLimit {
			entryLimitReached = true
			break
		}
		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}
		results = append(results, name)
	}

	sort.Slice(results, func(i, j int) bool {
		return strings.ToLower(results[i]) < strings.ToLower(results[j])
	})

	if len(results) == 0 {
		return &ToolResult{ToolName: "ls", Content: []any{TextContent{Type: "text", Text: "(empty directory)"}}, Details: nil}, nil
	}

	rawOutput := strings.Join(results, "\n")
	truncation, output := truncateHead(rawOutput)

	var details *LsToolDetails
	notices := []string{}
	if entryLimitReached {
		notices = append(notices, fmt.Sprintf("%d entries limit reached. Use limit=%d for more", effectiveLimit, effectiveLimit*2))
		details = &LsToolDetails{EntryLimitReached: effectiveLimit}
	}
	if truncation.Truncated {
		notices = append(notices, fmt.Sprintf("%d byte limit reached", DefaultMaxBytes))
		if details == nil {
			details = &LsToolDetails{}
		}
		details.Truncation = &truncation
	}
	if len(notices) > 0 {
		output += fmt.Sprintf("\n\n[%s]", strings.Join(notices, ". "))
	}

	return &ToolResult{ToolName: "ls", Content: []any{TextContent{Type: "text", Text: output}}, Details: details}, nil
}

func resolveOrDefault(cwd, p string) string {
	if p == "." {
		return cwd
	}
	if filepath.IsAbs(p) {
		return filepath.Clean(p)
	}
	return filepath.Clean(filepath.Join(cwd, p))
}

func relPathOrBase(path, base string) string {
	if rel, err := filepath.Rel(base, path); err == nil && !strings.HasPrefix(rel, "..") {
		return rel
	}
	return filepath.Base(path)
}

func readFileLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n"), nil
}

func truncateLineStr(s string) (string, bool) {
	runes := []rune(s)
	if len(runes) > grepMaxLineLength {
		return string(runes[:grepMaxLineLength]), true
	}
	return s, false
}

type lineScanner struct {
	r    *bufio.Reader
	line string
	done bool
}

func newScanner(r io.Reader) *lineScanner {
	return &lineScanner{r: bufio.NewReader(r)}
}

func (s *lineScanner) Scan() bool {
	if s.done {
		return false
	}
	line, err := s.r.ReadString('\n')
	if err != nil {
		if len(line) == 0 {
			s.done = true
			return false
		}
	}
	s.line = strings.TrimRight(line, "\r\n")
	return true
}

func (s *lineScanner) Text() string { return s.line }
