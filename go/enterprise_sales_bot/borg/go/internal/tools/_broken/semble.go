package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"time"
)

// HandleSearch runs `semble search` with the provided arguments.
// Expected args:
//   - query (string): natural language query.
//   - path (string, optional): repository path; defaults to current directory.
//   - top_k (int, optional): number of results; defaults to 10.
func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("missing required argument: query")
}

	path, _ :=getString(args, "path")
	if path == "" {
		path = "."
	}
	topK, _ :=getInt(args, "top_k")
	if topK <= 0 {
		topK = 10
	}

	cmdArgs := []string{"search", query, path, "--top-k", strconv.Itoa(topK)}
	execCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "semble", cmdArgs...)
	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return err(fmt.Sprintf("semble search failed: %v, output: %s", execErr, string(output)))
}

	return ok(string(output))
}

// HandleFindRelated runs `semble find-related` to locate code related to a given file/line.
// Expected args:
//   - file_path (string): path to the source file.
//   - line (int): line number within the file.
//   - repo_path (string, optional): repository root; defaults to current directory.
func HandleFindRelated(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filePath, _ :=getString(args, "file_path")
	if filePath == "" {
		return err("missing required argument: file_path")
}

	line, _ :=getInt(args, "line")
	if line <= 0 {
		return err("missing or invalid required argument: line")
}

	repoPath, _ :=getString(args, "repo_path")
	if repoPath == "" {
		repoPath = "."
	}

	cmdArgs := []string{"find-related", filePath, strconv.Itoa(line), repoPath}
	execCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "semble", cmdArgs...)
	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return err(fmt.Sprintf("semble find-related failed: %v, output: %s", execErr, string(output)))
}

	return ok(string(output))
}

// HandleInstall runs `semble install` to set up the Semble tool.
// No arguments are required.
func HandleInstall(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "semble", "install")
	output, execErr := cmd.CombinedOutput()
	if execErr != nil {
		return err(fmt.Sprintf("semble install failed: %v, output: %s", execErr, string(output)))
}

	return ok(string(output))
}