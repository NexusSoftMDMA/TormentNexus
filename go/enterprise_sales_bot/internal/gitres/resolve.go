package gitres

import (
	"bytes"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/robertpelloni/enterprise_sales_bot/internal/gitcheck"
)

// ResolveConflict performs a 'git merge' of the source branch into the current branch.
// It accepts an optional strategy (e.g., 'ours', 'theirs') to automatically resolve conflicts.
func ResolveConflict(source string, strategy string) error {
	args := []string{"merge", source, "--no-edit"}
	if strategy != "" {
		args = append(args, "-X", strategy)
	}

	// #nosec G204 -- This is a git automation bot; executing git commands with variable arguments is its primary function.
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If merge failed, attempt to find conflicting files before potential abort
		conflictCmd := exec.Command("git", "diff", "--name-only", "--diff-filter=U")
		conflicts, _ := conflictCmd.CombinedOutput()
		if len(conflicts) > 0 {
			slog.Info(fmt.Sprintf("Intelligent Merge: Conflicts detected in files:\n%s", string(conflicts)))
		}
		return fmt.Errorf("merge failed: %v, output: %s", err, string(output))
	}
	return nil
}

// AbortMerge resets the current merge state using 'git merge --abort'.
// This is used for recovery when an autonomous merge fails to resolve cleanly.
func AbortMerge() error {
	cmd := exec.Command("git", "merge", "--abort")
	return cmd.Run()
}

func hasUniqueProgress(branch string) bool {
	// #nosec G204 -- Intentional subprocess execution for autonomous git reconciliation
	cmd := exec.Command("git", "rev-list", "--count", "main.."+branch)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return false
	}
	count := strings.TrimSpace(out.String())
	return count != "0"
}

// ReconcileBranches implements the Dual-Direction Intelligent Merge Engine.
// It iterates through all 'autodev/' branches, attempting to forward-merge them into 'main'
// and reverse-merge 'main' back into each feature branch to prevent drift.
func ReconcileBranches() error {
	branches, err := gitcheck.ListFeatureBranches()
	if err != nil {
		return err
	}

	for _, branch := range branches {
		if !hasUniqueProgress(branch) {
			slog.Info(fmt.Sprintf("Intelligent Merge: Skipping %s, no unique progress.", branch))
			continue
		}

		slog.Info(fmt.Sprintf("Intelligent Merge: Reconciling branch: %s", branch))

		// 1. Forward Merge: Feature -> Main
		slog.Info(fmt.Sprintf("Intelligent Merge: Attempting Forward Merge (%s -> main)...", branch))
		if out, err := exec.Command("git", "checkout", "main").CombinedOutput(); err != nil {
			return fmt.Errorf("failed to checkout main: %v, output: %s", err, string(out))
		}
		if err := ResolveConflict(branch, "theirs"); err != nil {
			slog.Info(fmt.Sprintf("Intelligent Merge: Forward merge failed for %s: %v", branch, err))
			_ = AbortMerge()
		} else {
			slog.Info(fmt.Sprintf("Intelligent Merge: Successfully merged %s into main", branch))
		}

		// 2. Reverse Merge: Main -> Feature
		slog.Info(fmt.Sprintf("Intelligent Merge: Attempting Reverse Merge (main -> %s)...", branch))
		// #nosec G204 -- Intentional subprocess execution for autonomous git reconciliation
		if out, err := exec.Command("git", "checkout", branch).CombinedOutput(); err != nil {
			return fmt.Errorf("failed to checkout feature branch %s: %v, output: %s", branch, err, string(out))
		}
		// We use standard merge to catch drift and ensure metadata (TODO, VERSION) is preserved correctly
		if err := ResolveConflict("main", ""); err != nil {
			slog.Info(fmt.Sprintf("Intelligent Merge: Reverse merge failed for %s: %v", branch, err))
			_ = AbortMerge()
		} else {
			slog.Info(fmt.Sprintf("Intelligent Merge: Successfully reconciled %s with main", branch))
		}
	}

	// Switch back to main
	if out, err := exec.Command("git", "checkout", "main").CombinedOutput(); err != nil {
		return fmt.Errorf("failed to return to main: %v, output: %s", err, string(out))
	}

	// Step 3 of EXECUTIVE PROTOCOL: Finalize by pushing reconciled main branch
	slog.Info("Intelligent Merge: Finalizing reconciliation by pushing main to origin...")
	if err := gitcheck.PushBranch("main"); err != nil {
		slog.Info(fmt.Sprintf("Intelligent Merge Warning: Final push failed: %v", err))
		// We don't return error here to allow the cycle to continue locally
	}

	return nil
}
