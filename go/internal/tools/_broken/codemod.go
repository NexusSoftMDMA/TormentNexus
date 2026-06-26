package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// HandleApply applies a regex-based replacement to a single file.
func HandleApply(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	pattern, _ :=getString(args, "pattern")
	replacement, _ :=getString(args, "replacement")
	dryRun, _ :=getBool(args, "dry_run")

	content, readErr := os.ReadFile(path)
	if readErr != nil {
		return err(readErr.Error())
}

	re, reErr := regexp.Compile(pattern)
	if reErr != nil {
		return err(reErr.Error())
}

	newContent := re.ReplaceAllString(string(content), replacement)

	if dryRun {
		return ok(fmt.Sprintf("Preview for %s:\n%s", path, newContent))
}

	writeErr := os.WriteFile(path, []byte(newContent), 0644)
	if writeErr != nil {
		return err(writeErr.Error())
}

	return ok(fmt.Sprintf("Successfully modified %s", path))
}

// HandleBatch applies a regex-based replacement to multiple files matching a glob pattern.
func HandleBatch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	glob, _ :=getString(args, "glob")
	pattern, _ :=getString(args, "pattern")
	replacement, _ :=getString(args, "replacement")

	re, reErr := regexp.Compile(pattern)
	if reErr != nil {
		return err(reErr.Error())
}

	files, globErr := filepath.Glob(glob)
	if globErr != nil {
		return err(globErr.Error())
}

	if len(files) == 0 {
		return ok("No files found matching pattern: " + glob)
}

	var results []string
	for _, f := range files {
		select {
		case <-ctx.Done():
			return err("operation cancelled")
}
		default:
		}

		content, readErr := os.ReadFile(f)
		if readErr != nil {
			results = append(results, fmt.Sprintf("Error reading %s: %v", f, readErr))
			continue
		}

		newContent := re.ReplaceAllString(string(content), replacement)
		writeErr := os.WriteFile(f, []byte(newContent), 0644)
		if writeErr != nil {
			results = append(results, fmt.Sprintf("Error writing %s: %v", f, writeErr))
			continue
		}

		results = append(results, fmt.Sprintf("Modified %s", f))

	return ok(strings.Join(results, "\n"))
}

// HandleList lists files matching a glob pattern.
func HandleList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	glob, _ :=getString(args, "glob")

	files, globErr := filepath.Glob(glob)
	if globErr != nil {
		return err(globErr.Error())
}

	if len(files) == 0 {
		return ok("No files found matching pattern: " + glob)
}

	return ok(strings.Join(files, "\n"))
}