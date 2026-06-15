//go:build ignore
// +build ignore

package tools

/**
 * @file deepcontext.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of DeepContext MCP tools.
 * Replaces `deepcontext` (npx @wildcard-ai/deepcontext@latest) entry in mcp.json.
 *
 * DeepContext provides deep code understanding and context extraction:
 * - Analyzes codebase architecture and patterns
 * - Extracts semantic context from code
 * - Generates architectural summaries
 * - Identifies dependencies and relationships
 *
 * Improvements over original:
 * - No npx/Node dependency.
 * - Go-native file walking and AST analysis.
 * - Supports: analyze_codebase, get_context, find_patterns, summarize_architecture.
 */

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type CodebaseAnalysis struct {
	Path           string                 `json:"path"`
	TotalFiles     int                    `json:"total_files"`
	CodeFiles      int                    `json:"code_files"`
	TotalLines     int                    `json:"total_lines"`
	Languages      map[string]int         `json:"languages"`
	DirectoryTree  map[string]int         `json:"directory_counts"`
	Patterns       []PatternFinding       `json:"patterns,omitempty"`
	Architecture   string                 `json:"architecture_summary,omitempty"`
}

type PatternFinding struct {
	Pattern     string `json:"pattern"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	Description string `json:"description"`
}

// HandleDeepContextAnalyzeCodebase performs deep analysis of a codebase.
// Tool: deepcontext_analyze
func HandleDeepContextAnalyzeCodebase(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ := getString(args, "path", "directory")
	if path == "" {
		path = "."
	}

	analysis := &CodebaseAnalysis{
		Path:      path,
		Languages: make(map[string]int),
		DirectoryTree: make(map[string]int),
	}

	filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		// Skip hidden and vendor dirs
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") && name != "." && name != ".." {
				return filepath.SkipDir
			}
			if name == "node_modules" || name == "vendor" || name == ".git" ||
				name == "dist" || name == "build" || name == "__pycache__" ||
				name == ".next" || name == "target" {
				return filepath.SkipDir
			}
			return nil
		}

		analysis.TotalFiles++
		lang := detectLanguage(p)

		// Count code files
		codeExts := map[string]bool{
			".go": true, ".py": true, ".js": true, ".ts": true, ".tsx": true,
			".jsx": true, ".rs": true, ".java": true, ".cpp": true, ".c": true,
			".h": true, ".cs": true, ".rb": true, ".php": true, ".swift": true,
			".kt": true, ".lua": true, ".sh": true, ".sql": true,
			".html": true, ".css": true, ".scss": true,
		}
		ext := strings.ToLower(filepath.Ext(p))
		if codeExts[ext] {
			analysis.CodeFiles++
			analysis.Languages[lang]++

			data, e := os.ReadFile(p)
			if e == nil {
				analysis.TotalLines += len(strings.Split(string(data), "\n"))
			}
		}

		// Track directory distribution
		dir := filepath.Dir(p)
		rel, _ := filepath.Rel(path, dir)
		if rel == "." {
			rel = "root"
		}
		analysis.DirectoryTree[rel]++

		return nil
	})

	out, _ := json.MarshalIndent(analysis, "", "  ")
	return ok(string(out))
}

// HandleDeepContextGetContext extracts semantic context from specific code.
// Tool: deepcontext_get_context
func HandleDeepContextGetContext(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filePath, _ := getString(args, "file_path", "path")
	if filePath == "" {
		return err("file_path parameter is required")
	}

	data, e := os.ReadFile(filePath)
	if e != nil {
		return err(fmt.Sprintf("Failed to read file: %v", e))
	}

	content := string(data)
	lang := detectLanguage(filePath)
	lines := strings.Split(content, "\n")

	// Extract definitions, imports, exports
	defs := extractDefinitions(content, lang)
	imports := []string{}
	exports := []string{}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "import ") || strings.HasPrefix(trimmed, "from ") ||
			strings.HasPrefix(trimmed, "require(") || strings.HasPrefix(trimmed, "#include") ||
			strings.HasPrefix(trimmed, "use ") || strings.HasPrefix(trimmed, "package ") {
			imports = append(imports, trimmed)
		}
		if strings.HasPrefix(trimmed, "export ") || strings.HasPrefix(trimmed, "module.exports") ||
			strings.HasPrefix(trimmed, "pub ") {
			exports = append(exports, trimmed)
		}
	}

	result := map[string]interface{}{
		"file":          filePath,
		"language":      lang,
		"line_count":    len(lines),
		"definitions":   defs,
		"imports":       imports,
		"exports":       exports,
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(out))
}

// HandleDeepContextFindPatterns identifies patterns in code.
// Tool: deepcontext_find_patterns
func HandleDeepContextFindPatterns(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ := getString(args, "path", "directory")
	if path == "" {
		path = "."
	}

	patternTypes, _ := getString(args, "pattern_types")
	if patternTypes == "" {
		patternTypes = "all"
	}

	var findings []PatternFinding

	filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		if strings.Contains(p, "node_modules") || strings.Contains(p, ".git") ||
			strings.Contains(p, "vendor") {
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

		data, e := os.ReadFile(p)
		if e != nil {
			return nil
		}

		rel, _ := filepath.Rel(path, p)
		content := string(data)
		lines := strings.Split(content, "\n")

		// Detect patterns
		for lineNum, line := range lines {
			trimmed := strings.TrimSpace(line)

			// Singleton pattern
			if (patternTypes == "all" || strings.Contains(patternTypes, "singleton")) &&
				strings.Contains(trimmed, "getInstance") {
				findings = append(findings, PatternFinding{
					Pattern:     "singleton",
					File:        rel,
					Line:        lineNum + 1,
					Description: "Singleton pattern detected",
				})
			}

			// Factory pattern
			if (patternTypes == "all" || strings.Contains(patternTypes, "factory")) &&
				strings.Contains(trimmed, "Factory") || strings.Contains(trimmed, "create") && strings.Contains(trimmed, "new") {
				findings = append(findings, PatternFinding{
					Pattern:     "factory",
					File:        rel,
					Line:        lineNum + 1,
					Description: "Factory pattern detected",
				})
			}

			// Observer pattern
			if (patternTypes == "all" || strings.Contains(patternTypes, "observer")) &&
				(strings.Contains(trimmed, "subscribe") || strings.Contains(trimmed, "addEventListener") ||
					strings.Contains(trimmed, "on(") || strings.Contains(trimmed, "Emit")) {
				findings = append(findings, PatternFinding{
					Pattern:     "observer",
					File:        rel,
					Line:        lineNum + 1,
					Description: "Observer/event pattern detected",
				})
			}

			// Error handling patterns
			if (patternTypes == "all" || strings.Contains(patternTypes, "error")) &&
				strings.Contains(trimmed, "try") && strings.Contains(trimmed, "catch") {
				findings = append(findings, PatternFinding{
					Pattern:     "error_handling",
					File:        rel,
					Line:        lineNum + 1,
					Description: "Try-catch error handling pattern",
				})
			}
		}

		return nil
	})

	if len(findings) == 0 {
		return ok("No patterns found matching: " + patternTypes)
	}

	// Limit output
	if len(findings) > 50 {
		findings = findings[:50]
	}

	out, _ := json.MarshalIndent(findings, "", "  ")
	return ok(string(out))
}

// HandleDeepContextSummarizeArchitecture generates an architecture summary.
// Tool: deepcontext_summarize_architecture
func HandleDeepContextSummarizeArchitecture(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ := getString(args, "path", "directory")
	if path == "" {
		path = "."
	}

	// Gather structure info
	type DirInfo struct {
		Name        string
		FileCount   int
		CodeCount   int
		SubDirs     []string
		HasGoMod    bool
		HasPkgJson  bool
	}

	dirs := map[string]*DirInfo{}

	filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" ||
				name == ".git" || name == "dist" || name == "build" {
				return filepath.SkipDir
			}

			rel, _ := filepath.Rel(path, p)
			if _, ok := dirs[rel]; !ok {
				dirs[rel] = &DirInfo{Name: rel}
			}
			return nil
		}

		// Count files per directory
		dir := filepath.Dir(p)
		rel, _ := filepath.Rel(path, dir)
		if _, ok := dirs[rel]; !ok {
			dirs[rel] = &DirInfo{Name: rel}
		}
		dirs[rel].FileCount++

		ext := strings.ToLower(filepath.Ext(p))
		codeExts := map[string]bool{
			".go": true, ".py": true, ".js": true, ".ts": true, ".tsx": true,
			".rs": true, ".java": true, ".cpp": true, ".c": true,
		}
		if codeExts[ext] {
			dirs[rel].CodeCount++
		}

		if filepath.Base(p) == "go.mod" {
			dirs[rel].HasGoMod = true
		}
		if filepath.Base(p) == "package.json" {
			dirs[rel].HasPkgJson = true
		}

		return nil
	})

	// Build summary
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Architecture Summary: %s\n\n", path))
	sb.WriteString(fmt.Sprintf("Total directories analyzed: %d\n\n", len(dirs)))

	for dirName, info := range dirs {
		if info.CodeCount > 0 || info.HasGoMod || info.HasPkgJson {
			sb.WriteString(fmt.Sprintf("## %s\n", dirName))
			sb.WriteString(fmt.Sprintf("- Files: %d (Code: %d)\n", info.FileCount, info.CodeCount))
			if info.HasGoMod {
				sb.WriteString("- **Go Module**\n")
			}
			if info.HasPkgJson {
				sb.WriteString("- **Node.js Package**\n")
			}
			sb.WriteString("\n")
		}
	}

	return ok(sb.String())
}
