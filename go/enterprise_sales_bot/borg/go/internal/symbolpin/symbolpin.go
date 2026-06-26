package symbolpin

import (
	"sync"
	"time"
)

type PinnedSymbol struct {
	Name       string    `json:"name"`
	FilePath   string    `json:"filePath"`
	LineNumber int       `json:"lineNumber"`
	PinnedAt   time.Time `json:"pinnedAt"`
	Notes      string    `json:"notes,omitempty"`
}

type Service struct {
	mu      sync.RWMutex
	symbols map[string]*PinnedSymbol
}

func NewService() *Service {
	return &Service{symbols: make(map[string]*PinnedSymbol)}
}

func (s *Service) Pin(name, filePath string, line int, notes string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.symbols[name] = &PinnedSymbol{
		Name: name, FilePath: filePath, LineNumber: line,
		PinnedAt: time.Now(), Notes: notes,
	}
}

func (s *Service) Unpin(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.symbols, name)
}

func (s *Service) List() []*PinnedSymbol {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*PinnedSymbol, 0, len(s.symbols))
	for _, sym := range s.symbols {
		result = append(result, sym)
	}
	return result
}

func (s *Service) Get(name string) *PinnedSymbol {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.symbols[name]
}

func (s *Service) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.symbols = make(map[string]*PinnedSymbol)
}
