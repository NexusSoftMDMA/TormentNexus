package catalogingestor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type CatalogEntry struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	IngestedAt  time.Time `json:"ingestedAt"`
	Source      string    `json:"source"`
}

type Service struct {
	mu       sync.RWMutex
	entries  map[string]*CatalogEntry
	filePath string
}

func NewService(filePath string) *Service {
	s := &Service{
		entries:  make(map[string]*CatalogEntry),
		filePath: filePath,
	}
	s.load()
	return s
}

func (s *Service) Ingest(entry *CatalogEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entry.ID == "" {
		return fmt.Errorf("entry ID is required")
	}
	entry.IngestedAt = time.Now()
	s.entries[entry.ID] = entry
	return s.save()
}

func (s *Service) IngestBatch(entries []*CatalogEntry) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	imported := 0
	for _, entry := range entries {
		if entry.ID != "" {
			entry.IngestedAt = time.Now()
			s.entries[entry.ID] = entry
			imported++
		}
	}
	return imported, s.save()
}

func (s *Service) List() []*CatalogEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*CatalogEntry, 0, len(s.entries))
	for _, e := range s.entries {
		result = append(result, e)
	}
	return result
}

func (s *Service) Get(id string) *CatalogEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.entries[id]
}

func (s *Service) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.entries)
}

func (s *Service) load() {
	if s.filePath == "" {
		return
	}
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return
	}
	var entries []*CatalogEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return
	}
	for _, e := range entries {
		s.entries[e.ID] = e
	}
}

func (s *Service) save() error {
	if s.filePath == "" {
		return nil
	}
	dir := filepath.Dir(s.filePath)
	os.MkdirAll(dir, 0755)

	entries := make([]*CatalogEntry, 0, len(s.entries))
	for _, e := range s.entries {
		entries = append(entries, e)
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0644)
}
