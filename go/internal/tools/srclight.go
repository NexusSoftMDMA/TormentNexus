//go:build ignore
// +build ignore

package tools

/**
 * @file srclight.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of Srclight — code indexing for AI agents.
 * Replaces: github.com/srclight/srclight
 *
 * Provides offline, incremental code indexing with support for multiple
 * languages. Indexes codebases and provides searchable symbol databases.
 *
 * Tools:
 *  - srclight_index — index a directory or file
 *  - srclight_search — search indexed symbols
 *  - srclight_status — get index status
 *  - srclight_list_languages — list supported languages
 */

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Srclight represents the code index
type srclight struct {
	db map[string]srclightEntry
}

type srclightEntry struct {
	File     string   `json:"file"`
	Language string   `json:"language"`
	Symbols  []string `json:"symbols"`
	Size     int64    `json:"size"`
}

func newSrclight() *srclight {
	return &srclight{db: make(map[string]srclightEntry)}
}

var srclightIndex *srclight

func getSrclight() *srclight {
	if srclightIndex == nil {
		srclightIndex = newSrclight()
	}
	return srclightIndex
}

var supportedLanguages = []string{
	"go", "python", "javascript", "typescript", "rust",
	"c", "cpp", "java", "ruby", "php", "swift", "kotlin",
}

func srclightDetectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go":
		return "go"
	case ".py":
		return "python"
	case ".js", ".mjs", ".cjs":
		return "javascript"
	case ".ts", ".tsx", ".mts":
		return "typescript"
	case ".rs":
		return "rust"
	case ".c", ".h":
		return "c"
	case ".cpp", ".cc", ".cxx", ".hpp":
		return "cpp"
	case ".java", ".kt", ".kts":
		return "java"
	case ".rb":
		return "ruby"
	case ".php":
		return "php"
	case ".swift":
		return "swift"
	default:
		return "unknown"
	}
}

// HandleSrclightIndex indexes a directory or file.
func HandleSrclightIndex(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	target, _ := getString(args, "path", "target", "dir")
	if target == "" {
		target = "."
	}
	recursive := getBool(args, "recursive", "r")
	if !recursive {
		recursive = true // default recursive
	}

	idx := getSrclight()

	filepath.Walk(target, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			if info != nil && info.IsDir() && p != target && !recursive {
				return filepath.SkipDir
			}
			return nil
		}

		lang := srclightDetectLanguage(p)
		if lang == "unknown" {
			return nil
		}

		data, e := os.ReadFile(p)
		if e != nil {
			return nil
		}

		// Extract simple symbols (function/type/class definitions)
		content := string(data)
		var symbols []string
		for _, line := range strings.Split(content, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "func ") {
				if name := extractSymbol(line, "func "); name != "" {
					symbols = append(symbols, "func:"+name)
				}
			} else if strings.HasPrefix(line, "type ") {
				if name := extractSymbol(line, "type "); name != "" {
					symbols = append(symbols, "type:"+name)
				}
			} else if strings.HasPrefix(line, "class ") || strings.HasPrefix(line, "interface ") {
				if name := extractSymbol(line, "class ", "interface "); name != "" {
					symbols = append(symbols, "type:"+name)
				}
			} else if strings.HasPrefix(line, "def ") {
				parts := strings.Fields(line)
				if len(parts) > 1 {
					name := strings.Split(parts[1], "(")[0]
					symbols = append(symbols, "func:"+name)
				}
			} else if strings.HasPrefix(line, "fn ") {
				parts := strings.Fields(line)
				if len(parts) > 1 {
					name := strings.Split(parts[1], "(")[0]
					symbols = append(symbols, "func:"+name)
				}
			}
		}

		idx.db[p] = srclightEntry{
			File:     p,
			Language: lang,
			Symbols:  symbols,
			Size:     info.Size(),
		}
		return nil
	})

	count := len(idx.db)
	return ok(fmt.Sprintf("Indexed %d files across %d languages", count, countLanguages(idx)))
}

func extractSymbol(line string, prefixes ...string) string {
	trimmed := strings.TrimSpace(line)
	for _, p := range prefixes {
		if strings.HasPrefix(trimmed, p) {
			rest := strings.TrimPrefix(trimmed, p)
			name := strings.Split(rest, "(")[0]
			name = strings.Split(name, "{")[0]
			name = strings.Split(name, " ")[0]
			return strings.TrimSpace(name)
		}
	}
	return ""
}

func countLanguages(idx *srclight) int {
	langs := make(map[string]bool)
	for _, e := range idx.db {
		langs[e.Language] = true
	}
	return len(langs)
}

// HandleSrclightSearch searches indexed symbols.
func HandleSrclightSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q", "symbol")
	if query == "" {
		return err("query is required")
	}
	language, _ := getString(args, "language", "lang")

	idx := getSrclight()
	var results []map[string]interface{}
	for path, entry := range idx.db {
		if language != "" && entry.Language != language {
			continue
		}
		for _, sym := range entry.Symbols {
			if strings.Contains(strings.ToLower(sym), strings.ToLower(query)) {
				results = append(results, map[string]interface{}{
					"file":     path,
					"symbol":   sym,
					"language": entry.Language,
				})
			}
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i]["file"].(string) < results[j]["file"].(string)
	})
	if results == nil {
		results = []map[string]interface{}{}
	}

	out, _ := json.MarshalIndent(results, "", "  ")
	return ok(string(out))
}

// HandleSrclightStatus returns index status.
func HandleSrclightStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	idx := getSrclight()
	langCount := make(map[string]int)
	totalSymbols := 0
	for _, e := range idx.db {
		langCount[e.Language]++
		totalSymbols += len(e.Symbols)
	}

	status := map[string]interface{}{
		"indexed_files":    len(idx.db),
		"total_symbols":    totalSymbols,
		"by_language":      langCount,
		"supported_languages": supportedLanguages,
	}
	out, _ := json.MarshalIndent(status, "", "  ")
	return ok(string(out))
}

// HandleSrclightListLanguages returns supported languages.
func HandleSrclightListLanguages(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	out, _ := json.MarshalIndent(map[string]interface{}{
		"languages": supportedLanguages,
		"count":     len(supportedLanguages),
	}, "", "  ")
	return ok(string(out))
}
