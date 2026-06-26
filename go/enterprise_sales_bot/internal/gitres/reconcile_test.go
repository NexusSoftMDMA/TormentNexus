package gitres

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestReconcileBranches_Functional(t *testing.T) {
	// Create a temporary directory for the test repository
	tmpDir, err := os.MkdirTemp("", "gitres-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Helper to run git commands in the temp directory
	runGit := func(args ...string) error {
		cmd := exec.Command("git", args...)
		cmd.Dir = tmpDir
		return cmd.Run()
	}

	// Real helper that returns output
	gitOut := func(args ...string) (string, error) {
		cmd := exec.Command("git", args...)
		cmd.Dir = tmpDir
		out, err := cmd.CombinedOutput()
		return string(out), err
	}

	// 1. Initialize a mock repository
	if err := runGit("init"); err != nil {
		t.Fatalf("git init failed: %v", err)
	}
	if err := runGit("config", "user.email", "test@example.com"); err != nil {
		t.Fatalf("git config email failed: %v", err)
	}
	if err := runGit("config", "user.name", "Test User"); err != nil {
		t.Fatalf("git config name failed: %v", err)
	}

	// Create initial commit on main
	if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Test Repo"), 0644); err != nil {
		t.Fatalf("Failed to write README: %v", err)
	}
	_ = runGit("add", "README.md")
	_ = runGit("commit", "-m", "initial commit")
	_ = runGit("branch", "-M", "main")

	// 2. Create a feature branch with unique progress
	_ = runGit("checkout", "-b", "autodev/test-feature")
	if err := os.WriteFile(filepath.Join(tmpDir, "feature.txt"), []byte("feature content"), 0644); err != nil {
		t.Fatalf("Failed to write feature file: %v", err)
	}
	_ = runGit("add", "feature.txt")
	_ = runGit("commit", "-m", "feat: unique progress")

	// 3. Run reconciliation (mocking the environment)
	// We need to change the working directory for the duration of the test
	origDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change dir: %v", err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Errorf("Failed to restore working directory: %v", err)
		}
	}()

	if err := ReconcileBranches(); err != nil {
		t.Errorf("ReconcileBranches failed: %v", err)
	}

	// 4. Verify results
	// Main should now have the feature content
	if out, err := gitOut("ls-tree", "-r", "main", "--name-only"); err != nil {
		t.Errorf("Failed to list files in main: %v", err)
	} else if !containsString(out, "feature.txt") {
		t.Errorf("Forward merge failed: feature.txt not found in main. Output: %s", out)
	}

	// Feature branch should have README.md (reverse merge check, though it had it already)
	if out, err := gitOut("ls-tree", "-r", "autodev/test-feature", "--name-only"); err != nil {
		t.Errorf("Failed to list files in feature: %v", err)
	} else if !containsString(out, "README.md") {
		t.Errorf("Branch state invalid: README.md not found in feature branch")
	}
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
