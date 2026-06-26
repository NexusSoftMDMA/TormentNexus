package knowledge

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Resource represents a knowledge resource.
type Resource struct {
	Path        string `json:"path"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Size        int64  `json:"size"`
	Type        string `json:"type"`
}

// KnowledgeService manages knowledge resources (READMEs, docs, specs).
type KnowledgeService struct {
	workspaceRoot string
	mu            sync.RWMutex
	cached        []Resource
}

// NewKnowledgeService creates a new knowledge service.
func NewKnowledgeService(workspaceRoot string) *KnowledgeService {
	return &KnowledgeService{
		workspaceRoot: workspaceRoot,
	}
}

// GetResources returns all knowledge resources found in the workspace.
func (ks *KnowledgeService) GetResources() []Resource {
	ks.mu.RLock()
	if ks.cached != nil {
		defer ks.mu.RUnlock()
		return ks.cached
	}
	ks.mu.RUnlock()

	ks.mu.Lock()
	defer ks.mu.Unlock()

	ks.cached = ks.scanResources()
	return ks.cached
}

// Ingest adds a URL or file path as a knowledge resource.
func (ks *KnowledgeService) Ingest(path string) (string, error) {
	// For now, just validate the path exists
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return fmt.Sprintf("URL queued for ingestion: %s", path), nil
	}

	absPath := path
	if !filepath.IsAbs(path) {
		absPath = filepath.Join(ks.workspaceRoot, path)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return "", fmt.Errorf("path does not exist: %s", path)
	}

	// Invalidate cache so next scan picks it up
	ks.mu.Lock()
	ks.cached = nil
	ks.mu.Unlock()

	return fmt.Sprintf("Ingested: %s", path), nil
}

// Graph returns a dependency/knowledge graph view.
func (ks *KnowledgeService) Graph() map[string]any {
	resources := ks.GetResources()
	graph := map[string]any{
		"nodes": resources,
		"edges": []map[string]string{},
	}
	return graph
}

// Stats returns knowledge statistics.
func (ks *KnowledgeService) Stats() map[string]any {
	resources := ks.GetResources()
	var totalSize int64
	typeCount := make(map[string]int)
	for _, r := range resources {
		totalSize += r.Size
		typeCount[r.Type]++
	}

	return map[string]any{
		"totalResources": len(resources),
		"totalSize":      totalSize,
		"byType":         typeCount,
	}
}

func (ks *KnowledgeService) scanResources() []Resource {
	var resources []Resource

	// Scan common documentation files
	patterns := []string{
		"README.md", "README.txt", "README",
		"CONTRIBUTING.md", "CHANGELOG.md", "LICENSE",
		"docs/**/*.md", "docs/**/*.txt",
		"*.md", "*.txt",
	}

	visited := make(map[string]bool)

	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(ks.workspaceRoot, pattern))
		if err != nil {
			continue
		}
		for _, match := range matches {
			if visited[match] {
				continue
			}
			visited[match] = true

			info, err := os.Stat(match)
			if err != nil {
				continue
			}

			relPath, _ := filepath.Rel(ks.workspaceRoot, match)
			title := extractTitle(match)
			resType := detectType(match)

			resources = append(resources, Resource{
				Path:        relPath,
				Title:       title,
				Description: fmt.Sprintf("%s (%s, %d bytes)", relPath, resType, info.Size()),
				Size:        info.Size(),
				Type:        resType,
			})
		}
	}

	return resources
}

func extractTitle(filePath string) string {
	f, err := os.Open(filePath)
	if err != nil {
		return filepath.Base(filePath)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
		if line != "" {
			return line
		}
	}
	return filepath.Base(filePath)
}

func detectType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".md":
		return "markdown"
	case ".txt":
		return "text"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".go":
		return "go-source"
	case ".ts", ".tsx":
		return "typescript-source"
	case ".py":
		return "python-source"
	default:
		return "unknown"
	}
}

// --- Handler helpers ---

func (ks *KnowledgeService) HandleList(w http.ResponseWriter, r *http.Request) {
	resources := ks.GetResources()
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": resources})
}

func (ks *KnowledgeService) HandleGraph(w http.ResponseWriter, r *http.Request) {
	graph := ks.Graph()
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": graph})
}

func (ks *KnowledgeService) HandleStats(w http.ResponseWriter, r *http.Request) {
	stats := ks.Stats()
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": stats})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
