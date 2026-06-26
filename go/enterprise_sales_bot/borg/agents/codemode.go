package agents

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// CodeMode executes an LLM-generated string of code directly using an external compiler/runtime natively
// This enforces the "Escape Hatch - Code Mode - 94% context reduction" specification from AGENTS.md
func ExecuteCodeMode(scriptBundle string, language string) (string, error) {
	tmpDir := os.TempDir()

	switch language {
	case "typescript", "ts":
		tmpFile := filepath.Join(tmpDir, "tormentnexus_codemode.ts")
		if err := os.WriteFile(tmpFile, []byte(scriptBundle), 0644); err != nil {
			return "", err
		}

		// Map execution directly using local Bun runtime
		cmd := exec.Command("bun", "run", tmpFile)
		output, err := cmd.CombinedOutput()
		os.Remove(tmpFile)

		if err != nil {
			return string(output), fmt.Errorf("CodeMode exception: %w\n%s", err, output)
		}
		return string(output), nil

	case "bash", "sh", "pwsh":
		tmpFile := filepath.Join(tmpDir, "tormentnexus_codemode.sh")
		if err := os.WriteFile(tmpFile, []byte(scriptBundle), 0644); err != nil {
			return "", err
		}

		cmd := exec.Command("pwsh", "-File", tmpFile)
		output, err := cmd.CombinedOutput()
		os.Remove(tmpFile)

		if err != nil {
			return string(output), fmt.Errorf("CodeMode native exception: %w\n%s", err, output)
		}
		return string(output), nil
	default:
		return "", fmt.Errorf("missing language execution binding for %s", language)
	}
}
