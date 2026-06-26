package skillregistry

import (
	"testing"
)

func TestSkillSearch(t *testing.T) {
	sr := NewSkillRegistry()
	sr.Register(SkillInfo{ID: "test-id", Name: "Test Skill", Description: "This is a test description", Tags: []string{"test"}})
	sr.Register(SkillInfo{ID: "another", Name: "Another One", Description: "Something else"})

	results := sr.Search("test", 10)
	if len(results) == 0 {
		t.Fatal("expected results, got none")
	}

	if results[0].ID != "test-id" {
		t.Errorf("expected test-id as first result, got %s", results[0].ID)
	}

	if results[0].Score <= 0 {
		t.Errorf("expected positive score, got %f", results[0].Score)
	}
}
