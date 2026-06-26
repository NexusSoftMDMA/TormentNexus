package skillregistry

import (
	"context"
	"strings"
	"testing"
)

func TestSkillDecisionProgressiveLoading(t *testing.T) {
	reg := NewSkillRegistry()
	_ = reg.Register(SkillInfo{ID: "skill1", Name: "Skill One", Description: "Testing skill one"})
	_ = reg.Register(SkillInfo{ID: "skill2", Name: "Skill Two", Description: "Testing skill two"})

	ds := NewSkillDecisionSystem(SkillDecisionConfig{SoftCap: 1, HardCap: 1}, reg)
	ctx := context.Background()

	// Load first skill
	err := ds.LoadSkill(ctx, "skill1", false)
	if err != nil {
		t.Fatalf("failed to load skill1: %v", err)
	}

	loaded := ds.ListLoadedSkills()
	if len(loaded) != 1 || strings.ToLower(loaded[0].ID) != "skill1" {
		t.Errorf("expected 1 loaded skill (skill1), got %d", len(loaded))
	}

	// Load second skill, should evict skill1
	err = ds.LoadSkill(ctx, "skill2", false)
	if err != nil {
		t.Fatalf("failed to load skill2: %v", err)
	}

	loaded = ds.ListLoadedSkills()
	if len(loaded) != 1 || strings.ToLower(loaded[0].ID) != "skill2" {
		t.Errorf("expected 2nd skill to replace first, got %+v", loaded)
	}
}

func TestSkillDecisionSearch(t *testing.T) {
	reg := NewSkillRegistry()
	_ = reg.Register(SkillInfo{ID: "react", Name: "React Native", Description: "Frontend UI"})
	_ = reg.Register(SkillInfo{ID: "gin", Name: "Gin Gonic", Description: "Go Backend"})

	ds := NewSkillDecisionSystem(DefaultSkillDecisionConfig(), reg)
	ctx := context.Background()

	results, err := ds.SearchSkills(ctx, "React")
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	if len(results) == 0 || results[0].ID != "react" {
		t.Errorf("expected react to be top result, got %+v", results)
	}
}
