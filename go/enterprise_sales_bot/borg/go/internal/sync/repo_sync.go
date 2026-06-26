package sync

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// RepositorySync handles automated dependency management and repository synchronization.
type RepositorySync struct {
	WorkspaceRoot string
}

// NewRepositorySync creates a new sync manager.
func NewRepositorySync(root string) *RepositorySync {
	return &RepositorySync{WorkspaceRoot: root}
}

// SyncDependencies handles cloning and updating project dependencies (submodules).
func (s *RepositorySync) SyncDependencies(ctx context.Context) error {
	fmt.Println("[*] Synchronizing dependencies...")

	// Ensure submodules are initialized and updated
	cmd := exec.CommandContext(ctx, "git", "submodule", "update", "--init", "--recursive", "--remote")
	cmd.Dir = s.WorkspaceRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("submodule sync failed: %v\nOutput: %s", err, string(output))
	}

	fmt.Println("[+] Submodules synchronized successfully.")
	return nil
}

// UpdateVersions ensures all dependency versions are aligned.
func (s *RepositorySync) UpdateVersions(ctx context.Context) error {
	fmt.Println("[*] Updating dependency versions...")

	// Example: Run pnpm install to sync TS dependencies
	if _, err := os.Stat(filepath.Join(s.WorkspaceRoot, "package.json")); err == nil {
		cmd := exec.CommandContext(ctx, "pnpm", "install", "--frozen-lockfile")
		cmd.Dir = s.WorkspaceRoot
		if output, err := cmd.CombinedOutput(); err != nil {
			fmt.Printf("[!] PNPM install warning: %v\n", err)
		} else {
			_ = output
			fmt.Println("[+] TS dependencies synchronized.")
		}
	}

	// Example: Run go mod tidy to sync Go dependencies
	goModPath := filepath.Join(s.WorkspaceRoot, "go", "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		cmd := exec.CommandContext(ctx, "go", "mod", "tidy")
		cmd.Dir = filepath.Dir(goModPath)
		if err := cmd.Run(); err != nil {
			fmt.Printf("[!] Go mod tidy failed: %v\n", err)
		} else {
			fmt.Println("[+] Go dependencies synchronized.")
		}
	}

	return nil
}

// CleanOrphans removes any orphaned submodules or directories.
func (s *RepositorySync) CleanOrphans(ctx context.Context) error {
	fmt.Println("[*] Cleaning orphaned dependencies...")

	// Basic cleanup logic
	cmd := exec.CommandContext(ctx, "git", "clean", "-fd")
	cmd.Dir = s.WorkspaceRoot
	_ = cmd.Run()

	return nil
}

// Run executes the full synchronization loop.
func (s *RepositorySync) Run(ctx context.Context) error {
	if err := s.SyncDependencies(ctx); err != nil {
		return err
	}
	if err := s.UpdateVersions(ctx); err != nil {
		return err
	}
	return s.CleanOrphans(ctx)
}
