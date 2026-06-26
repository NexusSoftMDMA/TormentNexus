package hsync

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

type SuggestionStatus string

const (
	StatusPending  SuggestionStatus = "PENDING"
	StatusApproved SuggestionStatus = "APPROVED"
	StatusRejected SuggestionStatus = "REJECTED"
)

type SuggestionType string

const (
	TypeAction SuggestionType = "ACTION"
	TypeInfo   SuggestionType = "INFO"
)

type Suggestion struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Type        SuggestionType   `json:"type"`
	Source      string           `json:"source"`
	Payload     interface{}      `json:"payload,omitempty"`
	Timestamp   int64            `json:"timestamp"`
	Status      SuggestionStatus `json:"status"`
}

type SuggestionService struct {
	persistencePath string
	suggestions     []Suggestion
	mu              sync.RWMutex
}

func NewSuggestionService(workspaceRoot string) *SuggestionService {
	s := &SuggestionService{
		persistencePath: filepath.Join(workspaceRoot, "packages", "core", "data", "suggestions.json"),
	}
	_ = s.Load()
	return s
}

func (s *SuggestionService) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.persistencePath)
	if err != nil {
		if os.IsNotExist(err) {
			s.suggestions = []Suggestion{}
			return nil
		}
		return err
	}

	return json.Unmarshal(data, &s.suggestions)
}

func (s *SuggestionService) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.saveNoLock()
}

func (s *SuggestionService) saveNoLock() error {
	// Prune old resolved items to keep file small (keep last 50)
	var active []Suggestion
	var history []Suggestion
	for _, sugg := range s.suggestions {
		if sugg.Status == StatusPending {
			active = append(active, sugg)
		} else {
			history = append(history, sugg)
		}
	}

	if len(history) > 50 {
		history = history[len(history)-50:]
	}

	toSave := append(active, history...)

	data, err := json.MarshalIndent(toSave, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(s.persistencePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(s.persistencePath, data, 0644)
}

func (s *SuggestionService) AddSuggestion(title, description, source string, payload interface{}) Suggestion {
	suggestion := Suggestion{
		ID:          uuid.New().String(),
		Title:       title,
		Description: description,
		Type:        TypeAction,
		Source:      source,
		Payload:     payload,
		Timestamp:   time.Now().UnixMilli(),
		Status:      StatusPending,
	}

	s.mu.Lock()
	s.suggestions = append(s.suggestions, suggestion)
	_ = s.saveNoLock()
	s.mu.Unlock()

	return suggestion
}

func (s *SuggestionService) GetPendingSuggestions() []Suggestion {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var pending []Suggestion
	for _, sugg := range s.suggestions {
		if sugg.Status == StatusPending {
			pending = append(pending, sugg)
		}
	}

	sort.Slice(pending, func(i, j int) bool {
		return pending[i].Timestamp > pending[j].Timestamp
	})

	return pending
}

func (s *SuggestionService) ResolveSuggestion(id string, status SuggestionStatus) (Suggestion, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, sugg := range s.suggestions {
		if sugg.ID == id {
			s.suggestions[i].Status = status
			suggCopy := s.suggestions[i]
			_ = s.saveNoLock()
			return suggCopy, nil
		}
	}

	return Suggestion{}, fmt.Errorf("suggestion %s not found", id)
}

func (s *SuggestionService) ClearAll() {
	s.mu.Lock()
	s.suggestions = []Suggestion{}
	_ = s.saveNoLock()
	s.mu.Unlock()
}

// Deprecated: Use SuggestionService methods directly
func ResolveSuggestion(id string, status string) (SuggestionsResult, error) {
	return SuggestionsResult{Success: true}, nil
}

// Deprecated: Use SuggestionService methods directly
func ClearAllSuggestions() (SuggestionsResult, error) {
	return SuggestionsResult{Success: true}, nil
}

type SuggestionsResult struct {
	Success bool `json:"success"`
}
