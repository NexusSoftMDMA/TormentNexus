package skillregistry

import (
	"fmt"
	"sort"
	"strings"
	"sync"
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

// RankedSkill wraps SkillInfo with a relevance score.
type RankedSkill struct {
	SkillInfo
	Score       float64            `json:"score"`
	Rank        int                `json:"rank"`
	MatchReason string             `json:"matchReason"`
	Breakdown   map[string]float64 `json:"scoreBreakdown"`
}

// Search performs a ranked discovery search.
func (sr *SkillRegistry) Search(query string, limit int) []RankedSkill {
	return sr.SearchWithProfile(query, limit, "")
}

// SearchWithProfile performs a ranked discovery search with profile-based boosting.
func (sr *SkillRegistry) SearchWithProfile(query string, limit int, profile SkillProfile) []RankedSkill {
	if limit <= 0 {
		limit = 10
	}

	sr.mu.RLock()
	defer sr.mu.RUnlock()

	if query == "" {
		var results []RankedSkill
		all := sr.List()
		for i, s := range all {
			if i >= limit {
				break
			}
			results = append(results, RankedSkill{
				SkillInfo: s,
				Score:     0,
				Rank:      i + 1,
			})
		}
		return results
	}

	tokens := tokenize(query)
	var ranked []RankedSkill

	for _, s := range sr.skills {
		score, breakdown, reason := sr.calculateSkillScore(tokens, s, profile)
		if score > 0 {
			ranked = append(ranked, RankedSkill{
				SkillInfo:   *s,
				Score:       score,
				MatchReason: reason,
				Breakdown:   breakdown,
			})
		}
	}

	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].Score == ranked[j].Score {
			return ranked[i].Name < ranked[j].Name
		}
		return ranked[i].Score > ranked[j].Score
	})

	if len(ranked) > limit {
		ranked = ranked[:limit]
	}

	for i := range ranked {
		ranked[i].Rank = i + 1
	}

	return ranked
}

func tokenize(text string) []string {
	f := func(c rune) bool {
		return (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9')
	}
	parts := strings.FieldsFunc(strings.ToLower(text), f)
	var tokens []string
	for _, p := range parts {
		if len(p) >= 2 {
			tokens = append(tokens, p)
		}
	}
	return tokens
}

func (sr *SkillRegistry) calculateSkillScore(queryTokens []string, skill *SkillInfo, profile SkillProfile) (float64, map[string]float64, string) {
	score := 0.0
	breakdown := make(map[string]float64)
	var reasons []string

	nameTokens := tokenize(skill.Name + " " + skill.ID)
	descTokens := tokenize(skill.Description)

	// Weights
	const weightName = 10.0
	const weightDesc = 3.0
	const weightTag = 5.0
	const weightProfile = 15.0

	for _, q := range queryTokens {
		// Exact ID/Name match
		if strings.ToLower(skill.ID) == q || strings.ToLower(skill.Name) == q {
			score += 20.0
			breakdown["exact_match"] += 20.0
			reasons = append(reasons, "Exact match")
		}

		// Token matches in Name
		nameHit := 0.0
		for _, nt := range nameTokens {
			if nt == q {
				nameHit += weightName
			} else if strings.Contains(nt, q) {
				nameHit += weightName / 2
			}
		}
		if nameHit > 0 {
			score += nameHit
			breakdown["name"] += nameHit
			reasons = append(reasons, "Matched name")
		}

		// Token matches in Description
		descHit := 0.0
		for _, dt := range descTokens {
			if dt == q {
				descHit += weightDesc
			} else if strings.Contains(dt, q) {
				descHit += weightDesc / 2
			}
		}
		if descHit > 0 {
			score += descHit
			breakdown["description"] += descHit
			reasons = append(reasons, "Matched description")
		}

		// Tag matches
		tagHit := 0.0
		for _, tag := range skill.Tags {
			tagLower := strings.ToLower(tag)
			if tagLower == q {
				tagHit += weightTag
			} else if strings.Contains(tagLower, q) {
				tagHit += weightTag / 2
			}
		}
		if tagHit > 0 {
			score += tagHit
			breakdown["tags"] += tagHit
			reasons = append(reasons, "Matched tags")
		}
	}

	// Profile-based boosting
	if profile != "" {
		profileMatch := false
		switch profile {
		case ProfileRepoCoding:
			if strings.Contains(strings.ToLower(skill.Category), "code") || strings.Contains(strings.ToLower(skill.Category), "repo") {
				profileMatch = true
			}
		case ProfileWebResearch:
			if strings.Contains(strings.ToLower(skill.Category), "web") || strings.Contains(strings.ToLower(skill.Category), "search") {
				profileMatch = true
			}
		case ProfileKernelOps:
			if strings.Contains(strings.ToLower(skill.Category), "kernel") || strings.Contains(strings.ToLower(skill.Category), "sys") {
				profileMatch = true
			}
		}

		if profileMatch {
			score += weightProfile
			breakdown["profile_boost"] = weightProfile
			reasons = append(reasons, fmt.Sprintf("Boosted for profile: %s", profile))
		}
	}

	// Deduplicate reasons
	seen := make(map[string]bool)
	var uniqueReasons []string
	for _, r := range reasons {
		if !seen[r] {
			uniqueReasons = append(uniqueReasons, r)
			seen[r] = true
		}
	}

	return score, breakdown, strings.Join(uniqueReasons, "; ")
}

func (sr *SkillRegistry) Unregister(name string) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	delete(sr.skills, strings.ToLower(name))
}
