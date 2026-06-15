package tools

import (
	"context"
	"fmt"
	"os/exec"
)

func HandleGetGitStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	repoPath, _ :=getString(args, "repo_path")
	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	if repoPath != "" {
		cmd.Dir = repoPath
	}
	output, e := cmd.CombinedOutput()
	if e != nil {
		return err(fmt.Sprintf("git status failed: %v", e))
}

	return success(string(output))
}

func HandleGetGitLog(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	repoPath, _ :=getString(args, "repo_path")
	count, _ :=getInt(args, "count")
	if count <= 0 {
		count = 10
	}
	cmd := exec.CommandContext(ctx, "git", "log", fmt.Sprintf("--max-count=%d", count), "--oneline")
	if repoPath != "" {
		cmd.Dir = repoPath
	}
	output, e := cmd.CombinedOutput()
	if e != nil {
		return err(fmt.Sprintf("git log failed: %v", e))
}

	return success(string(output))
}