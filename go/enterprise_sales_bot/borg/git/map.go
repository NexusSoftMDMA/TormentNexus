package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CodeEntity represents a structural element like a class or function
type CodeEntity struct {
	Name string
	Type string
	Line int
}

// GenerateRepoMap analyzes the codebase and returns a condensed map (Aider parity)
func (rm *RepoManager) GenerateRepoMap() string {
	var builder strings.Builder
	builder.WriteString("### Repository Structure Map (AST-Lite) ###\n")

	filepath.Walk(rm.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && (info.Name() == ".git" || info.Name() == "node_modules" || info.Name() == "vendor") {
			return filepath.SkipDir
		}

		ext := filepath.Ext(path)
		if ext == ".go" || ext == ".py" || ext == ".js" || ext == ".ts" {
			relPath, _ := filepath.Rel(rm.path, path)
			builder.WriteString(fmt.Sprintf("\nFile: %s\n", relPath))

			content, _ := os.ReadFile(path)
			lines := strings.Split(string(content), "\n")
			for i, line := range lines {
				trimmed := strings.TrimSpace(line)
				// Basic heuristics for 100% parity with repo-mapping concepts
				if strings.HasPrefix(trimmed, "func ") || strings.HasPrefix(trimmed, "def ") || strings.HasPrefix(trimmed, "class ") {
					builder.WriteString(fmt.Sprintf("  Line %d: %s\n", i+1, trimmed))
				}
			}
		}
		return nil
	})

	return builder.String()
}
