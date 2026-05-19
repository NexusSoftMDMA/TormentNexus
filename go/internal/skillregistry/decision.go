package skillregistry

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// SkillLoaded extends SkillInfo with usage tracking
type SkillLoaded struct {
	SkillInfo
	LoadedAt   time.Time `json:"loadedAt"`
	LastUsedAt time.Time `json:"lastUsedAt"`
	UseCount   int       `json:"useCount"`
	AutoLoaded bool      `json:"autoLoaded"`
}

type SkillDecisionConfig struct {
	// SoftCap is the soft limit on loaded skill metadata entries.
	SoftCap int `json:"softCap"`
	// HardCap is the hard limit; triggers aggressive LRU eviction.
	HardCap int `json:"hardCap"`
	// HighConfidenceThreshold: scores above this trigger silent auto-load.
	HighConfidenceThreshold float64 `json:"highConfidenceThreshold"`
	// SearchResultLimit: max results returned from search.
	SearchResultLimit int `json:"searchResultLimit"`
	// IdleTimeout: skills idle longer than this are candidates for eviction.
	IdleTimeout time.Duration `json:"idleTimeout"`
}

func DefaultSkillDecisionConfig() SkillDecisionConfig {
	return SkillDecisionConfig{
		SoftCap:                 10,
		HardCap:                 20,
		HighConfidenceThreshold: 15.0,
		SearchResultLimit:       5,
		IdleTimeout:             30 * time.Minute,
	}
}

type SkillProfile string

const (
	ProfileRepoCoding  SkillProfile = "repo-coding"
	ProfileWebResearch SkillProfile = "web-research"
	ProfileKernelOps   SkillProfile = "kernel-ops"
)

type SkillDecisionSystem struct {
	cfg      SkillDecisionConfig
	mu       sync.RWMutex
	loaded   map[string]*SkillLoaded // keyed by ID
	registry *SkillRegistry
	profile  SkillProfile
}

func NewSkillDecisionSystem(cfg SkillDecisionConfig, registry *SkillRegistry) *SkillDecisionSystem {
	return &SkillDecisionSystem{
		cfg:      cfg,
		loaded:   make(map[string]*SkillLoaded),
		registry: registry,
		profile:  ProfileRepoCoding,
	}
}

func (ds *SkillDecisionSystem) SetProfile(p SkillProfile) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.profile = p
}

func (ds *SkillDecisionSystem) SearchSkills(ctx context.Context, query string) ([]RankedSkill, error) {
	ds.mu.RLock()
	profile := ds.profile
	ds.mu.RUnlock()

	return ds.registry.SearchWithProfile(query, ds.cfg.SoftCap, profile), nil
}

// SearchAndLoad performs a ranked search and auto-loads high-confidence skills.
func (ds *SkillDecisionSystem) SearchAndLoad(ctx context.Context, query string) ([]RankedSkill, error) {
	ds.mu.RLock()
	profile := ds.profile
	ds.mu.RUnlock()

	ranked := ds.registry.SearchWithProfile(query, ds.cfg.SoftCap, profile)
	if len(ranked) == 0 {
		return nil, nil
	}

	// Auto-load high confidence results
	for _, r := range ranked {
		if r.Score >= ds.cfg.HighConfidenceThreshold {
			_ = ds.LoadSkill(ctx, r.ID, true)
		}
	}

	return ranked, nil
}

func (ds *SkillDecisionSystem) LoadSkill(ctx context.Context, id string, autoLoaded bool) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	return ds.loadSkillLocked(id, autoLoaded)
}

func (ds *SkillDecisionSystem) loadSkillLocked(id string, autoLoaded bool) error {
	id = strings.ToLower(id)
	if sl, ok := ds.loaded[id]; ok {
		sl.LastUsedAt = time.Now()
		sl.UseCount++
		return nil
	}

	skill, ok := ds.registry.Get(id)
	if !ok {
		return fmt.Errorf("skill %s not found in registry", id)
	}

	// Evict if needed
	for len(ds.loaded) >= ds.cfg.HardCap {
		ds.evictLRULocked()
	}

	ds.loaded[id] = &SkillLoaded{
		SkillInfo:  *skill,
		LoadedAt:   time.Now(),
		LastUsedAt: time.Now(),
		UseCount:   1,
		AutoLoaded: autoLoaded,
	}

	return nil
}

// RefreshAlwaysOn loads all skills marked as AlwaysOn into the active set.
func (ds *SkillDecisionSystem) RefreshAlwaysOn() {
	all := ds.registry.List()
	ds.mu.Lock()
	defer ds.mu.Unlock()

	for _, s := range all {
		if s.AlwaysOn {
			_ = ds.loadSkillLocked(s.ID, false)
		}
	}
}

func (ds *SkillDecisionSystem) ListLoadedSkills() []SkillLoaded {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	var result []SkillLoaded
	for _, sl := range ds.loaded {
		result = append(result, *sl)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].LastUsedAt.After(result[j].LastUsedAt)
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
		delete(ds.loaded, oldest)
	}
}

func (ds *SkillDecisionSystem) EvictIdle() int {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	now := time.Now()
	evicted := 0
	for id, sl := range ds.loaded {
		if sl.AlwaysOn {
			continue
		}
		if now.Sub(sl.LastUsedAt) > ds.cfg.IdleTimeout {
			delete(ds.loaded, id)
			evicted++
		}
	}
	return evicted
}
