package mcpimpl

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestSkillRegistry(t *testing.T) {
	// Setup temporary database path
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "skills_test.db")

	// Set test environment variable or override global configuration path
	// Temporarily override the registry database path
	originalRegistry := skillRegistryInstance
	defer func() {
		skillRegistryInstance = originalRegistry
	}()

	var err error
	skillRegistryInstance, err = NewSkillRegistry(dbPath)
	if err != nil {
		t.Fatalf("failed to create skill registry: %v", err)
	}
	defer skillRegistryInstance.db.Close()

	ctx := context.Background()

	// 1. Store a new skill
	storeResp, err := HandleSkillStore(ctx, map[string]interface{}{
		"name":        "Go File Reading",
		"description": "Read files in Go",
		"category":    "go",
		"frontmatter": "Read files progressively",
		"content":     "package main\nimport (\n\t\"fmt\"\n\t\"io/ioutil\"\n)\nfunc main() {\n\tdata, _ := ioutil.ReadFile(\"test.txt\")\n\tfmt.Println(string(data))\n}",
	})
	if err != nil {
		t.Fatalf("unexpected error storing skill: %v", err)
	}
	if storeResp.IsError {
		t.Fatalf("expected success storing skill: %v", storeResp.Content[0].Text)
	}

	// 2. List skills (Verify progressive loading / only frontmatter is returned)
	listResp, err := HandleSkillList(ctx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error listing skills: %v", err)
	}
	if listResp.IsError {
		t.Fatalf("expected success listing skills: %v", listResp.Content[0].Text)
	}

	var listedSkills []map[string]interface{}
	if err := json.Unmarshal([]byte(listResp.Content[0].Text), &listedSkills); err != nil {
		t.Fatalf("failed to parse list response: %v", err)
	}

	if len(listedSkills) != 1 {
		t.Errorf("expected 1 listed skill, got %d", len(listedSkills))
	}
	if listedSkills[0]["name"] != "Go File Reading" {
		t.Errorf("expected name 'Go File Reading', got '%v'", listedSkills[0]["name"])
	}
	if _, hasContent := listedSkills[0]["content"]; hasContent {
		t.Errorf("expected list response to NOT contain full content (progressive loading check)")
	}

	// 3. Get skill by name (Verify full load)
	getResp, err := HandleSkillGet(ctx, map[string]interface{}{
		"name": "Go File Reading",
	})
	if err != nil {
		t.Fatalf("unexpected error getting skill: %v", err)
	}
	if getResp.IsError {
		t.Fatalf("expected success getting skill: %v", getResp.Content[0].Text)
	}

	var fetchedSkill Skill
	if err := json.Unmarshal([]byte(getResp.Content[0].Text), &fetchedSkill); err != nil {
		t.Fatalf("failed to parse get response: %v", err)
	}

	if fetchedSkill.Name != "Go File Reading" {
		t.Errorf("expected fetched skill name 'Go File Reading', got '%s'", fetchedSkill.Name)
	}
	if !strings.Contains(fetchedSkill.Content, "ioutil.ReadFile") {
		t.Errorf("expected fetched skill content to contain full code, got '%s'", fetchedSkill.Content)
	}

	// 4. Store a very similar skill (Verify deduplication / merging)
	// Similarity is Jaccard of words: should be >= 90%
	similarStoreResp, err := HandleSkillStore(ctx, map[string]interface{}{
		"name":        "Go File Reading V2",
		"description": "Read files in Go version 2",
		"category":    "go",
		"frontmatter": "Read files progressively version 2",
		"content":     "package main\nimport (\n\t\"fmt\"\n\t\"io/ioutil\"\n)\nfunc main() {\n\tdata, _ := ioutil.ReadFile(\"test.txt\")\n\tfmt.Println(string(data))\n}",
	})
	if err != nil {
		t.Fatalf("unexpected error storing similar skill: %v", err)
	}
	if similarStoreResp.IsError {
		t.Fatalf("expected success storing similar skill: %v", similarStoreResp.Content[0].Text)
	}

	if !strings.Contains(similarStoreResp.Content[0].Text, "merged") {
		t.Errorf("expected skill to be merged/deduplicated, but response got: %s", similarStoreResp.Content[0].Text)
	}

	// Verify that the database only has 1 skill (merged version), not 2
	getMergedResp, err := HandleSkillGet(ctx, map[string]interface{}{
		"name": "Go File Reading",
	})
	if err != nil {
		t.Fatalf("unexpected error getting merged skill: %v", err)
	}
	var mergedSkill Skill
	if err := json.Unmarshal([]byte(getMergedResp.Content[0].Text), &mergedSkill); err != nil {
		t.Fatalf("failed to parse merged response: %v", err)
	}

	if mergedSkill.Version != 2 {
		t.Errorf("expected merged skill to be version 2, got %d", mergedSkill.Version)
	}
}
