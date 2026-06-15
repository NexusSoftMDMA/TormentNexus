//go:build ignore
// +build ignore

package tools

/**
 * @file basic_memory.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of basic-memory (local markdown-based memory).
 * Replaces `basic-memory` (uvx basic-memory mcp) STDIO entry in mcp.json.
 *
 * Uses a local SQLite or file-based store for persistent memory notes.
 * Improvements over original:
 *  - No uvx/Python dependency.
 *  - Simple file-based memory with search, write, and read.
 *  - Compatible with basic-memory's entity/relation graph schema.
 */

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func basicMemoryDir() string {
	if d := os.Getenv("BASIC_MEMORY_DIR"); d != "" {
		return d
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".basic-memory")
}

// HandleBasicMemoryWrite writes a note to the basic-memory store.
// Tool: basic_memory_write
func HandleBasicMemoryWrite(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ := getString(args, "title", "name", "entity")
	if title == "" {
		return err("title parameter is required")
	}

	content, _ := getString(args, "content", "text", "body")
	if content == "" {
		return err("content parameter is required")
	}

	folder, _ := getString(args, "folder", "category")
	memDir := basicMemoryDir()

	if folder != "" {
		memDir = filepath.Join(memDir, folder)
	}

	if e := os.MkdirAll(memDir, 0755); e != nil {
		return err(fmt.Sprintf("Failed to create memory directory: %v", e))
	}

	// Sanitize filename
	filename := strings.ReplaceAll(title, " ", "_")
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")
	filename += ".md"

	filePath := filepath.Join(memDir, filename)

	// Build frontmatter + content
	now := time.Now().Format(time.RFC3339)
	fileContent := fmt.Sprintf("---\ntitle: %s\ncreated: %s\nupdated: %s\n---\n\n%s\n",
		title, now, now, content)

	if e := os.WriteFile(filePath, []byte(fileContent), 0644); e != nil {
		return err(fmt.Sprintf("Failed to write memory: %v", e))
	}

	return ok(fmt.Sprintf("Memory written: %s", filePath))
}

// HandleBasicMemoryRead reads a specific memory note.
// Tool: basic_memory_read
func HandleBasicMemoryRead(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ := getString(args, "title", "name", "entity")
	if title == "" {
		return err("title parameter is required")
	}

	folder, _ := getString(args, "folder", "category")
	memDir := basicMemoryDir()
	if folder != "" {
		memDir = filepath.Join(memDir, folder)
	}

	filename := strings.ReplaceAll(title, " ", "_") + ".md"
	filePath := filepath.Join(memDir, filename)

	data, e := os.ReadFile(filePath)
	if e != nil {
		// Try without sanitization
		filePath = filepath.Join(memDir, title+".md")
		data, e = os.ReadFile(filePath)
		if e != nil {
			return err(fmt.Sprintf("Memory not found: %s", title))
		}
	}

	return ok(string(data))
}

// HandleBasicMemorySearch searches memory notes by keyword.
// Tool: basic_memory_search
func HandleBasicMemorySearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ := getString(args, "query", "q", "search")
	if query == "" {
		return err("query parameter is required")
	}

	memDir := basicMemoryDir()
	lowerQuery := strings.ToLower(query)

	var matches []string
	err2 := filepath.Walk(memDir, func(path string, info os.FileInfo, e error) error {
		if e != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		data, e := os.ReadFile(path)
		if e != nil {
			return nil
		}

		if strings.Contains(strings.ToLower(string(data)), lowerQuery) {
			matches = append(matches, path)
		}
		return nil
	})

	if err2 != nil && !os.IsNotExist(err2) {
		return err(fmt.Sprintf("Search failed: %v", err2))
	}

	if len(matches) == 0 {
		return ok(fmt.Sprintf("No memories found matching: %s", query))
	}

	sort.Strings(matches)

	result := fmt.Sprintf("# Memory Search Results for: %s\n\nFound %d matching notes:\n\n", query, len(matches))
	for _, match := range matches {
		data, _ := os.ReadFile(match)
		relPath, _ := filepath.Rel(memDir, match)
		result += fmt.Sprintf("## %s\n\n```\n%s\n```\n\n", relPath, string(data))
	}
	return ok(result)
}

// HandleBasicMemoryList lists all memory notes.
// Tool: basic_memory_list
func HandleBasicMemoryList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	memDir := basicMemoryDir()
	folder, _ := getString(args, "folder", "category")
	if folder != "" {
		memDir = filepath.Join(memDir, folder)
	}

	var files []string
	err2 := filepath.Walk(memDir, func(path string, info os.FileInfo, e error) error {
		if e != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		relPath, _ := filepath.Rel(memDir, path)
		files = append(files, relPath)
		return nil
	})

	if err2 != nil && os.IsNotExist(err2) {
		return ok("No memories stored yet. Memory directory: " + memDir)
	}
	if err2 != nil {
		return err(fmt.Sprintf("Failed to list memories: %v", err2))
	}

	if len(files) == 0 {
		return ok("No memories found in: " + memDir)
	}

	sort.Strings(files)
	result := fmt.Sprintf("# Memory Store (%s)\n\n%d notes:\n\n", memDir, len(files))
	for _, f := range files {
		result += "- " + f + "\n"
	}
	return ok(result)
}

// HandleBasicMemoryDelete deletes a specific memory note.
// Tool: basic_memory_delete
func HandleBasicMemoryDelete(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ := getString(args, "title", "name")
	if title == "" {
		return err("title parameter is required")
	}

	folder, _ := getString(args, "folder", "category")
	memDir := basicMemoryDir()
	if folder != "" {
		memDir = filepath.Join(memDir, folder)
	}

	filename := strings.ReplaceAll(title, " ", "_") + ".md"
	filePath := filepath.Join(memDir, filename)

	if e := os.Remove(filePath); e != nil {
		return err(fmt.Sprintf("Failed to delete memory '%s': %v", title, e))
	}

	return ok("Memory deleted: " + title)
}
