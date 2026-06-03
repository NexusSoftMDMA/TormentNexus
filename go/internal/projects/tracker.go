package projects

/**
 * @file tracker.go
 * @module go/internal/projects
 *
 * WHAT: Project tracker — manages a registry of known projects/workspaces,
 *       tracks their activity, last session, and status.
 *
 * WHY: TormentNexus needs to know about all projects the user works on,
 *       track their activity, and enable multi-project orchestration.
 *
 * ADDED: v1.0.0-alpha.32
 */

import (
	"fmt"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Project represents a known project/workspace.
type Project struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Path         string            `json:"path"`
	Language     string            `json:"language,omitempty"`
	Framework    string            `json:"framework,omitempty"`
	Description  string            `json:"description,omitempty"`
	Tags         []string          `json:"tags,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	LastActiveAt time.Time         `json:"lastActiveAt"`
	SessionCount int               `json:"sessionCount"`
	FileCount    int               `json:"fileCount"`
	GitBranch    string            `json:"gitBranch,omitempty"`
	GitRemote    string            `json:"gitRemote,omitempty"`
	DiscoveredAt time.Time         `json:"discoveredAt"`
}

// ProjectTracker manages known projects.
type ProjectTracker struct {
	mu       sync.RWMutex
	projects map[string]*Project // keyed by path
	storePath string
}

// NewProjectTracker creates a new tracker.
func NewProjectTracker(storePath string) *ProjectTracker {
	pt := &ProjectTracker{
		projects:  make(map[string]*Project),
		storePath: storePath,
	}
	if storePath != "" {
		_ = pt.loadFromDisk()
	}
	return pt
}

// Register adds or updates a project.
func (pt *ProjectTracker) Register(path string) (*Project, error) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	if existing, ok := pt.projects[absPath]; ok {
		existing.LastActiveAt = time.Now().UTC()
		return existing, nil
	}

	name := filepath.Base(absPath)
	project := &Project{
		ID:           generateProjectID(absPath),
		Name:         name,
		Path:         absPath,
		LastActiveAt: time.Now().UTC(),
		DiscoveredAt: time.Now().UTC(),
		Metadata:     make(map[string]string),
	}

	// Detect language/framework
	pt.detectProjectType(absPath, project)

	pt.projects[absPath] = project
	_ = pt.saveToDisk()

	return project, nil
}

// Get retrieves a project by path.
func (pt *ProjectTracker) Get(path string) (*Project, bool) {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	absPath, _ := filepath.Abs(path)
	p, ok := pt.projects[absPath]
	return p, ok
}

// List returns all projects sorted by last active.
func (pt *ProjectTracker) List() []*Project {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	var projects []*Project
	for _, p := range pt.projects {
		projects = append(projects, p)
	}

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].LastActiveAt.After(projects[j].LastActiveAt)
	})

	return projects
}

// Recent returns the N most recently active projects.
func (pt *ProjectTracker) Recent(n int) []*Project {
	all := pt.List()
	if len(all) > n {
		return all[:n]
	}
	return all
}

// Remove removes a project from tracking.
func (pt *ProjectTracker) Remove(path string) bool {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	absPath, _ := filepath.Abs(path)
	if _, ok := pt.projects[absPath]; ok {
		delete(pt.projects, absPath)
		_ = pt.saveToDisk()
		return true
	}
	return false
}

// Count returns the number of tracked projects.
func (pt *ProjectTracker) Count() int {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	return len(pt.projects)
}

// Stats returns project tracking statistics.
func (pt *ProjectTracker) Stats() map[string]interface{} {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	languages := make(map[string]int)
	for _, p := range pt.projects {
		if p.Language != "" {
			languages[p.Language]++
		}
	}

	return map[string]interface{}{
		"total":      len(pt.projects),
		"languages":  languages,
	}
}

// --- Internal ---

func (pt *ProjectTracker) detectProjectType(path string, project *Project) {
	// Check for common project markers
	checks := []struct {
		file       string
		language   string
		framework  string
	}{
		{"go.mod", "go", ""},
		{"package.json", "javascript", "node"},
		{"tsconfig.json", "typescript", ""},
		{"Cargo.toml", "rust", ""},
		{"pyproject.toml", "python", ""},
		{"requirements.txt", "python", ""},
		{"Gemfile", "ruby", ""},
		{"pom.xml", "java", "maven"},
		{"build.gradle", "java", "gradle"},
		{".csproj", "csharp", "dotnet"},
	}

	for _, check := range checks {
		if _, err := os.Stat(filepath.Join(path, check.file)); err == nil {
			if project.Language == "" {
				project.Language = check.language
			}
			if project.Framework == "" {
				project.Framework = check.framework
			}
		}
	}

	// Refine TypeScript detection
	if project.Language == "javascript" {
		if _, err := os.Stat(filepath.Join(path, "tsconfig.json")); err == nil {
			project.Language = "typescript"
		}
	}

	// Detect framework from package.json
	if project.Language == "javascript" || project.Language == "typescript" {
		pkgData, err := os.ReadFile(filepath.Join(path, "package.json"))
		if err == nil {
			var pkg map[string]interface{}
			if json.Unmarshal(pkgData, &pkg) == nil {
				deps, _ := pkg["dependencies"].(map[string]interface{})
				devDeps, _ := pkg["devDependencies"].(map[string]interface{})
				allDeps := make(map[string]bool)
				for k := range deps {
					allDeps[k] = true
				}
				for k := range devDeps {
					allDeps[k] = true
				}

				switch {
				case allDeps["next"]:
					project.Framework = "nextjs"
				case allDeps["react"]:
					project.Framework = "react"
				case allDeps["vue"]:
					project.Framework = "vue"
				case allDeps["express"]:
					project.Framework = "express"
				case allDeps["fastify"]:
					project.Framework = "fastify"
				case allDeps["@angular/core"]:
					project.Framework = "angular"
				case allDeps["svelte"]:
					project.Framework = "svelte"
				}
			}
		}
	}
}

func (pt *ProjectTracker) saveToDisk() error {
	if pt.storePath == "" {
		return nil
	}

	var projects []*Project
	for _, p := range pt.projects {
		projects = append(projects, p)
	}

	data, err := json.MarshalIndent(projects, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(pt.storePath), 0o755); err != nil {
		return err
	}

	return os.WriteFile(pt.storePath, data, 0o644)
}

func (pt *ProjectTracker) loadFromDisk() error {
	data, err := os.ReadFile(pt.storePath)
	if err != nil {
		return err
	}

	var projects []*Project
	if err := json.Unmarshal(data, &projects); err != nil {
		return err
	}

	for _, p := range projects {
		pt.projects[p.Path] = p
	}

	return nil
}

func generateProjectID(path string) string {
	h := 0
	for _, c := range path {
		h = h*31 + int(c)
	}
	return fmt.Sprintf("proj_%08x", uint32(h))
}

// Ensure fmt and strings are used
var _ = strings.TrimSpace
var _ = fmt.Sprintf
