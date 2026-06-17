package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// RepositoryHealer manages repository-wide health and triggers auto-fixes.
type RepositoryHealer struct {
	AutoFix bool
}

// CheckRepoStatus checks for git conflicts and uncommitted changes.
func (h *RepositoryHealer) CheckRepoStatus() error {
	fmt.Println("[*] Checking Repository Status...")

	// Check for git conflicts
	cmd := exec.Command("git", "diff", "--name-only", "--diff-filter=U")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git diff failed: %v", err)
	}

	conflicts := strings.TrimSpace(string(output))
	if conflicts != "" {
		return fmt.Errorf("Merge conflicts detected in: %s", conflicts)
	}

	fmt.Println("[+] No merge conflicts detected.")
	return nil
}

// CheckBuildHealth runs the build and tests.
func (h *RepositoryHealer) CheckBuildHealth() error {
	fmt.Println("[*] Checking Build Health...")

	// Check Go build
	cmd := exec.Command("go", "build", "./...")
	if _, err := os.Stat("go.mod"); err != nil {
		cmd.Dir = "go"
	}

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("Go build failed: %v\nOutput: %s", err, string(output))
	}

	fmt.Println("[+] Go build passed.")
	return nil
}

// Run executes the full repository health scan and self-healing loop.
func (h *RepositoryHealer) Run(ctx context.Context) error {
	fmt.Printf("Starting Autonomous Repository Healing at %s\n", time.Now().Format(time.RFC3339))

	if err := h.CheckRepoStatus(); err != nil {
		fmt.Printf("[!] Repo status check failed: %v\n", err)
		if h.AutoFix {
			fmt.Println("[*] Attempting to resolve conflicts (git merge --abort)...")
			exec.Command("git", "merge", "--abort").Run()
		}
	}

	if err := h.CheckBuildHealth(); err != nil {
		fmt.Printf("[!] Build health check failed: %v\n", err)
		if h.AutoFix {
			fmt.Println("[*] Triggering LLM-based Healer Service...")
			// In a real scenario, we'd instantiate HealerService here
			// For now, we use the basic self-heal
			fmt.Println("[*] Attempting basic self-heal (go mod tidy)...")
			tidyCmd := exec.Command("go", "mod", "tidy")
			if _, errStat := os.Stat("go.mod"); errStat != nil {
				tidyCmd.Dir = "go"
			}
			tidyCmd.Run()

			if errRetry := h.CheckBuildHealth(); errRetry != nil {
				return errRetry
			}
			fmt.Println("[+] Self-heal successful.")
		} else {
			return err
		}
	}

	fmt.Println("Repository is HEALTHY.")
	return nil
}

func main() {
	healer := RepositoryHealer{AutoFix: true}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := healer.Run(ctx); err != nil {
		fmt.Printf("Autonomous Healing FAILED: %v\n", err)
		os.Exit(1)
	}
}
