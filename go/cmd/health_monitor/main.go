package main

import (
	"fmt"
	"os"
	"os/exec"
)

// HealthMonitor handles autonomous health checks and self-healing.
type HealthMonitor struct {
	AutoFix bool
}

// CheckBuild attempts to build the kernel to ensure code health.
func (h *HealthMonitor) CheckBuild() error {
	cmd := exec.Command("go", "build", "./...")
	if _, err := os.Stat("go.mod"); err != nil {
		cmd.Dir = "go"
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("build failed: %v\nOutput: %s", err, string(output))
	}
	return nil
}

// Run executes the health check suite.
func (h *HealthMonitor) Run() error {
	fmt.Println("Starting Autonomous Health Check...")

	if err := h.CheckBuild(); err != nil {
		fmt.Printf("[!] Build check failed: %v\n", err)
		if h.AutoFix {
			fmt.Println("[*] Attempting self-heal (go mod tidy)...")
			tidyCmd := exec.Command("go", "mod", "tidy")
			if _, errStat := os.Stat("go.mod"); errStat != nil {
				tidyCmd.Dir = "go"
			}
			tidyCmd.Run()

			if errRetry := h.CheckBuild(); errRetry != nil {
				return errRetry
			}
			fmt.Println("[+] Self-heal successful.")
			return nil
		}
		return err
	}

	fmt.Println("[+] Build check passed.")
	return nil
}

func main() {
	monitor := HealthMonitor{AutoFix: true}
	if err := monitor.Run(); err != nil {
		fmt.Printf("Health check failed: %v\n", err)
		os.Exit(1)
	} else {
		fmt.Println("System is HEALTHY.")
	}
}
