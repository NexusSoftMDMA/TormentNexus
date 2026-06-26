package config

import (
	"testing"
)

func TestSaveAndLoadState(t *testing.T) {
	dir := t.TempDir()
	state := &State{
		Projects: map[string]*ProjectState{
			"/home/user/project-a": {Skipped: []string{"server1", "server2"}},
		},
	}

	if err := SaveState(dir, state); err != nil {
		t.Fatalf("SaveState: %v", err)
	}

	loaded, err := LoadState(dir)
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}

	skipped := loaded.GetSkipped("/home/user/project-a")
	if len(skipped) != 2 {
		t.Fatalf("expected 2 skipped servers, got %d", len(skipped))
	}
	if skipped[0] != "server1" || skipped[1] != "server2" {
		t.Errorf("skipped = %v, want [server1 server2]", skipped)
	}
}

func TestLoadStateNotFound(t *testing.T) {
	dir := t.TempDir()
	state, err := LoadState(dir)
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	if state.Projects == nil {
		t.Fatal("expected initialized Projects map, got nil")
	}
	if len(state.Projects) != 0 {
		t.Errorf("expected empty Projects map, got %d entries", len(state.Projects))
	}
}

func TestIsSkipped(t *testing.T) {
	state := &State{
		Projects: map[string]*ProjectState{
			"/project": {Skipped: []string{"server-a", "server-b"}},
		},
	}

	if !state.IsSkipped("/project", "server-a") {
		t.Error("expected server-a to be skipped")
	}
	if !state.IsSkipped("/project", "server-b") {
		t.Error("expected server-b to be skipped")
	}
	if state.IsSkipped("/project", "server-c") {
		t.Error("expected server-c to NOT be skipped")
	}
	if state.IsSkipped("/other-project", "server-a") {
		t.Error("expected server-a to NOT be skipped for different project")
	}
}

func TestAddSkipped(t *testing.T) {
	state := &State{Projects: make(map[string]*ProjectState)}

	state.AddSkipped("/project", "server-x")
	if !state.IsSkipped("/project", "server-x") {
		t.Error("expected server-x to be skipped after AddSkipped")
	}

	// Adding duplicate should not create a second entry
	state.AddSkipped("/project", "server-x")
	skipped := state.GetSkipped("/project")
	if len(skipped) != 1 {
		t.Errorf("expected 1 skipped entry after duplicate add, got %d", len(skipped))
	}
}

func TestClearSkipped(t *testing.T) {
	state := &State{
		Projects: map[string]*ProjectState{
			"/project": {Skipped: []string{"server-a"}},
		},
	}

	state.ClearSkipped("/project")
	skipped := state.GetSkipped("/project")
	if len(skipped) != 0 {
		t.Errorf("expected empty skipped list after clear, got %v", skipped)
	}
}

func TestGetSkippedUnknownProject(t *testing.T) {
	state := &State{Projects: make(map[string]*ProjectState)}
	skipped := state.GetSkipped("/nonexistent")
	if skipped != nil {
		t.Errorf("expected nil for unknown project, got %v", skipped)
	}
}
