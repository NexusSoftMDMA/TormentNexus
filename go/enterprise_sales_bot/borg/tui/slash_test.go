package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NexusSoftMDMA/TormentNexus/agents"
)

func TestProcessSlashCommandPlanAndRepomap(t *testing.T) {
	cwd := t.TempDir()
	if err := os.WriteFile(filepath.Join(cwd, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m := model{director: agents.NewDirector(&agents.DefaultProvider{})}
	m.director.WorkingDir = cwd
	mdl, _ := ProcessSlashCommand("/plan analyze this repo", &m)
	updated := mdl.(model)
	if len(updated.history) == 0 || !strings.Contains(updated.history[len(updated.history)-1], "[Foundation Plan]") {
		t.Fatalf("expected plan output, got %#v", updated.history)
	}
	mdl, _ = ProcessSlashCommand("/repomap", &updated)
	updated = mdl.(model)
	if len(updated.history) == 0 || !strings.Contains(updated.history[len(updated.history)-1], "<repo_map>") {
		t.Fatalf("expected repomap output, got %#v", updated.history)
	}
}

func TestProcessSlashCommandProvidersAndAdapters(t *testing.T) {
	cwd := t.TempDir()
	m := model{director: agents.NewDirector(&agents.DefaultProvider{})}
	m.director.WorkingDir = cwd
	mdl, _ := ProcessSlashCommand("/providers", &m)
	updated := mdl.(model)
	if len(updated.history) == 0 || !strings.Contains(updated.history[len(updated.history)-1], "[Foundation Providers]") {
		t.Fatalf("expected provider output, got %#v", updated.history)
	}
	mdl, _ = ProcessSlashCommand("/adapters", &updated)
	updated = mdl.(model)
	if len(updated.history) == 0 || !strings.Contains(updated.history[len(updated.history)-1], "[Foundation Adapters]") {
		t.Fatalf("expected adapter output, got %#v", updated.history)
	}
}

func TestProcessSlashCommandClearResetsDirector(t *testing.T) {
	m := model{director: agents.NewDirector(&agents.DefaultProvider{})}
	m.history = []string{"old"}
	mdl, _ := ProcessSlashCommand("/clear", &m)
	updated := mdl.(model)
	if len(updated.history) != 1 {
		t.Fatalf("expected reset history, got %#v", updated.history)
	}
	if updated.director == nil {
		t.Fatal("expected director reset")
	}
}
