package submodules

/**
 * @file service.go
 * @module go/internal/submodules
 *
 * WHAT: Submodule management — tracks, updates, and syncs git submodules.
 *
 * WHY: TormentNexus uses many submodules as reference implementations.
 *      This service manages their lifecycle and detects upstream changes.
 *
 * ADDED: v1.0.0-alpha.32
 */

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// SubmoduleStatus represents the state of a git submodule.
type SubmoduleStatus struct {
	Path         string    `json:"path"`
	URL          string    `json:"url"`
	Branch       string    `json:"branch"`
	CurrentSHA   string    `json:"currentSha"`
	LatestSHA    string    `json:"latestSha"`
	IsDirty      bool      `json:"isDirty"`
	NeedsUpdate  bool      `json:"needsUpdate"`
	LastChecked  time.Time `json:"lastChecked"`
}

// SubmoduleService manages git submodules.
type SubmoduleService struct {
	rootPath string
}

// NewSubmoduleService creates a new service for the given repo root.
func NewSubmoduleService(rootPath string) *SubmoduleService {
	return &SubmoduleService{rootPath: rootPath}
}

// List returns the status of all submodules.
func (ss *SubmoduleService) List() ([]SubmoduleStatus, error) {
	output, err := ss.git("submodule", "status", "--recursive")
	if err != nil {
		return nil, err
	}

	var subs []SubmoduleStatus
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Format: "+<sha> <path> (<describe>)" or "-<sha> <path> (<describe>)"
		prefix := ""
		if len(line) > 0 && (line[0] == '+' || line[0] == '-' || line[0] == ' ') {
			prefix = string(line[0])
			line = line[1:]
		}

		parts := strings.SplitN(line, " ", 3)
		if len(parts) < 2 {
			continue
		}

		sha := parts[0]
		path := parts[1]

		sub := SubmoduleStatus{
			Path:        path,
			CurrentSHA:  sha,
			IsDirty:     prefix == "+",
			LastChecked: time.Now().UTC(),
		}

		// Get URL
		if url, err := ss.git("config", "--file", ".gitmodules", "--get", "submodule."+path+".url"); err == nil {
			sub.URL = strings.TrimSpace(url)
		}

		// Get branch
		if branch, err := ss.git("config", "--file", ".gitmodules", "--get", "submodule."+path+".branch"); err == nil {
			sub.Branch = strings.TrimSpace(branch)
		}

		subs = append(subs, sub)
	}

	return subs, nil
}

// UpdateAll updates all submodules to latest upstream.
func (ss *SubmoduleService) UpdateAll() ([]SubmoduleStatus, error) {
	_, err := ss.git("submodule", "update", "--remote", "--merge")
	if err != nil {
		return nil, err
	}

	return ss.List()
}

// Update updates a single submodule.
func (ss *SubmoduleService) Update(path string) error {
	_, err := ss.git("submodule", "update", "--remote", "--merge", path)
	return err
}

// Sync syncs submodule URLs from .gitmodules to .git/config.
func (ss *SubmoduleService) Sync() error {
	_, err := ss.git("submodule", "sync", "--recursive")
	return err
}

// Fetch fetches upstream changes for all submodules.
func (ss *SubmoduleService) Fetch() error {
	_, err := ss.git("submodule", "foreach", "--recursive", "git fetch")
	return err
}

// CheckUpdates checks which submodules have upstream changes.
func (ss *SubmoduleService) CheckUpdates() ([]SubmoduleStatus, error) {
	subs, err := ss.List()
	if err != nil {
		return nil, err
	}

	for i := range subs {
		// Check for latest SHA on the branch
		branch := subs[i].Branch
		if branch == "" {
			branch = "main"
		}
		subPath := filepath.Join(ss.rootPath, subs[i].Path)

		if latest, err := exec.Command("git", "-C", subPath, "rev-parse", "origin/"+branch).Output(); err == nil {
			subs[i].LatestSHA = strings.TrimSpace(string(latest))
			subs[i].NeedsUpdate = subs[i].LatestSHA != subs[i].CurrentSHA
		}
	}

	return subs, nil
}

// Diff returns the diff between current and upstream for a submodule.
func (ss *SubmoduleService) Diff(path string) (string, error) {
	subPath := filepath.Join(ss.rootPath, path)
	output, err := exec.Command("git", "-C", subPath, "log", "--oneline", "HEAD..origin/main").Output()
	if err != nil {
		// Try origin/master
		output, err = exec.Command("git", "-C", subPath, "log", "--oneline", "HEAD..origin/master").Output()
		if err != nil {
			return "", err
		}
	}
	return string(output), nil
}

// CommitCount returns the number of commits behind upstream for a submodule.
func (ss *SubmoduleService) CommitCount(path string) (int, error) {
	output, err := ss.Diff(path)
	if err != nil {
		return 0, err
	}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if output == "" {
		return 0, nil
	}
	return len(lines), nil
}

func (ss *SubmoduleService) git(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = ss.rootPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %s: %w", strings.Join(args, " "), string(output), err)
	}
	return string(output), nil
}

// Ensure unused imports are used
var _ = strconv.Itoa
