package skillregistry

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/robertpelloni/borg/internal/search"
)

// SkillInfo describes a registered skill.
type SkillInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Content     string   `json:"content,omitempty"`
	Category    string   `json:"category,omitempty"`
	AlwaysOn    bool     `json:"alwaysOn"`
	Tags        []string `json:"tags,omitempty"`
	Path        string   `json:"path,omitempty"`
}

// ScoredSkill represents a skill with its search score.
type ScoredSkill struct {
	SkillInfo
	Score          float64            `json:"score"`
	ScoreBreakdown map[string]float64 `json:"scoreBreakdown,omitempty"`
	MatchReason    string             `json:"matchReason"`
	Rank           int                `json:"rank"`
}

// SkillRegistry manages the global skill inventory.
type SkillRegistry struct {
	mu     sync.RWMutex
	skills map[string]*SkillInfo
}

// NewSkillRegistry creates a new empty registry.
func NewSkillRegistry() *SkillRegistry {
	return &SkillRegistry{
		skills: make(map[string]*SkillInfo),
	}
}

// Register adds or updates a skill in the registry.
func (sr *SkillRegistry) Register(skill SkillInfo) error {
	if skill.ID == "" {
		return fmt.Errorf("skill ID cannot be empty")
	}
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.skills[strings.ToLower(skill.ID)] = &skill
	return nil
}

// Get returns a skill by ID.
func (sr *SkillRegistry) Get(id string) (*SkillInfo, bool) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	s, ok := sr.skills[strings.ToLower(id)]
	if !ok {
		return nil, false
	}
	copy := *s
	return &copy, true
}

// List returns all registered skills.
func (sr *SkillRegistry) List() []SkillInfo {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	result := make([]SkillInfo, 0, len(sr.skills))
	for _, s := range sr.skills {
		result = append(result, *s)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// Search performs a ranked fuzzy search across skill names and descriptions.
// Implement search.Scorable interface
func (s *SkillInfo) GetName() string        { return s.Name }
func (s *SkillInfo) GetDescription() string { return s.Description }
func (s *SkillInfo) GetTags() []string      { return s.Tags }

// Search performs a ranked fuzzy search across skill names and descriptions.
func (sr *SkillRegistry) Search(query string, limit int) []ScoredSkill {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	var scorable []search.Scorable
	for _, s := range sr.skills {
		scorable = append(scorable, s)
	}

	ranked := search.RankItems(query, scorable, limit)

	var results []ScoredSkill
	for _, r := range ranked {
		results = append(results, ScoredSkill{
			SkillInfo:      *(r.Item.(*SkillInfo)),
			Score:          r.Score,
			ScoreBreakdown: r.ScoreBreakdown,
			MatchReason:    r.MatchReason,
			Rank:           r.Rank,
		})
	}

	return results
}
