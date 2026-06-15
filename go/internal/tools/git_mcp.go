//go:build ignore
// +build ignore

package tools

/**
 * @file git_mcp.go
 * @module go/internal/tools
 *
 * WHAT: Native Go implementation of Git MCP — git repository operations.
 * Replaces: git-mcp-server
 *
 * Provides git repository management: status, log, diff, branches, commits.
 * All operations are local to the current working directory.
 *
 * Tools:
 *  - git_status — show working tree status
 *  - git_log — show commit log
 *  - git_diff — show diff (working tree or between commits)
 *  - git_branches — list branches
 *  - git_show — show commit details
 *  - git_blame — show file blame
 *  - git_commit — create a commit
 *  - git_checkout — switch branches
 */

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func gitExec(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	out, e := cmd.CombinedOutput()
	if e != nil {
		return "", fmt.Errorf("git %s failed: %v\n%s", strings.Join(args, " "), e, string(out))
	}
	return string(out), nil
}

// HandleGitStatus shows the working tree status.
func HandleGitStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	short := ""
	if getBool(args, "short", "porcelain") {
		short = "--short"
	}
	result, e := gitExec(ctx, "status", short)
	if e != nil {
		return err(fmt.Sprintf("git status failed: %v", e))
	}
	return ok(result)
}

// HandleGitLog shows commit history.
func HandleGitLog(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	count := getInt(args, "count", "limit", "n")
	if count <= 0 || count > 100 {
		count = 10
	}
	format, _ := getString(args, "format")
	if format == "" {
		format = "%h %ad %an: %s"
	}
	result, e := gitExec(ctx, "log", fmt.Sprintf("-%d", count), fmt.Sprintf("--format=%s", format), "--date=short")
	if e != nil {
		return err(fmt.Sprintf("git log failed: %v", e))
	}
	return ok(result)
}

// HandleGitDiff shows changes.
func HandleGitDiff(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cached := getBool(args, "cached", "staged")
	file, _ := getString(args, "file", "path")

	gitArgs := []string{"diff"}
	if cached {
		gitArgs = append(gitArgs, "--cached")
	}
	if file != "" {
		gitArgs = append(gitArgs, "--", file)
	}

	result, e := gitExec(ctx, gitArgs...)
	if e != nil {
		return err(fmt.Sprintf("git diff failed: %v", e))
	}
	return ok(result)
}

// HandleGitBranches lists branches.
func HandleGitBranches(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	all := getBool(args, "all")
	gitArgs := []string{"branch"}
	if all {
		gitArgs = append(gitArgs, "-a")
	}
	result, e := gitExec(ctx, gitArgs...)
	if e != nil {
		return err(fmt.Sprintf("git branch failed: %v", e))
	}
	return ok(result)
}

// HandleGitShow shows a commit or object.
func HandleGitShow(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ref, _ := getString(args, "ref", "commit", "hash")
	if ref == "" {
		ref = "HEAD"
	}
	result, e := gitExec(ctx, "show", "--stat", "--format=fuller", ref)
	if e != nil {
		return err(fmt.Sprintf("git show failed: %v", e))
	}
	return ok(result)
}

// HandleGitBlame shows file blame.
func HandleGitBlame(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	file, _ := getString(args, "file", "path")
	if file == "" {
		return err("file is required")
	}
	result, e := gitExec(ctx, "blame", file)
	if e != nil {
		return err(fmt.Sprintf("git blame failed: %v", e))
	}
	return ok(result)
}

// HandleGitCommit creates a commit.
func HandleGitCommit(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ := getString(args, "message", "m", "msg")
	if message == "" {
		return err("commit message is required")
	}

	gitArgs := []string{"commit"}
	if getBool(args, "all", "a") {
		gitArgs = append(gitArgs, "-a")
	}
	if getBool(args, "allow_empty") {
		gitArgs = append(gitArgs, "--allow-empty")
	}
	gitArgs = append(gitArgs, "-m", message)

	result, e := gitExec(ctx, gitArgs...)
	if e != nil {
		return err(fmt.Sprintf("git commit failed: %v", e))
	}
	return ok(result)
}

// HandleGitCheckout switches branches.
func HandleGitCheckout(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	branch, _ := getString(args, "branch", "name", "b")
	if branch == "" {
		return err("branch name is required")
	}

	gitArgs := []string{"checkout"}
	if getBool(args, "create", "new", "c") {
		gitArgs = append(gitArgs, "-b")
	}
	gitArgs = append(gitArgs, branch)

	result, e := gitExec(ctx, gitArgs...)
	if e != nil {
		return err(fmt.Sprintf("git checkout failed: %v", e))
	}
	return ok(result)
}
