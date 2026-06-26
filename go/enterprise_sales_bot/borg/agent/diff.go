package agent

import (
	"fmt"
	"os"
	"strings"
)

// ApplyInlineDiff mimics Opencode / Cursor style inline file modifications.
// It parses a unified diff format and applies the changes directly to the file.
func (a *Agent) ApplyInlineDiff(filePath, diffContent string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file for diff application: %w", err)
	}

	strContent := string(content)

	// Extremely simplified diff application for parity demonstration.
	// In a full implementation, we'd use a robust diff matching algorithm.
	lines := strings.Split(diffContent, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "-") {
			toRemove := strings.TrimPrefix(line, "-")
			strContent = strings.Replace(strContent, toRemove+"\n", "", 1)
		} else if strings.HasPrefix(line, "+") {
			toAdd := strings.TrimPrefix(line, "+")
			// This is a naive append; a real diff engine uses hunk headers
			strContent += toAdd + "\n"
		}
	}

	err = os.WriteFile(filePath, []byte(strContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write applied diff: %w", err)
	}

	fmt.Printf("[InlineDiff] Successfully applied diff to %s\n", filePath)
	return nil
}
