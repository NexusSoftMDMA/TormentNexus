package memory

/**
 * @file manager.go
 * @module go/internal/memory
 *
 * WHAT: Unified memory management — short-term, medium-term, and long-term
 *       memory storage with automatic categorization, pruning, and retrieval.
 *
 * WHY: All AI models need persistent memory. This provides a unified interface
 *      for storing, retrieving, and managing memories with configurable backends.
 *
 * ADDED: v1.0.0-alpha.32
 */

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type MemoryKind string

const (
	KindFact         MemoryKind = "fact"
	KindPreference   MemoryKind = "preference"
	KindDecision     MemoryKind = "decision"
	KindPattern      MemoryKind = "pattern"
	KindSkill        MemoryKind = "skill"
	KindError        MemoryKind = "error"
	KindSolution     MemoryKind = "solution"
	KindContext      MemoryKind = "context"
	KindConversation MemoryKind = "conversation"
	KindProject      MemoryKind = "project"
)

type MemoryTier string

const (
	TierShortTerm  MemoryTier = "short"  // Current session, high detail
	TierMediumTerm MemoryTier = "medium" // Recent sessions, summarized
	TierLongTerm   MemoryTier = "long"   // Permanent facts and patterns
)

type Memory struct {
	ID             string            `json:"id"`
	Kind           MemoryKind        `json:"kind"`
	Tier           MemoryTier        `json:"tier"`
	Content        string            `json:"content"`
	Summary        string            `json:"summary,omitempty"`
	Tags           []string          `json:"tags,omitempty"`
	Source         string            `json:"source,omitempty"`
	Project        string            `json:"project,omitempty"`
	SessionID      string            `json:"sessionId,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	CreatedAt      time.Time         `json:"createdAt"`
	AccessedAt     time.Time         `json:"accessedAt"`
	AccessCount    int               `json:"accessCount"`
	RelevanceScore float64           `json:"relevanceScore"`
	Embedded       bool              `json:"embedded"` // True if vector embedding exists
}

type MemoryQuery struct {
	Query    string     `json:"query,omitempty"`
	Kind     MemoryKind `json:"kind,omitempty"`
	Tier     MemoryTier `json:"tier,omitempty"`
	Tags     []string   `json:"tags,omitempty"`
	Project  string     `json:"project,omitempty"`
	Source   string     `json:"source,omitempty"`
	Limit    int        `json:"limit,omitempty"`
	MinScore float64    `json:"minScore,omitempty"`
}

type MemoryManagerConfig struct {
	StorePath     string `json:"storePath"`
	MaxShortTerm  int    `json:"maxShortTerm"`
	MaxMediumTerm int    `json:"maxMediumTerm"`
	MaxLongTerm   int    `json:"maxLongTerm"`
}

func DefaultMemoryManagerConfig() MemoryManagerConfig {
	return MemoryManagerConfig{
		MaxShortTerm:  1000,
		MaxMediumTerm: 5000,
		MaxLongTerm:   50000,
	}
}

type MemoryManager struct {
	cfg       MemoryManagerConfig
	mu        sync.RWMutex
	memories  map[string]*Memory
	byKind    map[MemoryKind][]string // kind → memory IDs
	byTier    map[MemoryTier][]string // tier → memory IDs
	byTag     map[string][]string     // tag → memory IDs
	byProject map[string][]string     // project → memory IDs
}

func NewMemoryManager(cfg MemoryManagerConfig) *MemoryManager {
	if cfg.MaxShortTerm <= 0 {
		cfg.MaxShortTerm = 1000
	}
	if cfg.MaxMediumTerm <= 0 {
		cfg.MaxMediumTerm = 5000
	}
	if cfg.MaxLongTerm <= 0 {
		cfg.MaxLongTerm = 50000
	}

	mm := &MemoryManager{
		cfg:       cfg,
		memories:  make(map[string]*Memory),
		byKind:    make(map[MemoryKind][]string),
		byTier:    make(map[MemoryTier][]string),
		byTag:     make(map[string][]string),
		byProject: make(map[string][]string),
	}

	// Load from disk if available
	if cfg.StorePath != "" {
		_ = mm.loadFromDisk()
	}

	return mm
}

// Store adds a new memory.
func (mm *MemoryManager) Store(memory Memory) (string, error) {
	if memory.ID == "" {
		memory.ID = generateMemoryID(memory.Content, memory.Source)
	}

	now := time.Now().UTC()
	if memory.CreatedAt.IsZero() {
		memory.CreatedAt = now
	}
	memory.AccessedAt = now

	mm.mu.Lock()
	defer mm.mu.Unlock()

	// Check for duplicate
	if existing, ok := mm.memories[memory.ID]; ok {
		// Update existing
		existing.AccessedAt = now
		existing.AccessCount++
		existing.RelevanceScore = max(existing.RelevanceScore, memory.RelevanceScore)
		return existing.ID, nil
	}

	mm.memories[memory.ID] = &memory

	// Update indices
	mm.byKind[memory.Kind] = append(mm.byKind[memory.Kind], memory.ID)
	mm.byTier[memory.Tier] = append(mm.byTier[memory.Tier], memory.ID)
	for _, tag := range memory.Tags {
		mm.byTag[strings.ToLower(tag)] = append(mm.byTag[strings.ToLower(tag)], memory.ID)
	}
	if memory.Project != "" {
		mm.byProject[memory.Project] = append(mm.byProject[memory.Project], memory.ID)
	}

	return memory.ID, nil
}

// Retrieve finds memories matching the query.
func (mm *MemoryManager) Retrieve(query MemoryQuery) ([]*Memory, error) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	var candidates []*Memory
	candidateSet := make(map[string]bool)

	// Filter by kind
	if query.Kind != "" {
		for _, id := range mm.byKind[query.Kind] {
			if !candidateSet[id] && mm.memories[id] != nil {
				candidates = append(candidates, mm.memories[id])
				candidateSet[id] = true
			}
		}
	}

	// Filter by tier
	if query.Tier != "" {
		for _, id := range mm.byTier[query.Tier] {
			if !candidateSet[id] && mm.memories[id] != nil {
				candidates = append(candidates, mm.memories[id])
				candidateSet[id] = true
			}
		}
	}

	// Filter by tags
	for _, tag := range query.Tags {
		for _, id := range mm.byTag[strings.ToLower(tag)] {
			if !candidateSet[id] && mm.memories[id] != nil {
				candidates = append(candidates, mm.memories[id])
				candidateSet[id] = true
			}
		}
	}

	// Filter by project
	if query.Project != "" {
		for _, id := range mm.byProject[query.Project] {
			if !candidateSet[id] && mm.memories[id] != nil {
				candidates = append(candidates, mm.memories[id])
				candidateSet[id] = true
			}
		}
	}

	// If no filters, use all
	if len(candidates) == 0 && query.Kind == "" && query.Tier == "" && len(query.Tags) == 0 && query.Project == "" {
		for _, m := range mm.memories {
			candidates = append(candidates, m)
		}
	}

	// Score by text relevance if query provided
	if query.Query != "" {
		queryTokens := tokenize(query.Query)
		for _, m := range candidates {
			contentTokens := tokenize(m.Content + " " + m.Summary)
			score := 0.0
			for _, qt := range queryTokens {
				for _, ct := range contentTokens {
					if qt == ct {
						score += 3.0
					} else if strings.Contains(ct, qt) {
						score += 1.0
					}
				}
			}
			// Boost by access count and recency
			score += float64(m.AccessCount) * 0.5
			hoursSinceAccess := time.Since(m.AccessedAt).Hours()
			if hoursSinceAccess < 1 {
				score += 2.0
			} else if hoursSinceAccess < 24 {
				score += 1.0
			}
			m.RelevanceScore = score
		}

		// Filter by min score
		if query.MinScore > 0 {
			var filtered []*Memory
			for _, m := range candidates {
				if m.RelevanceScore >= query.MinScore {
					filtered = append(filtered, m)
				}
			}
			candidates = filtered
		}
	}

	// Sort by relevance
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].RelevanceScore > candidates[j].RelevanceScore
	})

	// Apply limit
	limit := query.Limit
	if limit <= 0 {
		limit = 20
	}
	if len(candidates) > limit {
		candidates = candidates[:limit]
	}

	// Update access stats
	for _, m := range candidates {
		m.AccessedAt = time.Now().UTC()
		m.AccessCount++
	}

	return candidates, nil
}

// Get retrieves a single memory by ID.
func (mm *MemoryManager) Get(id string) (*Memory, bool) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	m, ok := mm.memories[id]
	if ok {
		m.AccessedAt = time.Now().UTC()
		m.AccessCount++
	}
	return m, ok
}

// Delete removes a memory.
func (mm *MemoryManager) Delete(id string) bool {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	m, ok := mm.memories[id]
	if !ok {
		return false
	}

	delete(mm.memories, id)

	// Remove from indices
	mm.byKind[m.Kind] = removeFromSlice(mm.byKind[m.Kind], id)
	mm.byTier[m.Tier] = removeFromSlice(mm.byTier[m.Tier], id)
	for _, tag := range m.Tags {
		mm.byTag[strings.ToLower(tag)] = removeFromSlice(mm.byTag[strings.ToLower(tag)], id)
	}
	if m.Project != "" {
		mm.byProject[m.Project] = removeFromSlice(mm.byProject[m.Project], id)
	}

	return true
}

// Prune removes low-relevance memories to stay within limits.
func (mm *MemoryManager) Prune() int {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	pruned := 0

	// Prune short-term first
	pruned += mm.pruneTier(TierShortTerm, mm.cfg.MaxShortTerm)
	pruned += mm.pruneTier(TierMediumTerm, mm.cfg.MaxMediumTerm)
	// Don't prune long-term automatically

	return pruned
}

// Demote moves old short-term memories to medium-term (summarization point).
func (mm *MemoryManager) Demote(maxAge time.Duration) int {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	demoted := 0
	cutoff := time.Now().UTC().Add(-maxAge)

	for _, id := range mm.byTier[TierShortTerm] {
		m, ok := mm.memories[id]
		if !ok {
			continue
		}
		if m.CreatedAt.Before(cutoff) {
			mm.byTier[TierShortTerm] = removeFromSlice(mm.byTier[TierShortTerm], id)
			m.Tier = TierMediumTerm
			mm.byTier[TierMediumTerm] = append(mm.byTier[TierMediumTerm], id)
			demoted++
		}
	}

	return demoted
}

// Promote moves frequently-accessed medium-term memories to long-term.
func (mm *MemoryManager) Promote(minAccessCount int) int {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	promoted := 0

	for _, id := range mm.byTier[TierMediumTerm] {
		m, ok := mm.memories[id]
		if !ok {
			continue
		}
		if m.AccessCount >= minAccessCount {
			mm.byTier[TierMediumTerm] = removeFromSlice(mm.byTier[TierMediumTerm], id)
			m.Tier = TierLongTerm
			mm.byTier[TierLongTerm] = append(mm.byTier[TierLongTerm], id)
			promoted++
		}
	}

	return promoted
}

// Count returns the total number of memories.
func (mm *MemoryManager) Count() int {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	return len(mm.memories)
}

// CountByTier returns memory counts per tier.
func (mm *MemoryManager) CountByTier() map[MemoryTier]int {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	counts := make(map[MemoryTier]int)
	for tier, ids := range mm.byTier {
		counts[tier] = len(ids)
	}
	return counts
}

// Save persists memories to disk.
func (mm *MemoryManager) Save() error {
	if mm.cfg.StorePath == "" {
		return nil
	}

	mm.mu.RLock()
	defer mm.mu.RUnlock()

	return mm.saveToDiskLocked()
}

// Stats returns memory system statistics.
func (mm *MemoryManager) Stats() map[string]interface{} {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	tierCounts := make(map[MemoryTier]int)
	kindCounts := make(map[MemoryKind]int)
	for _, m := range mm.memories {
		tierCounts[m.Tier]++
		kindCounts[m.Kind]++
	}

	return map[string]interface{}{
		"total":    len(mm.memories),
		"byTier":   tierCounts,
		"byKind":   kindCounts,
		"tags":     len(mm.byTag),
		"projects": len(mm.byProject),
	}
}

// --- Internal ---

func (mm *MemoryManager) pruneTier(tier MemoryTier, maxCount int) int {
	ids := mm.byTier[tier]
	if len(ids) <= maxCount {
		return 0
	}

	// Sort by relevance (ascending) and remove the weakest
	type idScore struct {
		id    string
		score float64
	}
	var scored []idScore
	for _, id := range ids {
		if m, ok := mm.memories[id]; ok {
			scored = append(scored, idScore{id, m.RelevanceScore})
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score < scored[j].score
	})

	toRemove := len(ids) - maxCount
	pruned := 0
	for i := 0; i < toRemove && i < len(scored); i++ {
		id := scored[i].id
		delete(mm.memories, id)
		mm.byTier[tier] = removeFromSlice(mm.byTier[tier], id)
		pruned++
	}

	return pruned
}

func removeFromSlice(slice []string, target string) []string {
	var filtered []string
	for _, s := range slice {
		if s != target {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

func (mm *MemoryManager) saveToDiskLocked() error {
	if err := os.MkdirAll(filepath.Dir(mm.cfg.StorePath), 0o755); err != nil {
		return err
	}

	var memories []*Memory
	for _, m := range mm.memories {
		memories = append(memories, m)
	}

	data, err := json.MarshalIndent(memories, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(mm.cfg.StorePath, data, 0o644)
}

func (mm *MemoryManager) loadFromDisk() error {
	data, err := os.ReadFile(mm.cfg.StorePath)
	if err != nil {
		return err
	}

	var memories []*Memory
	if err := json.Unmarshal(data, &memories); err != nil {
		return err
	}

	for _, m := range memories {
		mm.memories[m.ID] = m
		mm.byKind[m.Kind] = append(mm.byKind[m.Kind], m.ID)
		mm.byTier[m.Tier] = append(mm.byTier[m.Tier], m.ID)
		for _, tag := range m.Tags {
			mm.byTag[strings.ToLower(tag)] = append(mm.byTag[strings.ToLower(tag)], m.ID)
		}
		if m.Project != "" {
			mm.byProject[m.Project] = append(mm.byProject[m.Project], m.ID)
		}
	}

	return nil
}

func generateMemoryID(content, source string) string {
	h := sha256.Sum256([]byte(content + source))
	return fmt.Sprintf("mem_%x", h[:12])
}

func tokenize(text string) []string {
	return strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9')
	})
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
