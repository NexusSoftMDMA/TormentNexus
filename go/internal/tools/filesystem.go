package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// HandleReadTextFile reads the contents of a file, supporting optional head or tail line limiting.
func HandleReadTextFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
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

	head := getInt(args, "head")
	tail := getInt(args, "tail")

	if head > 0 && tail > 0 {
		return err("Cannot specify both head and tail parameters simultaneously")
	}

	if head > 0 {
		if head > len(lines) {
			head = len(lines)
		}
		return ok(strings.Join(lines[:head], "\n"))
	}

	if tail > 0 {
		if tail > len(lines) {
			tail = len(lines)
		}
		start := len(lines) - tail
		return ok(strings.Join(lines[start:], "\n"))
	}

	return ok(content)
}

// HandleCreateDirectory creates a directory recursively.
func HandleCreateDirectory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ := getString(args, "path")
	if path == "" {
		return err("path parameter is required")
	}

	errMkdir := os.MkdirAll(path, 0755)
	if errMkdir != nil {
		return err(fmt.Sprintf("Failed to create directory: %v", errMkdir))
	}

	return ok(fmt.Sprintf("Successfully created directory: %s", path))
}

// HandleListDirectory lists directory contents (files and subdirectories).
func HandleListDirectory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ := getString(args, "path")
	if path == "" {
		return err("path parameter is required")
	}

	entries, e := os.ReadDir(path)
	if e != nil {
		return err(fmt.Sprintf("Failed to list directory: %v", e))
	}

	var results []string
	for _, entry := range entries {
		suffix := ""
		if entry.IsDir() {
			suffix = "/"
		}
		results = append(results, entry.Name()+suffix)
	}

	return ok(strings.Join(results, "\n"))
}

// FileEntry represents metadata about an entry in a directory list.
type FileEntry struct {
	Name    string `json:"name"`
	IsDir   bool   `json:"is_dir"`
	Size    int64  `json:"size"`
	ModTime string `json:"mod_time"`
}

// HandleListDirectoryWithSizes lists directory contents with sizes and sorts them.
func HandleListDirectoryWithSizes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ := getString(args, "path")
	if path == "" {
		return err("path parameter is required")
	}

	sortBy, _ := getString(args, "sortBy")
	if sortBy == "" {
		sortBy = "name"
	}

	entries, e := os.ReadDir(path)
	if e != nil {
		return err(fmt.Sprintf("Failed to list directory: %v", e))
	}

	var files []FileEntry
	for _, entry := range entries {
		info, _ := entry.Info()
		var size int64
		var modTime time.Time
		if info != nil {
			size = info.Size()
			modTime = info.ModTime()
		}

		files = append(files, FileEntry{
			Name:    entry.Name(),
			IsDir:   entry.IsDir(),
			Size:    size,
			ModTime: modTime.Format(time.RFC3339),
		})
	}

	// Sort logic
	if sortBy == "size" {
		sort.Slice(files, func(i, j int) bool {
			return files[i].Size < files[j].Size
		})
	} else {
		sort.Slice(files, func(i, j int) bool {
			return files[i].Name < files[j].Name
		})
	}

	b, errJson := json.Marshal(files)
	if errJson != nil {
		return err(fmt.Sprintf("Failed to format results: %v", errJson))
	}

	return ok(string(b))
}

// HandleDirectoryTree generates a visual structure of directories.
func HandleDirectoryTree(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ := getString(args, "path")
	if path == "" {
		return err("path parameter is required")
	}

	var excludePatterns []string
	if excludeVal, exists := args["excludePatterns"]; exists {
		if rawArray, ok := excludeVal.([]interface{}); ok {
			for _, item := range rawArray {
				if s, okS := item.(string); okS {
					excludePatterns = append(excludePatterns, s)
				}
			}
		}
	}

	var lines []string
	errWalk := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, errRel := filepath.Rel(path, p)
		if errRel != nil {
			return errRel
		}

		if rel == "." {
			return nil
		}

		name := d.Name()
		// Apply exclude filters
		for _, pat := range excludePatterns {
			if matched, _ := filepath.Match(pat, name); matched {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.Contains(rel, pat) {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		depth := len(strings.Split(rel, string(filepath.Separator)))
		indent := strings.Repeat("  ", depth-1)
		suffix := ""
		if d.IsDir() {
			suffix = "/"
		}
		lines = append(lines, fmt.Sprintf("%s%s%s", indent, name, suffix))
		return nil
	})

	if errWalk != nil {
		return err(fmt.Sprintf("Failed to generate tree: %v", errWalk))
	}

	return ok(strings.Join(lines, "\n"))
}

// HandleMoveFile moves or renames a file/directory.
func HandleMoveFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	source, _ := getString(args, "source")
	if source == "" {
		return err("source parameter is required")
	}

	destination, _ := getString(args, "destination")
	if destination == "" {
		return err("destination parameter is required")
	}

	// Ensure destination directory path exists
	destDir := filepath.Dir(destination)
	if e := os.MkdirAll(destDir, 0755); e != nil {
		return err(fmt.Sprintf("Failed to create destination directories: %v", e))
	}

	errRename := os.Rename(source, destination)
	if errRename != nil {
		return err(fmt.Sprintf("Failed to move/rename: %v", errRename))
	}

	return ok(fmt.Sprintf("Successfully moved %s to %s", source, destination))
}

// HandleGetFileInfo returns metadata for a file or directory.
func HandleGetFileInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ := getString(args, "path")
	if path == "" {
		return err("path parameter is required")
	}

	info, errStat := os.Stat(path)
	if errStat != nil {
		return err(fmt.Sprintf("Failed to get file info: %v", errStat))
	}

	meta := map[string]interface{}{
		"name":         info.Name(),
		"size":         info.Size(),
		"mode":         info.Mode().String(),
		"is_directory": info.IsDir(),
		"mod_time":     info.ModTime().Format(time.RFC3339),
	}

	b, errJson := json.Marshal(meta)
	if errJson != nil {
		return err(fmt.Sprintf("Failed to format results: %v", errJson))
	}

	return ok(string(b))
}

// HandleSearchFiles searches for files matching a glob pattern.
func HandleSearchFiles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ := getString(args, "path")
	if path == "" {
		return err("path parameter is required")
	}

	pattern, _ := getString(args, "pattern")
	if pattern == "" {
		return err("pattern parameter is required")
	}

	var excludePatterns []string
	if excludeVal, exists := args["excludePatterns"]; exists {
		if rawArray, ok := excludeVal.([]interface{}); ok {
			for _, item := range rawArray {
				if s, okS := item.(string); okS {
					excludePatterns = append(excludePatterns, s)
				}
			}
		}
	}

	var matches []string
	errWalk := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, errRel := filepath.Rel(path, p)
		if errRel != nil {
			return errRel
		}

		if rel == "." {
			return nil
		}

		name := d.Name()
		// Apply exclude filters
		for _, pat := range excludePatterns {
			if matched, _ := filepath.Match(pat, name); matched {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.Contains(rel, pat) {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		if matched, _ := filepath.Match(pattern, name); matched {
			matches = append(matches, rel)
		}

		return nil
	})

	if errWalk != nil {
		return err(fmt.Sprintf("Failed to search files: %v", errWalk))
	}

	return ok(strings.Join(matches, "\n"))
}
