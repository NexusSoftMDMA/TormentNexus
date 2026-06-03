// Package marketplace implements the Skill Marketplace REST API for
// downloading, publishing, and managing community-contributed skills.
package marketplace

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// SkillEntry represents a skill in the marketplace.
type SkillEntry struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Author       string            `json:"author"`
	Version      string            `json:"version"`
	Tags         []string          `json:"tags,omitempty"`
	DownloadCount int              `json:"download_count"`
	Rating       float64           `json:"rating"`
	CreatedAt    int64             `json:"created_at"`
	UpdatedAt    int64             `json:"updated_at"`
	Content      string            `json:"content,omitempty"` // Only in full get
	Manifest     map[string]any    `json:"manifest,omitempty"`
}

// InstallRequest is the payload for installing a skill.
type InstallRequest struct {
	SkillID  string `json:"skill_id"`
	Version  string `json:"version,omitempty"`
	Source   string `json:"source,omitempty"` // "registry", "url", "file"
	URL      string `json:"url,omitempty"`
	Force    bool   `json:"force,omitempty"`
}

// PublishRequest is the payload for publishing a skill to the marketplace.
type PublishRequest struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Version     string   `json:"version"`
	Tags        []string `json:"tags,omitempty"`
	Content     string   `json:"content"`
}

// RegistryConfig controls the marketplace registry behavior.
type RegistryConfig struct {
	// RegistryURL is the base URL for the remote skill registry.
	RegistryURL string
	// LocalDir is the directory for locally installed marketplace skills.
	LocalDir string
	// CacheTTL is how long the local skill cache is valid.
	CacheTTL time.Duration
}

// DefaultRegistryConfig returns sensible defaults.
func DefaultRegistryConfig() RegistryConfig {
	homeDir, _ := os.UserHomeDir()
	return RegistryConfig{
		RegistryURL: "https://registry.tormentnexus.dev/api/v1/skills",
		LocalDir:    filepath.Join(homeDir, ".tormentnexus", "marketplace"),
		CacheTTL:    5 * time.Minute,
	}
}

// Service provides marketplace operations.
type Service struct {
	cfg     RegistryConfig
	mu      sync.RWMutex
	cache   []SkillEntry
	cacheAt time.Time
	client  *http.Client
}

// NewService creates a new marketplace service.
func NewService(cfg RegistryConfig) *Service {
	if cfg.LocalDir == "" {
		cfg = DefaultRegistryConfig()
	}
	return &Service{
		cfg: cfg,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// List returns all available skills from the marketplace registry.
func (s *Service) List(query string, tags []string, limit int) ([]SkillEntry, error) {
	if limit <= 0 {
		limit = 50
	}

	// Try the remote registry first
	entries, err := s.fetchRemoteList(query, tags, limit)
	if err == nil && len(entries) > 0 {
		return entries, nil
	}

	// Fallback to local cache
	entries = s.getLocalSkills(query, tags, limit)
	return entries, nil
}

// Get returns a single skill's full details.
func (s *Service) Get(skillID string) (*SkillEntry, error) {
	// Try remote
	entry, err := s.fetchRemoteGet(skillID)
	if err == nil && entry != nil {
		return entry, nil
	}

	// Fallback to local
	entries := s.getLocalSkills(skillID, nil, 1)
	if len(entries) > 0 {
		return &entries[0], nil
	}

	return nil, fmt.Errorf("skill %q not found", skillID)
}

// Install downloads and installs a skill to the local marketplace directory.
func (s *Service) Install(req InstallRequest) (*SkillEntry, error) {
	skillID := strings.TrimSpace(req.SkillID)
	if skillID == "" {
		return nil, fmt.Errorf("missing skill_id")
	}

	// Get the skill content
	var content string
	var entry *SkillEntry

	switch req.Source {
	case "url":
		if req.URL == "" {
			return nil, fmt.Errorf("missing url for url source")
		}
		data, err := s.downloadContent(req.URL)
		if err != nil {
			return nil, fmt.Errorf("download from url: %w", err)
		}
		content = string(data)
		entry = &SkillEntry{ID: skillID, Name: skillID, Content: content}

	case "file":
		if req.URL == "" {
			return nil, fmt.Errorf("missing file path")
		}
		data, err := os.ReadFile(req.URL)
		if err != nil {
			return nil, fmt.Errorf("read file: %w", err)
		}
		content = string(data)
		entry = &SkillEntry{ID: skillID, Name: skillID, Content: content}

	default: // "registry"
		remote, err := s.fetchRemoteGet(skillID)
		if err != nil {
			return nil, fmt.Errorf("fetch from registry: %w", err)
		}
		entry = remote
		content = remote.Content
	}

	if content == "" {
		return nil, fmt.Errorf("no content available for skill %q", skillID)
	}

	// Install to local directory
	skillDir := filepath.Join(s.cfg.LocalDir, skillID)
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		return nil, fmt.Errorf("create skill dir: %w", err)
	}

	// Write the SKILL.md file
	skillFile := filepath.Join(skillDir, "SKILL.md")
	if _, err := os.Stat(skillFile); err == nil && !req.Force {
		return nil, fmt.Errorf("skill %q already installed (use force to overwrite)", skillID)
	}

	// Parse frontmatter from content, inject marketplace metadata
	enrichedContent := s.enrichContent(entry, content)
	if err := os.WriteFile(skillFile, []byte(enrichedContent), 0o644); err != nil {
		return nil, fmt.Errorf("write skill file: %w", err)
	}

	// Write install metadata
	meta := map[string]any{
		"installed_at": time.Now().UnixMilli(),
		"source":       req.Source,
		"version":      entry.Version,
		"author":       entry.Author,
	}
	metaJSON, _ := json.MarshalIndent(meta, "", "  ")
	metaFile := filepath.Join(skillDir, "marketplace.json")
	_ = os.WriteFile(metaFile, metaJSON, 0o644)

	return entry, nil
}

// Publish publishes a skill to the marketplace registry.
func (s *Service) Publish(req PublishRequest) (*SkillEntry, error) {
	if strings.TrimSpace(req.ID) == "" {
		return nil, fmt.Errorf("missing skill id")
	}
	if strings.TrimSpace(req.Content) == "" {
		return nil, fmt.Errorf("missing skill content")
	}

	entry := &SkillEntry{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
		Author:      req.Author,
		Version:     req.Version,
		Tags:        req.Tags,
		Content:     req.Content,
		CreatedAt:   time.Now().UnixMilli(),
		UpdatedAt:   time.Now().UnixMilli(),
	}

	// Try to publish to remote registry
	err := s.publishRemote(entry)
	if err != nil {
		// Store locally for later sync
		localDir := filepath.Join(s.cfg.LocalDir, "pending", req.ID)
		_ = os.MkdirAll(localDir, 0o755)

		pendingJSON, _ := json.MarshalIndent(entry, "", "  ")
		pendingFile := filepath.Join(localDir, "publish.json")
		_ = os.WriteFile(pendingFile, pendingJSON, 0o644)

		return entry, fmt.Errorf("remote publish failed (saved locally for later sync): %w", err)
	}

	return entry, nil
}

// Uninstall removes a locally installed marketplace skill.
func (s *Service) Uninstall(skillID string) error {
	skillDir := filepath.Join(s.cfg.LocalDir, skillID)
	if _, err := os.Stat(skillDir); os.IsNotExist(err) {
		return fmt.Errorf("skill %q is not installed", skillID)
	}
	return os.RemoveAll(skillDir)
}

// Installed returns all locally installed marketplace skills.
func (s *Service) Installed() ([]SkillEntry, error) {
	entries, err := os.ReadDir(s.cfg.LocalDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []SkillEntry{}, nil
		}
		return nil, err
	}

	var results []SkillEntry
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillFile := filepath.Join(s.cfg.LocalDir, entry.Name(), "SKILL.md")
		metaFile := filepath.Join(s.cfg.LocalDir, entry.Name(), "marketplace.json")

		content, err := os.ReadFile(skillFile)
		if err != nil {
			continue
		}

		skill := SkillEntry{
			ID:      entry.Name(),
			Name:    entry.Name(),
			Content: string(content),
		}

		// Load install metadata if available
		metaData, err := os.ReadFile(metaFile)
		if err == nil {
			var meta map[string]any
			if json.Unmarshal(metaData, &meta) == nil {
				if v, ok := meta["version"].(string); ok {
					skill.Version = v
				}
				if v, ok := meta["author"].(string); ok {
					skill.Author = v
				}
				if v, ok := meta["installed_at"].(float64); ok {
					skill.CreatedAt = int64(v)
				}
			}
		}

		results = append(results, skill)
	}

	return results, nil
}

// ─── Remote Operations ──────────────────────────────────────────────────────

func (s *Service) fetchRemoteList(query string, tags []string, limit int) ([]SkillEntry, error) {
	url := fmt.Sprintf("%s?q=%s&limit=%d", s.cfg.RegistryURL, query, limit)
	if len(tags) > 0 {
		url += "&tags=" + strings.Join(tags, ",")
	}

	resp, err := s.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned %d", resp.StatusCode)
	}

	var entries []SkillEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, err
	}

	// Update cache
	s.mu.Lock()
	s.cache = entries
	s.cacheAt = time.Now()
	s.mu.Unlock()

	return entries, nil
}

func (s *Service) fetchRemoteGet(skillID string) (*SkillEntry, error) {
	url := fmt.Sprintf("%s/%s", s.cfg.RegistryURL, skillID)

	resp, err := s.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned %d", resp.StatusCode)
	}

	var entry SkillEntry
	if err := json.NewDecoder(resp.Body).Decode(&entry); err != nil {
		return nil, err
	}

	return &entry, nil
}

func (s *Service) publishRemote(entry *SkillEntry) error {
	payload, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/publish", s.cfg.RegistryURL)
	resp, err := s.client.Post(url, "application/json", strings.NewReader(string(payload)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("registry returned %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (s *Service) downloadContent(url string) ([]byte, error) {
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download returned %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// ─── Local Operations ───────────────────────────────────────────────────────

func (s *Service) getLocalSkills(query string, tags []string, limit int) []SkillEntry {
	s.mu.RLock()
	cached := s.cache
	cacheAt := s.cacheAt
	s.mu.RUnlock()

	// Use cache if fresh
	if len(cached) > 0 && time.Since(cacheAt) < s.cfg.CacheTTL {
		return s.filterSkills(cached, query, tags, limit)
	}

	// Scan local marketplace directory
	installed, err := s.Installed()
	if err != nil {
		return s.filterSkills(cached, query, tags, limit)
	}

	return s.filterSkills(installed, query, tags, limit)
}

func (s *Service) filterSkills(entries []SkillEntry, query string, tags []string, limit int) []SkillEntry {
	q := strings.ToLower(strings.TrimSpace(query))
	tagSet := make(map[string]bool, len(tags))
	for _, t := range tags {
		tagSet[strings.ToLower(t)] = true
	}

	var filtered []SkillEntry
	for _, entry := range entries {
		if q != "" {
			haystack := strings.ToLower(entry.ID + " " + entry.Name + " " + entry.Description)
			if !strings.Contains(haystack, q) {
				continue
			}
		}

		if len(tagSet) > 0 {
			found := false
			for _, tag := range entry.Tags {
				if tagSet[strings.ToLower(tag)] {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		filtered = append(filtered, entry)
		if len(filtered) >= limit {
			break
		}
	}

	return filtered
}

func (s *Service) enrichContent(entry *SkillEntry, content string) string {
	// If content already has frontmatter, leave it
	if strings.HasPrefix(strings.TrimSpace(content), "---\n") {
		return content
	}

	// Add frontmatter with marketplace metadata
	frontmatter := fmt.Sprintf("---\nname: %s\ndescription: %s\nauthor: %s\nversion: %s\nmarketplace: true\n---\n",
		entry.Name, entry.Description, entry.Author, entry.Version)

	return frontmatter + content
}
