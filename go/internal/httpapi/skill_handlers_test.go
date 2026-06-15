package httpapi

/**
 * @file skill_handlers_test.go
 * @module go/internal/httpapi
 *
 * WHAT: Unit tests for Skill API handlers.
 * Tests list, get, and search endpoints against orchestration.GlobalSkillRegistry.
 */

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tormentnexushq/tormentnexus-go/internal/orchestration"
)

// TestSkillHandlerList verifies GET /api/skills/list returns registered skills.
func TestSkillHandlerList(t *testing.T) {
	// Seed the global registry
	orchestration.GlobalSkillRegistry.RegisterAgentSkill("http://agent1:4300", "skill-alpha")
	orchestration.GlobalSkillRegistry.RegisterAgentSkill("http://agent2:4300", "skill-beta")
	orchestration.GlobalSkillRegistry.RegisterAgentSkill("http://agent3:4300", "skill-alpha")

	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/api/skills/list", nil)
	w := httptest.NewRecorder()
	s.handleSkillList(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp struct {
		Success bool `json:"success"`
		Skills  []struct {
			ID        string   `json:"id"`
			AgentURLs []string `json:"agent_urls"`
		} `json:"skills"`
		Count int `json:"count"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Fatal("expected success=true")
	}
	if resp.Count == 0 {
		t.Fatal("expected skills to be returned")
	}

	// skill-alpha should be registered
	found := false
	for _, sk := range resp.Skills {
		if sk.ID == "skill-alpha" {
			found = true
			if len(sk.AgentURLs) < 2 {
				t.Fatalf("expected at least 2 agent URLs for skill-alpha, got %d", len(sk.AgentURLs))
			}
			break
		}
	}
	if !found {
		t.Fatal("expected skill-alpha to be in the list")
	}
}

// TestSkillHandlerGetOK verifies GET /api/skills/get?id=... returns a skill.
func TestSkillHandlerGetOK(t *testing.T) {
	orchestration.GlobalSkillRegistry.RegisterAgentSkill("http://agent1:4300", "skill-gamma")

	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/api/skills/get?id=skill-gamma", nil)
	w := httptest.NewRecorder()
	s.handleSkillGet(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Success bool `json:"success"`
		Skill   struct {
			ID        string   `json:"id"`
			AgentURLs []string `json:"agent_urls"`
		} `json:"skill"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Fatal("expected success=true")
	}
	if resp.Skill.ID != "skill-gamma" {
		t.Fatalf("expected skill-gamma, got %s", resp.Skill.ID)
	}
	if len(resp.Skill.AgentURLs) == 0 {
		t.Fatal("expected at least one agent URL")
	}
}

// TestSkillHandlerGetNotFound verifies 404 for missing skill.
func TestSkillHandlerGetNotFound(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/api/skills/get?id=nonexistent-skill", nil)
	w := httptest.NewRecorder()
	s.handleSkillGet(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

// TestSkillHandlerGetMissingParam verifies 400 when id is missing.
func TestSkillHandlerGetMissingParam(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/api/skills/get", nil)
	w := httptest.NewRecorder()
	s.handleSkillGet(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

// TestSkillHandlerSearch verifies skill search by substring.
func TestSkillHandlerSearch(t *testing.T) {
	orchestration.GlobalSkillRegistry.RegisterAgentSkill("http://agent1:4300", "test-search-alpha")
	orchestration.GlobalSkillRegistry.RegisterAgentSkill("http://agent1:4300", "test-search-beta")
	orchestration.GlobalSkillRegistry.RegisterAgentSkill("http://agent2:4300", "irrelevant")

	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/api/skills/search?q=test-search", nil)
	w := httptest.NewRecorder()
	s.handleSkillSearch(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Success bool `json:"success"`
		Skills  []struct {
			ID        string   `json:"id"`
			AgentURLs []string `json:"agent_urls"`
		} `json:"skills"`
		Count int `json:"count"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Fatal("expected success=true")
	}
	if resp.Count < 2 {
		t.Fatalf("expected at least 2 matching skills, got %d", resp.Count)
	}
}

// TestSkillHandlerSearchMissingParam verifies 400 when q is missing.
func TestSkillHandlerSearchMissingParam(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/api/skills/search", nil)
	w := httptest.NewRecorder()
	s.handleSkillSearch(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

// TestSkillHandlerListMethodNotAllowed verifies POST is rejected for list.
func TestSkillHandlerListMethodNotAllowed(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest(http.MethodPost, "/api/skills/list", nil)
	w := httptest.NewRecorder()
	s.handleSkillList(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

// TestSkillHandlerLoadStub verifies the load stub returns 501.
func TestSkillHandlerLoadStub(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/api/skills/load", nil)
	w := httptest.NewRecorder()
	s.handleSkillLoad(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", w.Code)
	}
}

// TestSkillHandlerUnloadStub verifies the unload stub returns 501.
func TestSkillHandlerUnloadStub(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/api/skills/unload", nil)
	w := httptest.NewRecorder()
	s.handleSkillUnload(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", w.Code)
	}
}

// TestSkillHandlerListLoadedStub verifies the list-loaded stub returns 501.
func TestSkillHandlerListLoadedStub(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/api/skills/list-loaded", nil)
	w := httptest.NewRecorder()
	s.handleSkillListLoaded(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", w.Code)
	}
}
