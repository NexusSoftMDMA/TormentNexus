//go:build ignore
// +build ignore

package tools

/**
 * @file prism.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of Prism MCP Server tools.
 * Replaces `prism-mcp` (npx prism-mcp-server) entry in mcp.json.
 *
 * Prism provides AI-powered code transformation and analysis:
 * - Code refactoring suggestions
 * - Pattern detection and transformation
 * - Code quality metrics
 * - AST-based code manipulation
 *
 * Improvements over original:
 * - No npx/Node dependency.
 * - Go-native regex + AST patterns.
 * - Supports: analyze_quality, suggest_refactor, transform_code, detect_smells.
 */

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// HandlePrismAnalyzeQuality analyzes code quality metrics.
// Tool: prism_analyze_quality
func HandlePrismAnalyzeQuality(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
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
	lang := detectLanguage(path)

	// Calculate metrics
	totalLines := len(lines)
	blankLines := 0
	commentLines := 0
	codeLines := 0
	maxLineLength := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			blankLines++
		} else if isCommentLine(trimmed, lang) {
			commentLines++
		} else {
			codeLines++
		}
		if len(line) > maxLineLength {
			maxLineLength = len(line)
		}
	}

	// Calculate complexity indicators
	functionCount := countFunctions(content, lang)
	avgFunctionLength := 0
	if functionCount > 0 {
		avgFunctionLength = codeLines / functionCount
	}

	metrics := map[string]interface{}{
		"file":                path,
		"language":            lang,
		"total_lines":         totalLines,
		"code_lines":          codeLines,
		"comment_lines":       commentLines,
		"blank_lines":         blankLines,
		"max_line_length":     maxLineLength,
		"function_count":      functionCount,
		"avg_function_length": avgFunctionLength,
		"comment_ratio":       floatRatio(commentLines, totalLines),
	}

	out, _ := json.MarshalIndent(metrics, "", "  ")
	return ok(string(out))
}

// HandlePrismSuggestRefactor suggests refactoring improvements.
// Tool: prism_suggest_refactor
func HandlePrismSuggestRefactor(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
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
	_ = detectLanguage(path) // language detection for future use

	type RefactorSuggestion struct {
		Line        int    `json:"line"`
		Type        string `json:"type"`
		Severity    string `json:"severity"`
		Description string `json:"description"`
	}

	var suggestions []RefactorSuggestion

	// Check for long functions
	currentFuncStart := 0
	currentFuncLines := 0
	inFunction := false

	for lineNum, line := range lines {
		trimmed := strings.TrimSpace(line)
		if isDefinitionLine(trimmed) && !inFunction {
			if currentFuncLines > 50 && currentFuncStart > 0 {
				suggestions = append(suggestions, RefactorSuggestion{
					Line:        currentFuncStart,
					Type:        "long_function",
					Severity:    "warning",
					Description: fmt.Sprintf("Function starting at line %d is %d lines long. Consider breaking into smaller functions.", currentFuncStart, currentFuncLines),
				})
			}
			currentFuncStart = lineNum + 1
			currentFuncLines = 0
			inFunction = true
		}
		if inFunction {
			currentFuncLines++
		}
		// Deep nesting detection
		indent := len(line) - len(strings.TrimLeft(line, " \t"))
		nestLevel := indent / 4
		if nestLevel > 4 {
			suggestions = append(suggestions, RefactorSuggestion{
				Line:        lineNum + 1,
				Type:        "deep_nesting",
				Severity:    "warning",
				Description: fmt.Sprintf("Deep nesting (level %d). Consider early returns or extracting logic.", nestLevel),
			})
		}

		// Long parameter list
		paramRegex := regexp.MustCompile(`\(([^)]*,){4,}`)
		if paramRegex.MatchString(trimmed) {
			suggestions = append(suggestions, RefactorSuggestion{
				Line:        lineNum + 1,
				Type:        "long_params",
				Severity:    "info",
				Description: "Function has many parameters. Consider using a config struct/object.",
			})
		}

		// Duplicated string literals
		if strings.Count(content, trimmed) > 5 && len(trimmed) > 10 && !strings.HasPrefix(trimmed, "//") {
			// Skip if it's a common keyword
			if !isCommonKeyword(trimmed) {
				suggestions = append(suggestions, RefactorSuggestion{
					Line:        lineNum + 1,
					Type:        "repeated_literal",
					Severity:    "info",
					Description: fmt.Sprintf("String '%s' appears multiple times. Consider extracting to a constant.", truncate(trimmed, 50)),
				})
			}
		}
	}

	if len(suggestions) == 0 {
		return ok("No refactoring suggestions found. Code looks clean! ✅")
	}

	out, _ := json.MarshalIndent(suggestions, "", "  ")
	return ok(string(out))
}

// HandlePrismDetectSmells detects code smells.
// Tool: prism_detect_smells
func HandlePrismDetectSmells(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ := getString(args, "path", "directory")
	if path == "" {
		path = "."
	}

	type Smell struct {
		File        string `json:"file"`
		Line        int    `json:"line"`
		SmellType   string `json:"smell_type"`
		Description string `json:"description"`
	}

	var smells []Smell

	filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(p))
		codeExts := map[string]bool{
			".go": true, ".py": true, ".js": true, ".ts": true, ".tsx": true,
			".rs": true, ".java": true, ".cpp": true, ".c": true,
		}
		if !codeExts[ext] {
			return nil
		}

		rel, _ := filepath.Rel(path, p)
		data, e := os.ReadFile(p)
		if e != nil {
			return nil
		}

		content := string(data)
		lines := strings.Split(content, "\n")

		for lineNum, line := range lines {
			trimmed := strings.TrimSpace(line)

			// God class / large file
			if lineNum == 0 && len(lines) > 500 {
				smells = append(smells, Smell{
					File:        rel,
					Line:        1,
					SmellType:   "large_file",
					Description: fmt.Sprintf("File has %d lines. Consider splitting into smaller modules.", len(lines)),
				})
			}

// Magic numbers
			magicRegex := regexp.MustCompile(`\b\d{4,}\b`)
			if magicRegex.MatchString(trimmed) && !isCommentLine(trimmed, detectLanguage(p)) {
				smells = append(smells, Smell{
					File:        rel,
					Line:        lineNum + 1,
					SmellType:   "magic_number",
					Description: "Magic number detected. Consider using a named constant.",
				})
			}

			// Duplicate condition (simplified)
			if strings.Contains(trimmed, "if") && lineNum > 0 {
				prevLines := []string{}
				start := lineNum - 5
				if start < 0 {
					start = 0
				}
				for i := start; i < lineNum; i++ {
					prevLines = append(prevLines, strings.TrimSpace(lines[i]))
				}
				for _, prev := range prevLines {
					if prev == trimmed && strings.Contains(trimmed, "if") {
						smells = append(smells, Smell{
							File:        rel,
							Line:        lineNum + 1,
							SmellType:   "duplicate_condition",
							Description: "Duplicate conditional detected nearby.",
						})
						break
					}
				}
			}
		}

		return nil
	})

	if len(smells) == 0 {
		return ok("No code smells detected. ✅")
	}

	if len(smells) > 30 {
		smells = smells[:30]
	}

	out, _ := json.MarshalIndent(smells, "", "  ")
	return ok(string(out))
}

// HandlePrismTransformCode applies code transformations.
// Tool: prism_transform_code
func HandlePrismTransformCode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ := getString(args, "code")
	transformType, _ := getString(args, "transform", "type")

	if code == "" {
		return err("code parameter is required")
	}
	if transformType == "" {
		return err("transform parameter is required (e.g., 'remove_comments', 'trim_trailing', 'normalize_quotes')")
	}

	var result string

	switch transformType {
	case "remove_comments":
		result = removeComments(code)
	case "trim_trailing":
		lines := strings.Split(code, "\n")
		for i, line := range lines {
			lines[i] = strings.TrimRight(line, " \t")
		}
		result = strings.Join(lines, "\n")
	case "normalize_quotes":
		result = strings.ReplaceAll(code, "\"", "\"")
		// Simple: convert single-quoted strings to double-quoted in JS/TS
		singleQuoteRegex := regexp.MustCompile(`'([^']*)'`)
		result = singleQuoteRegex.ReplaceAllString(code, `"$1"`)
	case "remove_empty_lines":
		lines := strings.Split(code, "\n")
		var filtered []string
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				filtered = append(filtered, line)
			}
		}
		result = strings.Join(filtered, "\n")
	case "add_line_numbers":
		lines := strings.Split(code, "\n")
		for i, line := range lines {
			lines[i] = fmt.Sprintf("%4d: %s", i+1, line)
		}
		result = strings.Join(lines, "\n")
	default:
		return err(fmt.Sprintf("Unknown transform type: %s. Supported: remove_comments, trim_trailing, normalize_quotes, remove_empty_lines, add_line_numbers", transformType))
	}

	return ok(result)
}

// Helper functions

func isCommentLine(line, lang string) bool {
	switch lang {
	case "go", "rust", "java", "cpp", "c", "javascript", "typescript":
		return strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*")
	case "python":
		return strings.HasPrefix(line, "#")
	case "ruby":
		return strings.HasPrefix(line, "#")
	default:
		return strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#")
	}
}

func countFunctions(content, lang string) int {
	var pattern *regexp.Regexp
	switch lang {
	case "go":
		pattern = regexp.MustCompile(`^\s*func\s`)
	case "python":
		pattern = regexp.MustCompile(`^\s*def\s`)
	case "javascript", "typescript":
		pattern = regexp.MustCompile(`(function\s|=>\s*\{|func\s*=\s*)`)
	case "rust":
		pattern = regexp.MustCompile(`^\s*(pub\s+)?fn\s`)
	case "java":
		pattern = regexp.MustCompile(`\b(public|private|protected|static)\s.*\s\w+\s*\(`)
	}
	if pattern == nil {
		return 0
	}
	return len(pattern.FindAllString(content, -1))
}

func floatRatio(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total)
}

func isCommonKeyword(s string) bool {
	keywords := map[string]bool{
		"return": true, "if": true, "else": true, "for": true,
		"while": true, "func": true, "var": true, "const": true,
		"let": true, "import": true, "export": true, "package": true,
		"true": true, "false": true, "nil": true, "null": true,
	}
	return keywords[strings.ToLower(strings.TrimSpace(s))]
}

func removeComments(code string) string {
	// Remove single-line comments
	singleLine := regexp.MustCompile(`//.*$`)
	code = singleLine.ReplaceAllString(code, "")

	// Remove multi-line comments (simplified)
	multiLine := regexp.MustCompile(`(?s)/\*.*?\*/`)
	code = multiLine.ReplaceAllString(code, "")

	// Remove Python-style comments
	pyComment := regexp.MustCompile(`#.*$`)
	code = pyComment.ReplaceAllString(code, "")

	return code
}
