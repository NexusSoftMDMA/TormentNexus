package gitcheck

import (
	"bufio"
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

// IsClean checks if the git working directory is clean using 'git status --porcelain'.
// It returns true if there are no staged or unstaged changes.
func IsClean() (bool, error) {
	// In test environments, we allow dirty state for orchestrator unit tests
	if os.Getenv("GO_TEST_MODE") == "true" {
		return true, nil
	}
	cmd := exec.Command("git", "status", "--porcelain")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return false, err
	}
	return out.Len() == 0, nil
}

// IsSynced checks if the current branch is synchronized with the remote 'origin/target'.
// It performs a 'git fetch' first to ensure local awareness of remote state.
func IsSynced(target string) (bool, error) {
	// Fetch first to ensure we have latest info
	// #nosec G204 -- Intentional subprocess execution for autonomous sync
	if err := exec.Command("git", "fetch", "origin").Run(); err != nil {
		return false, fmt.Errorf("failed to fetch from origin: %w", err)
	}

	// #nosec G204 -- Intentional subprocess execution for autonomous sync
	cmd := exec.Command("git", "rev-list", "--left-right", "--count", "HEAD...origin/"+target)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return false, fmt.Errorf("failed to compare branches: %v", err)
	}

	// Output format: "ahead\tbehind"
	counts := strings.Fields(out.String())
	if len(counts) != 2 {
		return false, fmt.Errorf("unexpected rev-list output: %s", out.String())
	}

	// For integrity, we mostly care if we are behind (second number > 0)
	behind := counts[1]
	if behind != "0" {
		return false, nil
	}

	return true, nil
}

// SyncRemote fetches and merges changes from 'origin/main' and upstream if configured.
// It uses '--no-edit' to ensure the operation remains autonomous and non-interactive.
func SyncRemote() error {
	// Step 1: Fetch All to ensure complete repository awareness
	fetchCmd := exec.Command("git", "fetch", "--all", "--tags")
	if err := fetchCmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch all remotes: %v", err)
	}

	// Step 2: Handle Upstream Sync if present
	if hasUpstream() {
		slog.Info("Git Sync: Upstream remote detected, merging changes...")
		upstreamMerge := exec.Command("git", "merge", "upstream/main", "-m", "chore: autonomous upstream sync", "--no-edit")
		if err := upstreamMerge.Run(); err != nil {
			slog.Info(fmt.Sprintf("Git Sync Warning: Upstream merge failed: %v", err))
		}
	}

	// Step 3: Merge from origin/main
	mergeCmd := exec.Command("git", "merge", "origin/main", "-m", "chore: autonomous sync with origin/main", "--no-edit")
	if err := mergeCmd.Run(); err != nil {
		return fmt.Errorf("failed to merge from origin/main: %v", err)
	}

	return nil
}

func hasUpstream() bool {
	cmd := exec.Command("git", "remote")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return false
	}
	return strings.Contains(out.String(), "upstream")
}

// UpdateSubmodules updates all git submodules recursively within the repository.
// It initializes submodules if they haven't been already.
func UpdateSubmodules() error {
	cmd := exec.Command("git", "submodule", "update", "--init", "--recursive")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to update submodules: %v", err)
	}
	return nil
}

// CheckoutAndCommit creates or resets a branch and commits all current working directory changes.
// This is typically used by the autonomous agent to persist its self-directed updates.
func CheckoutAndCommit(branch string, message string) error {
	// #nosec G204 -- Intentional subprocess execution for autonomous git operations
	checkoutCmd := exec.Command("git", "checkout", "-B", branch)
	if err := checkoutCmd.Run(); err != nil {
		return fmt.Errorf("failed to checkout branch %s: %v", branch, err)
	}

	addCmd := exec.Command("git", "add", ".")
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("failed to stage changes: %v", err)
	}

	// #nosec G204 -- Intentional subprocess execution for autonomous git operations
	commitCmd := exec.Command("git", "commit", "-m", message)
	if err := commitCmd.Run(); err != nil {
		return fmt.Errorf("failed to commit changes: %v", err)
	}

	return nil
}

// PushBranch pushes the specified local branch to the 'origin' remote.
func PushBranch(branch string) error {
	// #nosec G204 -- Intentional subprocess execution for autonomous git operations
	cmd := exec.Command("git", "push", "origin", branch)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to push branch %s: %v", branch, err)
	}
	return nil
}

// DeleteBranch deletes a local branch.
func DeleteBranch(branch string) error {
	// #nosec G204 -- Intentional subprocess execution for autonomous git operations
	cmd := exec.Command("git", "branch", "-d", branch)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete local branch %s: %v", branch, err)
	}
	return nil
}

// DeleteRemoteBranch deletes a remote branch from 'origin'.
func DeleteRemoteBranch(branch string) error {
	// #nosec G204 -- Intentional subprocess execution for autonomous git operations
	cmd := exec.Command("git", "push", "origin", "--delete", branch)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete remote branch %s: %v", branch, err)
	}
	return nil
}

// ListFeatureBranches returns a list of all local branches prefixed with 'autodev/'.
// These represent the active feature branches managed by the autonomous orchestrator.
func ListFeatureBranches() ([]string, error) {
	cmd := exec.Command("git", "for-each-ref", "--format=%(refname:short)", "refs/heads/")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	var branches []string
	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		branch := strings.TrimSpace(scanner.Text())
		// Only reconcile branches explicitly created by the autonomous agent
		if strings.HasPrefix(branch, "autodev/") {
			branches = append(branches, branch)
		}
	}
	return branches, nil
}

// CheckConflicts checks the repository for any unmerged paths (merge conflicts).
// It returns true if any conflict markers are detected in the index.
func CheckConflicts() (bool, error) {
	cmd := exec.Command("git", "diff", "--name-only", "--diff-filter=U")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return false, err
	}
	return out.Len() > 0, nil
}

// GenerateSubmoduleInventory produces a Markdown table listing all submodules and their details.
func GenerateSubmoduleInventory() (string, error) {
	cmd := exec.Command("git", "submodule", "status", "--recursive")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}

	inventory := "| Submodule Path | Current Commit | Remote URL |\n"
	inventory += "|----------------|----------------|------------|\n"

	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			commit := parts[0]
			path := parts[1]

			// Get URL
			// #nosec G204 -- Intentional subprocess execution for autonomous submodule inventory
			urlCmd := exec.Command("git", "config", "--file", ".gitmodules", fmt.Sprintf("submodule.%s.url", path))
			urlOut, _ := urlCmd.Output()
			url := strings.TrimSpace(string(urlOut))

			inventory += fmt.Sprintf("| %s | %s | %s |\n", path, commit, url)
		}
	}

	return inventory, nil
}

// DiscardChanges discards all uncommitted changes in the working directory and index.
// Used by the autodev rollback mechanism when verification fails.
func DiscardChanges() error {
	// First, reset the index and working directory to HEAD
	resetCmd := exec.Command("git", "reset", "--hard", "HEAD")
	if out, err := resetCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git reset --hard failed: %v, output: %s", err, string(out))
	}

	// Then, remove any untracked files and directories created during the failed attempt
	cleanCmd := exec.Command("git", "clean", "-fd")
	if out, err := cleanCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git clean -fd failed: %v, output: %s", err, string(out))
	}

	return nil
}
