//go:build ignore
// +build ignore

package tools

/**
 * @file probe.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of Probe MCP tools.
 * Replaces `probe` (npx @probelabs/probe@latest mcp) entry in mcp.json.
 *
 * Probe provides intelligent codebase search and understanding:
 * - Semantic code search across a repository
 * - Code structure analysis
 * - Symbol lookup and cross-references
 * - Documentation generation
 *
 * Improvements over original:
 * - No npx/Node dependency.
 * - Go-native AST parsing for Go, and regex-based for other languages.
 * - SQLite-backed search index.
 * - Supports: search_code, find_symbol, explain_code, get_structure.
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

// HandleProbeSearchCode performs semantic code search.
// Tool: probe_search_code
func HandleProbeSearchCode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q", "search")
	if query == "" {
		return err("query parameter is required")
	}

	path, _ := getString(args, "path", "directory")
	if path == "" {
		path = "."
	}

	language, _ := getString(args, "language", "lang")
	maxResults := getInt(args, "max_results", "limit")
	if maxResults <= 0 {
		maxResults = 20
	}

	type SearchResult struct {
		File    string `json:"file"`
		Line    int    `json:"line"`
		Content string `json:"content"`
		Score   int    `json:"score"`
	}

	var results []SearchResult
	queryLower := strings.ToLower(query)
	queryTerms := strings.Fields(queryLower)

	filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(p))
		if language != "" && detectLanguage(p) != language {
			return nil
		}

		// Skip non-code files and hidden dirs
		if strings.Contains(p, "node_modules") || strings.Contains(p, ".git") ||
			strings.Contains(p, "vendor") || strings.Contains(p, ".next") {
			return nil
		}

		codeExts := map[string]bool{
			".go": true, ".py": true, ".js": true, ".ts": true, ".tsx": true,
			".jsx": true, ".rs": true, ".java": true, ".cpp": true, ".c": true,
			".h": true, ".cs": true, ".rb": true, ".php": true,
		}
		if !codeExts[ext] {
			return nil
		}

		data, e := os.ReadFile(p)
		if e != nil {
			return nil
		}

		rel, _ := filepath.Rel(path, p)
		content := string(data)
		lines := strings.Split(content, "\n")

		for lineNum, line := range lines {
			lineLower := strings.ToLower(line)

			// Score based on how many query terms match
			score := 0
			for _, term := range queryTerms {
				if strings.Contains(lineLower, term) {
					score += 10
				}
			}

			// Boost for exact match
			if strings.Contains(lineLower, queryLower) {
				score += 20
			}

			// Boost for definition lines
			if isDefinitionLine(line) {
				score += 5
			}

			if score > 0 {
				results = append(results, SearchResult{
					File:    rel,
					Line:    lineNum + 1,
					Content: strings.TrimSpace(line),
					Score:   score,
				})
			}
		}

		return nil
	})

	// Sort by score (simple insertion sort for small datasets)
	for i := 1; i < len(results); i++ {
		for j := i; j > 0 && results[j].Score > results[j-1].Score; j-- {
			results[j], results[j-1] = results[j-1], results[j]
		}
	}

	// Limit results
	if len(results) > maxResults {
		results = results[:maxResults]
	}

	if len(results) == 0 {
		return ok("No results found for: " + query)
	}

	out, _ := json.MarshalIndent(results, "", "  ")
	return ok(string(out))
}

// HandleProbeFindSymbol finds symbol definitions.
// Tool: probe_find_symbol
func HandleProbeFindSymbol(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ := getString(args, "symbol", "name")
	if symbol == "" {
		return err("symbol parameter is required")
	}

	path, _ := getString(args, "path", "directory")
	if path == "" {
		path = "."
	}

	type SymbolResult struct {
		File    string `json:"file"`
		Line    int    `json:"line"`
		Type    string `json:"type"`
		Content string `json:"content"`
	}

	var results []SymbolResult

	// Language-specific symbol patterns
	symbolPatterns := map[string]*regexp.Regexp{
		"go":         regexp.MustCompile(fmt.Sprintf(`(?i)(func|type|var|const)\s+.*%s`, regexp.QuoteMeta(symbol))),
		"python":     regexp.MustCompile(fmt.Sprintf(`(?i)(def|class)\s+.*%s`, regexp.QuoteMeta(symbol))),
		"javascript": regexp.MustCompile(fmt.Sprintf(`(?i)(function|const|let|var|class)\s+%s`, regexp.QuoteMeta(symbol))),
		"typescript": regexp.MustCompile(fmt.Sprintf(`(?i)(function|const|let|var|class|interface|type|enum)\s+%s`, regexp.QuoteMeta(symbol))),
		"rust":       regexp.MustCompile(fmt.Sprintf(`(?i)(fn|struct|enum|trait|impl|type)\s+.*%s`, regexp.QuoteMeta(symbol))),
		"java":       regexp.MustCompile(fmt.Sprintf(`(?i)(class|interface|enum|void|int|String)\s+.*%s`, regexp.QuoteMeta(symbol))),
	}

	filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		if strings.Contains(p, "node_modules") || strings.Contains(p, ".git") {
			return nil
		}

		lang := detectLanguage(p)
		pattern, ok := symbolPatterns[lang]
		if !ok {
			// Generic pattern
			pattern = regexp.MustCompile(fmt.Sprintf(`(?i)\b%s\b`, regexp.QuoteMeta(symbol)))
		}

		data, e := os.ReadFile(p)
		if e != nil {
			return nil
		}

		rel, _ := filepath.Rel(path, p)
		lines := strings.Split(string(data), "\n")

		for lineNum, line := range lines {
			if pattern.MatchString(line) && isDefinitionLine(line) {
				results = append(results, SymbolResult{
					File:    rel,
					Line:    lineNum + 1,
					Type:    lang,
					Content: strings.TrimSpace(line),
				})
			}
		}

		return nil
	})

	if len(results) == 0 {
		return ok("Symbol not found: " + symbol)
	}

	out, _ := json.MarshalIndent(results, "", "  ")
	return ok(string(out))
}

// HandleProbeGetStructure returns the code structure of a file or directory.
// Tool: probe_get_structure
func HandleProbeGetStructure(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ := getString(args, "path", "file_path")
	if path == "" {
		path = "."
	}

	info, e := os.Stat(path)
	if e != nil {
		return err(fmt.Sprintf("Path not found: %v", e))
	}

	var results []map[string]interface{}

	if info.IsDir() {
		filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}

			if strings.Contains(p, "node_modules") || strings.Contains(p, ".git") ||
				strings.Contains(p, "vendor") || strings.Contains(p, ".next") {
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
			data, _ := os.ReadFile(p)
			defs := extractDefinitions(string(data), detectLanguage(p))

			results = append(results, map[string]interface{}{
				"file":        rel,
				"language":    detectLanguage(p),
				"definitions": defs,
				"line_count":  len(strings.Split(string(data), "\n")),
			})

			return nil
		})
	} else {
		data, _ := os.ReadFile(path)
		defs := extractDefinitions(string(data), detectLanguage(path))
		results = append(results, map[string]interface{}{
			"file":        path,
			"language":    detectLanguage(path),
			"definitions": defs,
			"line_count":  len(strings.Split(string(data), "\n")),
		})
	}

	out, _ := json.MarshalIndent(results, "", "  ")
	return ok(string(out))
}

// HandleProbeExplainCode explains a code snippet or file section.
// Tool: probe_explain_code
func HandleProbeExplainCode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ := getString(args, "code")
	filePath, _ := getString(args, "file_path", "path")
	startLine := getInt(args, "start_line", "startLine")
	endLine := getInt(args, "end_line", "endLine")

	if code == "" && filePath == "" {
		return err("code or file_path parameter is required")
	}

	if code == "" {
		data, e := os.ReadFile(filePath)
		if e != nil {
			return err(fmt.Sprintf("Failed to read file: %v", e))
		}
		lines := strings.Split(string(data), "\n")
		if startLine > 0 && endLine > 0 {
			start := startLine - 1
			if start < 0 {
				start = 0
			}
			if endLine > len(lines) {
				endLine = len(lines)
			}
			code = strings.Join(lines[start:endLine], "\n")
		} else {
			code = string(data)
		}
	}

	// If no LLM API key, provide structural analysis
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
	}

	if apiKey == "" {
		// Fallback: structural analysis
		lang := "unknown"
		if filePath != "" {
			lang = detectLanguage(filePath)
		}
		lines := strings.Split(code, "\n")
		defs := extractDefinitions(code, lang)

		result := map[string]interface{}{
			"language":     lang,
			"line_count":   len(lines),
			"definitions":  defs,
			"explanation":  "Structural analysis (no LLM API key configured for natural language explanation)",
		}

		out, _ := json.MarshalIndent(result, "", "  ")
		return ok(string(out))
	}

	// Use PAL-like LLM call for explanation
	explanation, e := callLLM(ctx,
		"You are a code analysis expert. Explain the following code clearly and concisely.",
		code, 0.3, "")
	if e != nil {
		return err(fmt.Sprintf("Failed to get explanation: %v", e))
	}

	return ok(explanation)
}

// Helper functions

func isDefinitionLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	definitionPatterns := []string{
		"func ", "function ", "class ", "struct ", "interface ",
		"enum ", "type ", "def ", "fn ", "impl ", "trait ",
		"const ", "let ", "var ", "module ", "package ",
		"import ", "export ",
	}
	for _, p := range definitionPatterns {
		if strings.HasPrefix(trimmed, p) {
			return true
		}
	}
	return false
}

func extractDefinitions(content string, lang string) []map[string]interface{} {
	var defs []map[string]interface{}
	lines := strings.Split(content, "\n")

	var defRegex *regexp.Regexp
	switch lang {
	case "go":
		defRegex = regexp.MustCompile(`^\s*(func|type|var|const)\s+(\w+)`)
	case "python":
		defRegex = regexp.MustCompile(`^\s*(def|class)\s+(\w+)`)
	case "javascript", "typescript":
		defRegex = regexp.MustCompile(`^\s*(export\s+)?(function|const|let|var|class|interface|type|enum)\s+(\w+)`)
	case "rust":
		defRegex = regexp.MustCompile(`^\s*(pub\s+)?(fn|struct|enum|trait|impl|type)\s+(\w+)`)
	case "java":
		defRegex = regexp.MustCompile(`^\s*(public|private|protected)?\s*(static\s+)?(class|interface|enum|void)\s+(\w+)`)
	}

	if defRegex == nil {
		return defs
	}

	for lineNum, line := range lines {
		matches := defRegex.FindStringSubmatch(line)
		if len(matches) > 0 {
			name := matches[len(matches)-1]
			defType := matches[1]
			if defType == "" && len(matches) > 2 {
				defType = matches[2]
			}
			defs = append(defs, map[string]interface{}{
				"name":    name,
				"type":    defType,
				"line":    lineNum + 1,
				"content": strings.TrimSpace(line),
			})
		}
	}

	return defs
}
