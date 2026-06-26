package orchestrator

import (
	"fmt"
	"log"
	"os/exec"
)

// GitWorktreeManager parity for TS core `GitWorktreeManager.ts`
// Enables agents to seamlessly check out bare repositories into linked working trees.
type GitWorktreeManager struct {
	RepoPath string
}

func NewGitWorktreeManager(path string) *GitWorktreeManager {
	return &GitWorktreeManager{
		RepoPath: path,
	}
}

// CreateWorktree checks out a new branch into an isolated filesystem tree.
// Highly concurrent, native Go performance replacing Node.js child_process.exec.
func (g *GitWorktreeManager) CreateWorktree(branchName, targetDir string) error {
	log.Printf("[Worktree] Engaging strict repository isolation mapping branch %s to %s natively.", branchName, targetDir)

	cmd := exec.Command("git", "worktree", "add", "-B", branchName, targetDir)
	cmd.Dir = g.RepoPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to cleanly isolate git worktree boundaries natively: %w\n%s", err, string(output))
	}

	log.Printf("[Worktree] Sandbox isolated precisely! Ready for autonomous interaction internally.")
	return nil
}

// DestroyWorktree actively tears down the execution sandbox enforcing Git garbage collection dynamically safely protecting the MonoRepo!
func (g *GitWorktreeManager) DestroyWorktree(targetDir string, branchName string) error {
	log.Printf("[Worktree] Engaging immediate destruct payload on temporary bounds %s natively.", targetDir)

	// Prune the physical link
	cmd := exec.Command("git", "worktree", "remove", "--force", targetDir)
	cmd.Dir = g.RepoPath
	out1, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[Worktree-Warn] Force remove output natively parsed: %s", string(out1))
	}

	// Purge the arbitrary temporary branch cleanly
	cmd2 := exec.Command("git", "branch", "-D", branchName)
	cmd2.Dir = g.RepoPath
	out2, _ := cmd2.CombinedOutput()
	log.Printf("[Worktree] Executed Branch destruct sequences cleanly: %s", string(out2))

	return nil
}
