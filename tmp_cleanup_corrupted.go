//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	patterns := []string{
		"*deepseek-reasoner*",
		"*openai/gpt*",
		"*qwen*",
		"*Mistral*",
		"*huggingface*",
		"*nvidia*",
	}
	dir := "go/internal/tools"
	count := 0
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || !strings.HasSuffix(path, ".go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		content := string(data)
		original := content
		// Remove lines that are model reference markers
		lines := strings.Split(content, "\n")
		var cleanLines []string
		skipNext := false
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "---" {
				skipNext = true
				continue
			}
			if skipNext {
				skipNext = false
				// Check if this line matches a model reference pattern
				for _, pat := range patterns {
					if strings.Contains(trimmed, pat) {
						// Skip this line too
						goto skip
					}
				}
				// If we get here, the "---" was not followed by a model reference, keep both
				// Actually, re-add the "---" line and continue
				goto keep
			skip:
			}
			continue
		keep:
			cleanLines = append(cleanLines, line)
			_ = i
		}
		_ = lines
		_ = cleanLines
		// Actually, just remove "---" lines and subsequent model reference lines
		_ = original
		_ = count
		return nil
	})
	fmt.Println("Done - partial implementation, using Python instead")
}
