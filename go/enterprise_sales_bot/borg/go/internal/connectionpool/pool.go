package connectionpool

import (
	"fmt"
	"sync"
	"time"
)

type PoolEntry struct {
	ID        string            `json:"id"`
	CreatedAt time.Time         `json:"createdAt"`
	InUse     bool              `json:"inUse"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type Service struct {
	mu   sync.RWMutex
	pool map[string]*PoolEntry
}

func NewService() *Service {
	return &Service{pool: make(map[string]*PoolEntry)}
}

func (s *Service) Acquire(id string) (*PoolEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if entry, ok := s.pool[id]; ok {
		entry.InUse = true
		entry.CreatedAt = time.Now()
		return entry, nil
	}
	entry := &PoolEntry{ID: id, CreatedAt: time.Now(), InUse: true}
	s.pool[id] = entry
	return entry, nil
}

func (s *Service) Release(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if entry, ok := s.pool[id]; ok {
		entry.InUse = false
	}
}

func (s *Service) Stats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	total := len(s.pool)
	inUse := 0
	for _, e := range s.pool {
		if e.InUse {
			inUse++
		}
	}
	return map[string]interface{}{
		"total": total,
		"inUse": inUse,
		"idle":  total - inUse,
	}
}

func (s *Service) Get(id string) (*PoolEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if entry, ok := s.pool[id]; ok {
		return entry, nil
	}
	return nil, fmt.Errorf("connection not found: %s", id)
}
