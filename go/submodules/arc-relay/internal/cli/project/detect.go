package project

import (
	"fmt"
	"os"
	"path/filepath"
)

// projectMarkers are files/dirs that indicate a project root.
// We check these in order of specificity. Note: bare ".claude/" is NOT
// a project marker because it exists in every user's home directory as
// Claude's global config. We look for ".claude/CLAUDE.md" instead,
// which indicates a project-level Claude config. Codex projects use a
// project-local ".codex/" directory, so it is treated as a marker.
var projectMarkers = []string{
	".mcp.json",
	".codex",
	".git",
}

// secondaryMarkers are checked only if they're NOT in a user home directory.
var secondaryMarkers = []string{
	".claude",
}

// DetectProjectDir walks up from startDir looking for project markers.
// Returns the first directory that contains any marker, or startDir itself
// if no marker is found (the user likely wants to create a new config here).
func DetectProjectDir(startDir string) (string, error) {
	absDir, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("resolving path %s: %w", startDir, err)
	}

	dir := absDir
	for {
		// Check primary markers (always valid)
		for _, marker := range projectMarkers {
			path := filepath.Join(dir, marker)
			if _, err := os.Stat(path); err == nil {
				return dir, nil
			}
		}

		// Check secondary markers only if this isn't a home directory.
		// ~/.claude/ is Claude's global config, not a project marker.
		if !isHomeDir(dir) {
			for _, marker := range secondaryMarkers {
				path := filepath.Join(dir, marker)
				if _, err := os.Stat(path); err == nil {
					return dir, nil
				}
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root without finding a marker.
			// Use the original directory as the project dir.
			return absDir, nil
		}
		dir = parent
	}
}

// DetectTargets returns which targets have existing config files or are
// applicable in the given project directory.
func DetectTargets(projectDir string, targets []Target) []Target {
	var detected []Target
	for _, t := range targets {
		if t.Detect(projectDir) {
			detected = append(detected, t)
		}
	}
	return detected
}

// isHomeDir returns true if dir appears to be a user home directory.
// Checks both the OS-reported home and common patterns (e.g., /home/*, /Users/*,
// /mnt/c/Users/* on WSL, C:\Users\* on Windows).
func isHomeDir(dir string) bool {
	if homeDir, err := os.UserHomeDir(); err == nil && dir == homeDir {
		return true
	}
	// Check if the directory is directly under a known "users" parent.
	// Works cross-platform by using filepath operations.
	parent := filepath.Dir(dir)
	parentBase := filepath.Base(parent)

	// /home/<user>, /Users/<user>, C:\Users\<user>
	if parentBase == "home" || parentBase == "Users" {
		return true
	}

	// WSL: /mnt/c/Users/<user> — parent is /mnt/c/Users, grandparent is /mnt/c
	grandparent := filepath.Dir(parent)
	if parentBase == "Users" && filepath.Base(filepath.Dir(grandparent)) == "mnt" {
		return true
	}

	return false
}

// AllTargets returns the list of all known targets.
func AllTargets() []Target {
	return []Target{
		&ClaudeCodeTarget{},
		&CodexTarget{},
	}
}
