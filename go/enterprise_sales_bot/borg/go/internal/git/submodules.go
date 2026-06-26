package git

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type SubmoduleResult struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Success bool   `json:"success"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

type UpdateReport struct {
	Total      int               `json:"total"`
	Successful int               `json:"successful"`
	Failed     int               `json:"failed"`
	Details    []SubmoduleResult `json:"details"`
}

// ListSubmodules parses the output of `git submodule status` to retrieve paths.
func ListSubmodules(ctx context.Context, repoRoot string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "submodule", "status")
	cmd.Dir = repoRoot
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list submodules: %w", err)
	}

	return parseSubmoduleStatusOutput(string(output)), nil
}

func parseSubmoduleStatusOutput(output string) []string {
	var paths []string
	for _, line := range strings.Split(output, "\n") {
		if path, ok := parseSubmoduleStatusLine(line); ok {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	return paths
}

func parseSubmoduleStatusLine(line string) (string, bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return "", false
	}

	// git submodule status format is effectively:
	// [status+hash] [path] ([ref])
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return "", false
	}

	path := strings.TrimSpace(parts[1])
	if path == "" {
		return "", false
	}

	return path, true
}

// UpdateAll concurrently updates all submodules to their remote tracked branches.
func UpdateAll(ctx context.Context, repoRoot string) (*UpdateReport, error) {
	paths, err := ListSubmodules(ctx, repoRoot)
	if err != nil {
		return nil, err
	}

	report := &UpdateReport{
		Total:   len(paths),
		Details: make([]SubmoduleResult, 0, len(paths)),
	}

	if len(paths) == 0 {
		return report, nil
	}

	// First run init recursively
	initCmd := exec.CommandContext(ctx, "git", "submodule", "update", "--init", "--recursive")
	initCmd.Dir = repoRoot
	if err := initCmd.Run(); err != nil {
		// Non-fatal, some submodules might be broken, continue to individual updates
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	// Update each submodule concurrently
	for _, path := range paths {
		wg.Add(1)
		go func(subPath string) {
			defer wg.Done()
			
			// We run fetch + checkout + pull inside the submodule directory
			// To mimic `git submodule update --remote` but with precise error tracking
			name := filepathBase(subPath) // Simple extraction

			cmd := exec.CommandContext(ctx, "git", "submodule", "update", "--remote", subPath)
			cmd.Dir = repoRoot
			
			output, err := cmd.CombinedOutput()
			
			mu.Lock()
			defer mu.Unlock()
			
			res := SubmoduleResult{
				Name:    name,
				Path:    subPath,
				Success: err == nil,
				Output:  string(output),
			}
			
			if err != nil {
				res.Error = err.Error()
				report.Failed++
			} else {
				report.Successful++
			}
			report.Details = append(report.Details, res)
		}(path)
	}

	wg.Wait()
	sort.Slice(report.Details, func(i, j int) bool {
		if report.Details[i].Path == report.Details[j].Path {
			return report.Details[i].Name < report.Details[j].Name
		}
		return report.Details[i].Path < report.Details[j].Path
	})
	return report, nil
}

func filepathBase(path string) string {
	if path == "" {
		return path
	}
	return filepath.Base(strings.ReplaceAll(path, "\\", "/"))
}
