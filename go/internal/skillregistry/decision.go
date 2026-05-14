package skillregistry

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type SkillLoaded struct {
	SkillInfo
	LoadedAt   time.Time `json:"loadedAt"`
	LastUsedAt time.Time `json:"lastUsedAt"`
	UseCount   int       `json:"useCount"`
	AutoLoaded bool      `json:"autoLoaded"`
}

type SkillDecisionConfig struct {
	SoftCap                 int           `json:"softCap"`
	HardCap                 int           `json:"hardCap"`
	HighConfidenceThreshold float64       `json:"highConfidenceThreshold"`
	IdleTimeout             time.Duration `json:"idleTimeout"`
}

func DefaultSkillDecisionConfig() SkillDecisionConfig {
	return SkillDecisionConfig{
		SoftCap:                 10,
		HardCap:                 20,
		HighConfidenceThreshold: 15.0, // Match search scoring
		IdleTimeout:             30 * time.Minute,
	}
}

type SkillDecisionEvent struct {
	Timestamp time.Time              `json:"timestamp"`
	Type      string                 `json:"type"` // "search", "load", "unload", "evict"
	SkillID   string                 `json:"skillId,omitempty"`
	Query     string                 `json:"query,omitempty"`
	Score     float64                `json:"score,omitempty"`
	Reason    string                 `json:"reason,omitempty"`
	Success   bool                   `json:"success"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type SkillDecisionSystem struct {
	cfg      SkillDecisionConfig
	mu       sync.RWMutex
	loaded   map[string]*SkillLoaded
	registry *SkillRegistry
	events   []SkillDecisionEvent
	eventIdx int
}

func NewSkillDecisionSystem(cfg SkillDecisionConfig, registry *SkillRegistry) *SkillDecisionSystem {
	return &SkillDecisionSystem{
		cfg:      cfg,
		loaded:   make(map[string]*SkillLoaded),
		registry: registry,
		events:   make([]SkillDecisionEvent, 100),
	}
}

func (ds *SkillDecisionSystem) SearchSkills(ctx context.Context, query string) ([]ScoredSkill, error) {
	results := ds.registry.Search(query, ds.cfg.SoftCap)

	ds.recordEvent(SkillDecisionEvent{
		Type:    "search",
		Query:   query,
		Success: true,
		Metadata: map[string]interface{}{
			"resultCount": len(results),
		},
	})

	// Auto-load high confidence matches
	for _, r := range results {
		if r.Score >= ds.cfg.HighConfidenceThreshold {
			_ = ds.loadSkillInternal(r.ID, true)
		}
	}

	return results, nil
}

func (ds *SkillDecisionSystem) LoadSkill(id string) error {
	return ds.loadSkillInternal(id, false)
}

func (ds *SkillDecisionSystem) loadSkillInternal(id string, auto bool) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	id = strings.ToLower(id)
	if sl, ok := ds.loaded[id]; ok {
		sl.LastUsedAt = time.Now()
		sl.UseCount++
		return nil
	}

	skill, ok := ds.registry.Get(id)
	if !ok {
		ds.recordEventLocked(SkillDecisionEvent{
			Type:    "load",
			SkillID: id,
			Success: false,
			Reason:  "not found",
		})
		return fmt.Errorf("skill %s not found in registry", id)
	}

	// Evict if needed
	ds.evictIfNeededLocked()

	ds.loaded[id] = &SkillLoaded{
		SkillInfo:  *skill,
		LoadedAt:   time.Now(),
		LastUsedAt: time.Now(),
		UseCount:   1,
		AutoLoaded: auto,
	}

	ds.recordEventLocked(SkillDecisionEvent{
		Type:    "load",
		SkillID: id,
		Success: true,
		Reason:  fmt.Sprintf("auto=%v", auto),
	})

	return nil
}

func (ds *SkillDecisionSystem) evictIfNeededLocked() {
	if len(ds.loaded) >= ds.cfg.HardCap {
		ds.evictLRULocked()
	}

	if len(ds.loaded) >= ds.cfg.SoftCap {
		now := time.Now()
		var oldest string
		var oldestTime time.Time

		for id, sl := range ds.loaded {
			if sl.AlwaysOn {
				continue
			}
			if now.Sub(sl.LastUsedAt) > ds.cfg.IdleTimeout {
				if oldest == "" || sl.LastUsedAt.Before(oldestTime) {
					oldest = id
					oldestTime = sl.LastUsedAt
				}
			}
		}

		if oldest != "" {
			ds.recordEventLocked(SkillDecisionEvent{
				Type:    "evict",
				SkillID: oldest,
				Success: true,
				Reason:  "idle timeout",
			})
			delete(ds.loaded, oldest)
		}
	}
}

func (ds *SkillDecisionSystem) evictLRULocked() {
	var oldest string
	var oldestTime time.Time

	for id, sl := range ds.loaded {
		if sl.AlwaysOn {
			continue
		}
		if oldest == "" || sl.LastUsedAt.Before(oldestTime) {
			oldest = id
			oldestTime = sl.LastUsedAt
		}
	}

	if oldest != "" {
		ds.recordEventLocked(SkillDecisionEvent{
			Type:    "evict",
			SkillID: oldest,
			Success: true,
			Reason:  "hard cap",
		})
		delete(ds.loaded, oldest)
	}
}

func (ds *SkillDecisionSystem) recordEvent(event SkillDecisionEvent) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.recordEventLocked(event)
}

func (ds *SkillDecisionSystem) recordEventLocked(event SkillDecisionEvent) {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	ds.events[ds.eventIdx%len(ds.events)] = event
	ds.eventIdx++
}

func (ds *SkillDecisionSystem) ListLoadedSkills() []SkillInfo {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	var result []SkillInfo
	for _, sl := range ds.loaded {
		result = append(result, sl.SkillInfo)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

func (ds *SkillDecisionSystem) UnloadSkill(id string) bool {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	id = strings.ToLower(id)
	_, existed := ds.loaded[id]
	if existed {
		delete(ds.loaded, id)
	}
	return existed
}
