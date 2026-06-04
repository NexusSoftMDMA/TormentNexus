package tools

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// HandleGitIngest implements the Go-native GitIngest tool logic comprehensively.
func HandleGitIngest(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	source, _ := getString(args, "source")
	if source == "" {
		return err("source parameter is required")
	}

	maxFileSize := getInt(args, "max_file_size", "maxFileSize")
	if maxFileSize <= 0 {
		maxFileSize = 10 * 1024 * 1024 // 10 MB default
	}

	includePatternsStr, _ := getString(args, "include_patterns", "includePatterns")
	excludePatternsStr, _ := getString(args, "exclude_patterns", "excludePatterns")
	branch, _ := getString(args, "branch")
	if branch == "" {
		branch = "main"
	}

	// Parse pattern strings (split by comma and trim whitespace)
	var includePatterns []string
	if includePatternsStr != "" {
		for _, p := range strings.Split(includePatternsStr, ",") {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				includePatterns = append(includePatterns, trimmed)
			}
		}
	}
	var excludePatterns []string
	if excludePatternsStr != "" {
		for _, p := range strings.Split(excludePatternsStr, ",") {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				excludePatterns = append(excludePatterns, trimmed)
			}
		}
	}

	isURL := strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") || strings.HasPrefix(source, "git@")

	var targetDir string
	if isURL {
		// Create temporary directory for cloning
		tempDir, e := os.MkdirTemp("", "gitingest-clone-*")
		if e != nil {
			return err(fmt.Sprintf("Failed to create temporary directory: %v", e))
		}
		defer os.RemoveAll(tempDir)

		// Run git clone
		cmdArgs := []string{"clone", "--depth", "1", "-b", branch, source, tempDir}
		cmd := exec.CommandContext(ctx, "git", cmdArgs...)
		if output, e := cmd.CombinedOutput(); e != nil {
			return err(fmt.Sprintf("Git clone failed: %v\nOutput: %s", e, string(output)))
		}
		targetDir = tempDir
	} else {
		// Verify local directory exists
		absPath, e := filepath.Abs(source)
		if e != nil {
			return err(fmt.Sprintf("Invalid local path: %v", e))
		}
		if stat, e := os.Stat(absPath); e != nil || !stat.IsDir() {
			return err(fmt.Sprintf("Local path does not exist or is not a directory: %s", absPath))
		}
		targetDir = absPath
	}

	// Walk directory and process files
	type fileInfo struct {
		relPath string
		content string
		size    int64
	}

	var files []fileInfo
	var fileTreeLines []string
	var totalSize int64
	var fileCount int

	errWalk := filepath.WalkDir(targetDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		rel, errRel := filepath.Rel(targetDir, path)
		if errRel != nil {
			return errRel
		}

		if rel == "." {
			return nil
		}

		// Skip git directory
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}

		// Generate tree representation
		depth := len(strings.Split(rel, string(filepath.Separator)))
		indent := strings.Repeat("  ", depth-1)
		name := d.Name()
		if d.IsDir() {
			name += "/"
		}
		fileTreeLines = append(fileTreeLines, fmt.Sprintf("%s%s", indent, name))

		if d.IsDir() {
			return nil
		}

		// Filter files by size
		info, errInfo := d.Info()
		if errInfo != nil {
			return nil
		}

		if info.Size() > int64(maxFileSize) {
			return nil // ignore large files
		}

		// Apply include/exclude pattern filters
		if len(includePatterns) > 0 {
			matched := false
			for _, pat := range includePatterns {
				if matchedGlob, _ := filepath.Match(pat, d.Name()); matchedGlob {
					matched = true
					break
				}
				if strings.Contains(rel, pat) {
					matched = true
					break
				}
			}
			if !matched {
				return nil
			}
		}

		if len(excludePatterns) > 0 {
			for _, pat := range excludePatterns {
				if matchedGlob, _ := filepath.Match(pat, d.Name()); matchedGlob {
					return nil
				}
				if strings.Contains(rel, pat) {
					return nil
				}
			}
		}

		// Read file contents
		data, errRead := os.ReadFile(path)
		if errRead != nil {
			return nil
		}

		// Basic check for text vs binary (null byte check)
		isBinary := false
		for i := 0; i < len(data) && i < 1024; i++ {
			if data[i] == 0 {
				isBinary = true
				break
			}
		}

		var fileContent string
		if isBinary {
			fileContent = "[Binary File]"
		} else {
			fileContent = string(data)
		}

		files = append(files, fileInfo{
			relPath: rel,
			content: fileContent,
			size:    info.Size(),
		})
		totalSize += info.Size()
		fileCount++
		return nil
	})

	if errWalk != nil {
		return err(fmt.Sprintf("Failed to walk directory: %v", errWalk))
	}

	// Format Summary
	summary := fmt.Sprintf("Repository Name: %s\nIngestion Date: %s\nTotal Files: %d\nTotal Size: %.2f KB",
		filepath.Base(source),
		time.Now().Format(time.RFC3339),
		fileCount,
		float64(totalSize)/1024.0,
	)

	// Format Tree
	tree := "Directory Structure:\n" + strings.Join(fileTreeLines, "\n")

	// Format Content
	var contentBuilder strings.Builder
	contentBuilder.WriteString("File Contents:\n")
	for _, f := range files {
		contentBuilder.WriteString(fmt.Sprintf("\n==================================================\nFile: %s (Size: %d bytes)\n==================================================\n", f.relPath, f.size))
		contentBuilder.WriteString(f.content)
		contentBuilder.WriteString("\n")
	}

	finalOutput := fmt.Sprintf("%s\n\n%s\n\n%s", summary, tree, contentBuilder.String())
	return ok(finalOutput)
}
