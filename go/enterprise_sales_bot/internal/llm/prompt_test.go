package llm

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPromptRegistry_RegisterAndResolve(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "registry.json")
	pr := NewPromptRegistry(file)

	v1 := pr.RegisterVersion("outreach", "Hello ${Contact}, welcome to ${Company}!")
	if v1 == nil || v1.ID == "" {
		t.Fatal("expected valid version")
	}
	if !v1.Enabled {
		t.Fatal("expected version to be enabled by default")
	}

	out, err := pr.ResolvePrompt("outreach", map[string]string{
		"Contact": "Alice",
		"Company": "Acme Corp",
	})
	if err != nil {
		t.Fatalf("resolve error: %v", err)
	}
	expected := "Hello Alice, welcome to Acme Corp!"
	if out != expected {
		t.Fatalf("expected %q, got %q", expected, out)
	}
}

func TestPromptRegistry_NoVersion(t *testing.T) {
	dir := t.TempDir()
	pr := NewPromptRegistry(filepath.Join(dir, "reg.json"))

	_, err := pr.ResolvePrompt("nonexistent", nil)
	if err == nil {
		t.Fatal("expected error for missing prompt name")
	}
}

func TestPromptRegistry_ABExperiment(t *testing.T) {
	dir := t.TempDir()
	pr := NewPromptRegistry(filepath.Join(dir, "reg.json"))

	v1 := pr.RegisterVersion("outreach", "Version A: ${Contact}")
	v2 := pr.RegisterVersion("outreach", "Version B: ${Contact}")
	if v1 == nil || v2 == nil {
		t.Fatal("register failed")
	}

	err := pr.AssignExperiment("outreach", []string{v1.ID, v2.ID}, []float64{0.5, 0.5})
	if err != nil {
		t.Fatalf("assign experiment: %v", err)
	}

	// Run many resolutions and confirm both versions appear.
	seen := make(map[string]bool)
	for i := 0; i < 30; i++ {
		out, err := pr.ResolvePrompt("outreach", map[string]string{"Contact": "Bob"})
		if err != nil {
			t.Fatalf("resolve error: %v", err)
		}
		seen[out] = true
	}
	if len(seen) != 2 {
		t.Fatalf("expected both A/B versions to appear, got %d: %v", len(seen), seen)
	}
}

func TestPromptRegistry_RecordOutcomes(t *testing.T) {
	dir := t.TempDir()
	pr := NewPromptRegistry(filepath.Join(dir, "reg.json"))

	v1 := pr.RegisterVersion("test", "Template ${V}")
	_ = pr.AssignExperiment("test", []string{v1.ID}, []float64{1.0})

	pr.RecordOutcome("test", v1.ID, true)
	pr.RecordOutcome("test", v1.ID, true)
	pr.RecordOutcome("test", v1.ID, false)

	outcomes := pr.GetOutcomes()
	found := false
	for _, r := range outcomes {
		if r.Experiment == "test" && r.VersionID == v1.ID {
			found = true
			if r.Success != 2 || r.Failure != 1 || r.Total != 3 {
				t.Fatalf("unexpected result: Success=%d Failure=%d Total=%d", r.Success, r.Failure, r.Total)
			}
		}
	}
	if !found {
		t.Fatal("expected outcome record not found")
	}
}

func TestPromptRegistry_Persistence(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "persist.json")

	// Write a registry with data
	pr1 := NewPromptRegistry(file)
	v := pr1.RegisterVersion("outreach", "Hello ${Name}")
	pr1.RecordOutcome("test_exp", v.ID, true)

	// Create new registry from same file (should load)
	pr2 := NewPromptRegistry(file)

	// Verify version loaded
	out, err := pr2.ResolvePrompt("outreach", map[string]string{"Name": "Carol"})
	if err != nil {
		t.Fatalf("resolve after reload: %v", err)
	}
	if out != "Hello Carol" {
		t.Fatalf("expected 'Hello Carol', got %q", out)
	}

	// Verify outcomes loaded
	outcomes := pr2.GetOutcomes()
	found := false
	for _, r := range outcomes {
		if r.Experiment == "test_exp" && r.VersionID == v.ID {
			found = true
		}
	}
	if !found {
		t.Fatal("expected outcomes to persist across reloads")
	}

	// Cleanup
	_ = os.Remove(file)
}
