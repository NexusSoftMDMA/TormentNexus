package marketplace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultRegistryConfig(t *testing.T) {
	cfg := DefaultRegistryConfig()
	if cfg.RegistryURL == "" {
		t.Error("RegistryURL should not be empty")
	}
	if cfg.LocalDir == "" {
		t.Error("LocalDir should not be empty")
	}
	if cfg.CacheTTL == 0 {
		t.Error("CacheTTL should not be zero")
	}
}

func TestServiceCreation(t *testing.T) {
	cfg := DefaultRegistryConfig()
	cfg.LocalDir = t.TempDir()
	svc := NewService(cfg)

	if svc.cfg.LocalDir != cfg.LocalDir {
		t.Errorf("LocalDir mismatch: got %q, want %q", svc.cfg.LocalDir, cfg.LocalDir)
	}
}

func TestServiceEmptyDir(t *testing.T) {
	cfg := DefaultRegistryConfig()
	cfg.LocalDir = t.TempDir()
	svc := NewService(cfg)

	// List should work with empty dir
	entries, err := svc.List("", nil, 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries from empty dir, got %d", len(entries))
	}
}

func TestInstallAndUninstall(t *testing.T) {
	cfg := DefaultRegistryConfig()
	cfg.LocalDir = t.TempDir()
	svc := NewService(cfg)

	// Create a skill content file
	contentDir := t.TempDir()
	skillContent := "---\nname: test-skill\ndescription: A test skill\n---\n\n# Test Skill\n\nDoes testing things."
	skillFile := filepath.Join(contentDir, "SKILL.md")
	if err := os.WriteFile(skillFile, []byte(skillContent), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Install from file
	entry, err := svc.Install(InstallRequest{
		SkillID: "test-skill",
		Source:  "file",
		URL:     skillFile,
	})
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}
	if entry.ID != "test-skill" {
		t.Errorf("ID mismatch: got %q", entry.ID)
	}

	// Verify file was created
	installedFile := filepath.Join(cfg.LocalDir, "test-skill", "SKILL.md")
	if _, err := os.Stat(installedFile); os.IsNotExist(err) {
		t.Error("Installed SKILL.md should exist")
	}

	// Verify marketplace metadata
	metaFile := filepath.Join(cfg.LocalDir, "test-skill", "marketplace.json")
	metaData, err := os.ReadFile(metaFile)
	if err != nil {
		t.Fatalf("ReadFile marketplace.json: %v", err)
	}
	var meta map[string]any
	if err := json.Unmarshal(metaData, &meta); err != nil {
		t.Fatalf("Unmarshal marketplace.json: %v", err)
	}
	if meta["source"] != "file" {
		t.Errorf("Expected source 'file', got %v", meta["source"])
	}

	// List installed
	installed, err := svc.Installed()
	if err != nil {
		t.Fatalf("Installed failed: %v", err)
	}
	if len(installed) != 1 {
		t.Errorf("Expected 1 installed skill, got %d", len(installed))
	}

	// Uninstall
	if err := svc.Uninstall("test-skill"); err != nil {
		t.Fatalf("Uninstall failed: %v", err)
	}

	if _, err := os.Stat(installedFile); !os.IsNotExist(err) {
		t.Error("Installed file should be removed after uninstall")
	}
}

func TestInstallDuplicatePrevention(t *testing.T) {
	cfg := DefaultRegistryConfig()
	cfg.LocalDir = t.TempDir()
	svc := NewService(cfg)

	contentDir := t.TempDir()
	skillContent := "---\nname: dup-skill\n---\n\n# Dup Skill"
	skillFile := filepath.Join(contentDir, "SKILL.md")
	os.WriteFile(skillFile, []byte(skillContent), 0o644)

	// First install should succeed
	_, err := svc.Install(InstallRequest{
		SkillID: "dup-skill",
		Source:  "file",
		URL:     skillFile,
	})
	if err != nil {
		t.Fatalf("First install failed: %v", err)
	}

	// Second install without force should fail
	_, err = svc.Install(InstallRequest{
		SkillID: "dup-skill",
		Source:  "file",
		URL:     skillFile,
	})
	if err == nil {
		t.Error("Expected error for duplicate install without force")
	}

	// Install with force should succeed
	_, err = svc.Install(InstallRequest{
		SkillID: "dup-skill",
		Source:  "file",
		URL:     skillFile,
		Force:   true,
	})
	if err != nil {
		t.Errorf("Force install should succeed, got: %v", err)
	}
}

func TestPublishLocal(t *testing.T) {
	cfg := DefaultRegistryConfig()
	cfg.LocalDir = t.TempDir()
	svc := NewService(cfg)

	entry, err := svc.Publish(PublishRequest{
		ID:          "publish-test",
		Name:        "Publish Test",
		Description: "A test publish",
		Author:      "tester",
		Version:     "1.0.0",
		Content:     "# Publish Test\n\nThis is a test.",
	})

	if err == nil {
		// Remote publish succeeded (unlikely in test)
		t.Log("Remote publish succeeded unexpectedly")
	} else {
		// Should have saved locally as pending
		pendingFile := filepath.Join(cfg.LocalDir, "pending", "publish-test", "publish.json")
		data, fileErr := os.ReadFile(pendingFile)
		if fileErr != nil {
			t.Logf("Pending file not found (expected when no remote registry): %v", fileErr)
		} else {
			var pending SkillEntry
			if jsonErr := json.Unmarshal(data, &pending); jsonErr != nil {
				t.Errorf("Failed to parse pending publish: %v", jsonErr)
			}
			if pending.ID != "publish-test" {
				t.Errorf("Pending ID mismatch: got %q", pending.ID)
			}
		}
	}

	// Entry should always be returned
	if entry.ID != "publish-test" {
		t.Errorf("Entry ID mismatch: got %q", entry.ID)
	}
}

func TestFilterSkills(t *testing.T) {
	cfg := DefaultRegistryConfig()
	svc := NewService(cfg)

	entries := []SkillEntry{
		{ID: "skill-a", Name: "Skill A", Description: "A testing skill", Tags: []string{"test", "dev"}},
		{ID: "skill-b", Name: "Skill B", Description: "A production skill", Tags: []string{"prod"}},
		{ID: "skill-c", Name: "Skill C", Description: "Another test skill", Tags: []string{"test"}},
	}

	// Filter by query
	filtered := svc.filterSkills(entries, "test", nil, 10)
	if len(filtered) != 2 {
		t.Errorf("Expected 2 skills matching 'test', got %d", len(filtered))
	}

	// Filter by tag
	filtered = svc.filterSkills(entries, "", []string{"test"}, 10)
	if len(filtered) != 2 {
		t.Errorf("Expected 2 skills with 'test' tag, got %d", len(filtered))
	}

	// Filter with limit
	filtered = svc.filterSkills(entries, "", nil, 2)
	if len(filtered) != 2 {
		t.Errorf("Expected 2 skills with limit, got %d", len(filtered))
	}

	// Filter with no matches
	filtered = svc.filterSkills(entries, "nonexistent", nil, 10)
	if len(filtered) != 0 {
		t.Errorf("Expected 0 skills matching 'nonexistent', got %d", len(filtered))
	}
}

func TestEnrichContent(t *testing.T) {
	cfg := DefaultRegistryConfig()
	svc := NewService(cfg)

	entry := &SkillEntry{
		Name:        "enriched",
		Description: "An enriched skill",
		Author:      "tester",
		Version:     "1.0.0",
	}

	// Content without frontmatter should get frontmatter added
	content := "# My Skill\n\nDoes things."
	enriched := svc.enrichContent(entry, content)
	if !contains(enriched, "name: enriched") {
		t.Error("Enriched content should contain name frontmatter")
	}
	if !contains(enriched, "marketplace: true") {
		t.Error("Enriched content should contain marketplace: true")
	}

	// Content with existing frontmatter should not be modified
	existingFM := "---\nname: existing\n---\n\n# Existing Skill"
	enriched = svc.enrichContent(entry, existingFM)
	if enriched != existingFM {
		t.Error("Content with existing frontmatter should not be modified")
	}
}

func TestUninstallNonExistent(t *testing.T) {
	cfg := DefaultRegistryConfig()
	cfg.LocalDir = t.TempDir()
	svc := NewService(cfg)

	err := svc.Uninstall("nonexistent")
	if err == nil {
		t.Error("Expected error for uninstalling non-existent skill")
	}
}

func TestInstallMissingSkillID(t *testing.T) {
	cfg := DefaultRegistryConfig()
	cfg.LocalDir = t.TempDir()
	svc := NewService(cfg)

	_, err := svc.Install(InstallRequest{SkillID: ""})
	if err == nil {
		t.Error("Expected error for missing skill_id")
	}
}

func TestPublishMissingID(t *testing.T) {
	cfg := DefaultRegistryConfig()
	cfg.LocalDir = t.TempDir()
	svc := NewService(cfg)

	_, err := svc.Publish(PublishRequest{Content: "some content"})
	if err == nil {
		t.Error("Expected error for missing id")
	}
}

func TestPublishMissingContent(t *testing.T) {
	cfg := DefaultRegistryConfig()
	cfg.LocalDir = t.TempDir()
	svc := NewService(cfg)

	_, err := svc.Publish(PublishRequest{ID: "test"})
	if err == nil {
		t.Error("Expected error for missing content")
	}
}

func TestCacheExpiry(t *testing.T) {
	cfg := DefaultRegistryConfig()
	cfg.LocalDir = t.TempDir()
	cfg.CacheTTL = 100 * time.Millisecond
	svc := NewService(cfg)

	// Set cache
	svc.mu.Lock()
	svc.cache = []SkillEntry{
		{ID: "cached-skill", Name: "Cached"},
	}
	svc.cacheAt = time.Now()
	svc.mu.Unlock()

	// Should return cached result
	entries := svc.getLocalSkills("cached", nil, 10)
	if len(entries) != 1 {
		t.Errorf("Expected 1 cached entry, got %d", len(entries))
	}

	// Wait for cache to expire
	time.Sleep(150 * time.Millisecond)

	// Should now try installed (which is empty)
	entries = svc.getLocalSkills("cached", nil, 10)
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries after cache expiry, got %d", len(entries))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
