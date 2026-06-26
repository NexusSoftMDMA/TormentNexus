package memorystore

/**
 * @file hydration.go
 * @module go/internal/memorystore
 *
 * WHAT: Memory hydration engine that bootstraps the Go sidecar's context
 * store with essential project knowledge for autonomous operation.
 *
 * WHY: Total Autonomy — The Go sidecar needs a populated memory store to
 * operate independently. Without hydrated context, the sidecar cannot make
 * informed decisions about tool selection, code architecture, or project state.
 */

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// HydrationEntry is a single context record in the hydration store.
type HydrationEntry struct {
	ID        string            `json:"id"`
	Section   string            `json:"section"`
	Key       string            `json:"key"`
	Content   string            `json:"content"`
	Source    string            `json:"source"`
	Tags      []string          `json:"tags,omitempty"`
	CreatedAt string            `json:"createdAt"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// HydrationStore is the Go-native memory hydration store.
type HydrationStore struct {
	mu       sync.RWMutex
	path     string
	entries  []HydrationEntry
	sections map[string][]*HydrationEntry
	dirty    bool
}

// NewHydrationStore creates or loads a hydration store at the given path.
func NewHydrationStore(workspaceRoot string) *HydrationStore {
	storeDir := filepath.Join(workspaceRoot, ".tormentnexus", "hydration")
	storePath := filepath.Join(storeDir, "context.json")

	hs := &HydrationStore{
		path:     storePath,
		entries:  make([]HydrationEntry, 0),
		sections: make(map[string][]*HydrationEntry),
	}

	// Attempt to load existing store
	if data, err := os.ReadFile(storePath); err == nil {
		var entries []HydrationEntry
		if err := json.Unmarshal(data, &entries); err == nil {
			hs.entries = entries
			for i := range hs.entries {
				entry := &hs.entries[i]
				hs.sections[entry.Section] = append(hs.sections[entry.Section], entry)
			}
		}
	}

	return hs
}

// HydrateFromWorkspace scans the workspace for essential context and populates
// the hydration store. This is the primary bootstrap operation.
func (hs *HydrationStore) HydrateFromWorkspace(ctx context.Context, workspaceRoot string) (*HydrationReport, error) {
	report := &HydrationReport{
		StartedAt: time.Now().UTC().Format(time.RFC3339),
	}

	// 1. Project context from package.json / go.mod / Cargo.toml
	report.ProjectContext = hs.ingestProjectContext(workspaceRoot)

	// 2. Architecture from directory structure
	report.ArchitectureEntries = hs.ingestArchitecture(workspaceRoot)

	// 3. Agent instructions from AGENTS.md
	report.AgentInstructions = hs.ingestAgentInstructions(workspaceRoot)

	// 4. Key configuration files
	report.ConfigEntries = hs.ingestConfigFiles(workspaceRoot)

	// 5. Repository graph summary (if repograph data exists)
	report.RepoGraphEntries = hs.ingestRepoGraphSummary(workspaceRoot)

	// 6. Environment context
	report.EnvironmentEntries = hs.ingestEnvironment()

	// Save the hydrated store
	if err := hs.Save(); err != nil {
		return nil, fmt.Errorf("failed to save hydration store: %w", err)
	}

	report.CompletedAt = time.Now().UTC().Format(time.RFC3339)
	report.TotalEntries = len(hs.entries)
	report.Sections = hs.SectionNames()

	return report, nil
}

// Get retrieves entries by section and optional key.
func (hs *HydrationStore) Get(section string, key string) []*HydrationEntry {
	hs.mu.RLock()
	defer hs.mu.RUnlock()

	entries := hs.sections[section]
	if key == "" {
		return entries
	}

	var result []*HydrationEntry
	for _, e := range entries {
		if e.Key == key {
			result = append(result, e)
		}
	}
	return result
}

// Query searches the hydration store by content substring.
func (hs *HydrationStore) Query(query string) []*HydrationEntry {
	hs.mu.RLock()
	defer hs.mu.RUnlock()

	lowerQuery := strings.ToLower(query)
	var results []*HydrationEntry
	for i := range hs.entries {
		if strings.Contains(strings.ToLower(hs.entries[i].Content), lowerQuery) ||
			strings.Contains(strings.ToLower(hs.entries[i].Key), lowerQuery) ||
			strings.Contains(strings.ToLower(hs.entries[i].Section), lowerQuery) {
			results = append(results, &hs.entries[i])
		}
	}
	return results
}

// All returns all entries.
func (hs *HydrationStore) All() []HydrationEntry {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	return append([]HydrationEntry(nil), hs.entries...)
}

// SectionNames returns the names of all populated sections.
func (hs *HydrationStore) SectionNames() []string {
	hs.mu.RLock()
	defer hs.mu.RUnlock()

	names := make([]string, 0, len(hs.sections))
	for name, entries := range hs.sections {
		if len(entries) > 0 {
			names = append(names, name)
		}
	}
	return names
}

// SectionCounts returns entry counts per section.
func (hs *HydrationStore) SectionCounts() map[string]int {
	hs.mu.RLock()
	defer hs.mu.RUnlock()

	counts := make(map[string]int, len(hs.sections))
	for name, entries := range hs.sections {
		counts[name] = len(entries)
	}
	return counts
}

// Add inserts a new entry into the store.
func (hs *HydrationStore) Add(entry HydrationEntry) {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	if entry.ID == "" {
		entry.ID = fmt.Sprintf("hyd-%d", time.Now().UnixMilli())
	}
	if entry.CreatedAt == "" {
		entry.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	hs.entries = append(hs.entries, entry)
	hs.sections[entry.Section] = append(hs.sections[entry.Section], &hs.entries[len(hs.entries)-1])
	hs.dirty = true
}

// Save persists the store to disk.
func (hs *HydrationStore) Save() error {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	if !hs.dirty {
		return nil
	}

	data, err := json.MarshalIndent(hs.entries, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(hs.path), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(hs.path, data, 0644); err != nil {
		return err
	}

	hs.dirty = false
	return nil
}

// --- Ingest Methods ---

func (hs *HydrationStore) ingestProjectContext(workspaceRoot string) int {
	count := 0

	// Go module (workspace root or nested go/ dir)
	goModPaths := []string{
		filepath.Join(workspaceRoot, "go.mod"),
		filepath.Join(workspaceRoot, "go", "go.mod"),
	}
	for _, goModPath := range goModPaths {
		if data, err := os.ReadFile(goModPath); err == nil {
			hs.Add(HydrationEntry{
				Section: "project_context",
				Key:     "go.mod",
				Content: strings.TrimSpace(string(data)),
				Source:  "file-scan",
				Tags:    []string{"go", "module", "dependencies"},
			})
			count++
			break
		}
	}

	// Node.js package.json
	if data, err := os.ReadFile(filepath.Join(workspaceRoot, "package.json")); err == nil {
		var pkg struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		}
		if json.Unmarshal(data, &pkg) == nil {
			hs.Add(HydrationEntry{
				Section: "project_context",
				Key:     "package.json",
				Content: fmt.Sprintf("Project: %s, Version: %s", pkg.Name, pkg.Version),
				Source:  "file-scan",
				Tags:    []string{"node", "package", "version"},
			})
			count++
		}
	}

	// Cargo.toml
	if data, err := os.ReadFile(filepath.Join(workspaceRoot, "Cargo.toml")); err == nil {
		hs.Add(HydrationEntry{
			Section: "project_context",
			Key:     "Cargo.toml",
			Content: string(data),
			Source:  "file-scan",
			Tags:    []string{"rust", "cargo", "dependencies"},
		})
		count++
	}

	// pyproject.toml
	if data, err := os.ReadFile(filepath.Join(workspaceRoot, "pyproject.toml")); err == nil {
		hs.Add(HydrationEntry{
			Section: "project_context",
			Key:     "pyproject.toml",
			Content: string(data),
			Source:  "file-scan",
			Tags:    []string{"python", "project", "dependencies"},
		})
		count++
	}

	return count
}

func (hs *HydrationStore) ingestArchitecture(workspaceRoot string) int {
	count := 0

	// Scan top-level directories
	entries, err := os.ReadDir(workspaceRoot)
	if err != nil {
		return 0
	}

	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			dirs = append(dirs, entry.Name())
		}
	}

	if len(dirs) > 0 {
		hs.Add(HydrationEntry{
			Section: "architecture",
			Key:     "top-level-structure",
			Content: fmt.Sprintf("Workspace top-level directories: %s", strings.Join(dirs, ", ")),
			Source:  "directory-scan",
			Tags:    []string{"architecture", "structure"},
		})
		count++
	}

	// Scan Go internal packages
	goInternalPath := filepath.Join(workspaceRoot, "go", "internal")
	if entries, err := os.ReadDir(goInternalPath); err == nil {
		var packages []string
		for _, entry := range entries {
			if entry.IsDir() {
				packages = append(packages, entry.Name())
			}
		}
		if len(packages) > 0 {
			hs.Add(HydrationEntry{
				Section: "architecture",
				Key:     "go-internal-packages",
				Content: fmt.Sprintf("Go internal packages: %s", strings.Join(packages, ", ")),
				Source:  "directory-scan",
				Tags:    []string{"go", "architecture", "packages"},
			})
			count++
		}
	}

	// Scan Next.js app routes
	appPath := filepath.Join(workspaceRoot, "apps", "web", "src", "app")
	if entries, err := os.ReadDir(appPath); err == nil {
		var routes []string
		for _, entry := range entries {
			if entry.IsDir() {
				routes = append(routes, entry.Name())
			}
		}
		if len(routes) > 0 {
			hs.Add(HydrationEntry{
				Section: "architecture",
				Key:     "nextjs-routes",
				Content: fmt.Sprintf("Next.js dashboard routes: %s", strings.Join(routes, ", ")),
				Source:  "directory-scan",
				Tags:    []string{"nextjs", "routes", "dashboard"},
			})
			count++
		}
	}

	return count
}

func (hs *HydrationStore) ingestAgentInstructions(workspaceRoot string) int {
	count := 0

	// AGENTS.md
	if data, err := os.ReadFile(filepath.Join(workspaceRoot, "AGENTS.md")); err == nil {
		hs.Add(HydrationEntry{
			Section: "agent_instructions",
			Key:     "AGENTS.md",
			Content: string(data),
			Source:  "file-scan",
			Tags:    []string{"agents", "instructions", "guidelines"},
		})
		count++
	}

	// .tormentnexus/instructions.md
	if data, err := os.ReadFile(filepath.Join(workspaceRoot, ".tormentnexus", "instructions.md")); err == nil {
		hs.Add(HydrationEntry{
			Section: "agent_instructions",
			Key:     "tormentnexus-instructions",
			Content: string(data),
			Source:  "file-scan",
			Tags:    []string{"tormentnexus", "instructions"},
		})
		count++
	}

	return count
}

func (hs *HydrationStore) ingestConfigFiles(workspaceRoot string) int {
	count := 0

	configFiles := []struct {
		path string
		key  string
		tags []string
	}{
		{filepath.Join(workspaceRoot, "go", "cmd", "tormentnexus", "main.go"), "go-main-entrypoint", []string{"go", "entrypoint"}},
		{filepath.Join(workspaceRoot, "apps", "web", "next.config.js"), "nextjs-config", []string{"nextjs", "config"}},
		{filepath.Join(workspaceRoot, "go", "internal", "config", "config.go"), "go-config", []string{"go", "config"}},
		{filepath.Join(workspaceRoot, "tsconfig.json"), "tsconfig", []string{"typescript", "config"}},
	}

	for _, cf := range configFiles {
		if data, err := os.ReadFile(cf.path); err == nil {
			// Truncate large files
			content := string(data)
			if len(content) > 2000 {
				content = content[:2000] + "\n... (truncated)"
			}
			hs.Add(HydrationEntry{
				Section: "configuration",
				Key:     cf.key,
				Content: content,
				Source:  "file-scan",
				Tags:    cf.tags,
			})
			count++
		}
	}

	return count
}

func (hs *HydrationStore) ingestRepoGraphSummary(workspaceRoot string) int {
	count := 0

	// Check for repograph cache
	graphPath := filepath.Join(workspaceRoot, ".tormentnexus", "repograph.json")
	if data, err := os.ReadFile(graphPath); err == nil {
		var graph struct {
			Modules []struct {
				Name     string `json:"name"`
				Language string `json:"language"`
			} `json:"modules"`
		}
		if json.Unmarshal(data, &graph) == nil && len(graph.Modules) > 0 {
			var summary []string
			for _, m := range graph.Modules {
				summary = append(summary, fmt.Sprintf("%s (%s)", m.Name, m.Language))
			}
			hs.Add(HydrationEntry{
				Section: "repo_graph",
				Key:     "module-summary",
				Content: fmt.Sprintf("Indexed modules: %s", strings.Join(summary, ", ")),
				Source:  "repograph-cache",
				Tags:    []string{"repograph", "modules", "index"},
			})
			count++
		}
	}

	return count
}

func (hs *HydrationStore) ingestEnvironment() int {
	count := 0

	envInfo := map[string]string{
		"os":      runtime.GOOS,
		"arch":    runtime.GOARCH,
		"cpuCount": fmt.Sprintf("%d", runtime.NumCPU()),
	}

	// Check for key env vars
	keyEnvs := []string{
		"TORMENTNEXUS_TRPC_UPSTREAM",
		"TORMENTNEXUS_GO_PORT",
		"TORMENTNEXUS_WORKSPACE_ROOT",
		"TORMENTNEXUS_MAIN_CONFIG_DIR",
	}
	for _, key := range keyEnvs {
		if val := os.Getenv(key); val != "" {
			envInfo[key] = val
		}
	}

	data, _ := json.Marshal(envInfo)
	hs.Add(HydrationEntry{
		Section: "environment",
		Key:     "runtime-context",
		Content: string(data),
		Source:  "runtime-probe",
		Tags:    []string{"environment", "runtime", "system"},
	})
	count++

	return count
}

// HydrationReport summarizes the results of a hydration operation.
type HydrationReport struct {
	StartedAt            string   `json:"startedAt"`
	CompletedAt          string   `json:"completedAt"`
	TotalEntries         int      `json:"totalEntries"`
	Sections             []string `json:"sections"`
	ProjectContext       int      `json:"projectContextEntries"`
	ArchitectureEntries  int      `json:"architectureEntries"`
	AgentInstructions    int      `json:"agentInstructionsEntries"`
	ConfigEntries        int      `json:"configEntries"`
	RepoGraphEntries     int      `json:"repoGraphEntries"`
	EnvironmentEntries   int      `json:"environmentEntries"`
}
