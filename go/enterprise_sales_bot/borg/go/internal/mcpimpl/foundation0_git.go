package mcpimpl

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func HandleListBranches(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	repoPath, _ :=getString(args, "repo_path")
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "branch", "--format=%(refname:short)")
	var out bytes.Buffer
	cmd.Stdout = &out
	if e := cmd.Run(); e != nil {
		return err(fmt.Sprintf("failed to list branches: %v", e))
}

	branches := strings.Split(strings.TrimSpace(out.String()), "\n")
	return success(fmt.Sprintf("Branches (%d): %s", len(branches), strings.Join(branches, ", ")))
}

func HandleListCommits(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	repoPath, _ :=getString(args, "repo_path")
	count, _ :=getInt(args, "count")
	if count <= 0 {
		count = 10
	}
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "log", fmt.Sprintf("--max-count=%d", count), "--oneline")
	var out bytes.Buffer
	cmd.Stdout = &out
	if e := cmd.Run(); e != nil {
		return err(fmt.Sprintf("failed to list commits: %v", e))
}

	commits := strings.Split(strings.TrimSpace(out.String()), "\n")
	return success(fmt.Sprintf("Recent commits (%d): %s", len(commits), strings.Join(commits, "\n")))
}