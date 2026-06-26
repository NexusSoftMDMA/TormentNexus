package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type WorktreeManager struct {
	rootDir string
}

func NewWorktreeManager(rootDir string) *WorktreeManager {
	return &WorktreeManager{rootDir: strings.TrimSpace(rootDir)}
}

func (m *WorktreeManager) CreateTaskEnvironment(taskID string) (string, error) {
	if strings.TrimSpace(m.rootDir) == "" {
		return "", fmt.Errorf("missing worktree root")
	}
	if !isGitRepository(m.rootDir) {
		return "", fmt.Errorf("worktree root is not a git repository: %s", m.rootDir)
	}
	if _, err := exec.LookPath("git"); err != nil {
		return "", fmt.Errorf("git not available on PATH")
	}

	branchName := "task/" + sanitizeTaskID(taskID)
	relativePath := filepath.Join(".tormentnexus", "worktrees", sanitizeTaskID(taskID))
	fullPath := filepath.Join(m.rootDir, relativePath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return "", err
	}

	args := []string{"worktree", "add"}
	if !branchExists(m.rootDir, branchName) {
		args = append(args, "-b", branchName, fullPath)
	} else {
		args = append(args, fullPath, branchName)
	}
	cmd := exec.CommandContext(context.Background(), "git", args...)
	cmd.Dir = m.rootDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s failed: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return fullPath, nil
}

func (m *WorktreeManager) CleanupTaskEnvironment(taskID string) error {
	if strings.TrimSpace(m.rootDir) == "" {
		return fmt.Errorf("missing worktree root")
	}
	fullPath := filepath.Join(m.rootDir, ".tormentnexus", "worktrees", sanitizeTaskID(taskID))
	cmd := exec.CommandContext(context.Background(), "git", "worktree", "remove", "--force", fullPath)
	cmd.Dir = m.rootDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree remove failed: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func branchExists(rootDir string, branchName string) bool {
	cmd := exec.CommandContext(context.Background(), "git", "show-ref", "--verify", "refs/heads/"+branchName)
	cmd.Dir = rootDir
	return cmd.Run() == nil
}

func isGitRepository(rootDir string) bool {
	if strings.TrimSpace(rootDir) == "" {
		return false
	}
	cmd := exec.CommandContext(context.Background(), "git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = rootDir
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "true"
}

func sanitizeTaskID(taskID string) string {
	trimmed := strings.TrimSpace(taskID)
	if trimmed == "" {
		return "session"
	}
	trimmed = strings.ReplaceAll(trimmed, "\\", "-")
	trimmed = strings.ReplaceAll(trimmed, "/", "-")
	trimmed = strings.ReplaceAll(trimmed, ":", "-")
	trimmed = strings.ReplaceAll(trimmed, " ", "-")
	return trimmed
}
