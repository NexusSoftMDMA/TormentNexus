package httpapi

/**
 * @file skill_handlers.go
 * @module go/internal/httpapi
 *
 * WHAT: HTTP handlers for the Skill API.
 * Provides list, get, and search endpoints over orchestration.GlobalSkillRegistry.
 *
 * WHY: External clients (tRPC, dashboard) need to query the assembled skill catalog
 * to discover agent capabilities.
 */

import (
	"net/http"
	"strings"

	"github.com/tormentnexushq/tormentnexus-go/internal/orchestration"
)

// skillEntry is the JSON shape returned for a single skill.
type skillEntry struct {
	ID        string   `json:"id"`
	AgentURLs []string `json:"agent_urls"`
}

// handleSkillList returns all skills registered in the global A2A skill registry.
// GET /api/skills/list
func (s *Server) handleSkillList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	all := orchestration.GlobalSkillRegistry.ListAllSkillAgents()
	entries := make([]skillEntry, 0, len(all))
	for id, urls := range all {
		entries = append(entries, skillEntry{
			ID:        id,
			AgentURLs: urls,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"skills":  entries,
		"count":   len(entries),
	})
}

// handleSkillGet returns details for a single skill by ID.
// GET /api/skills/get?id=<skillID>
func (s *Server) handleSkillGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing 'id' query parameter"})
		return
	}

	urls := orchestration.GlobalSkillRegistry.ListSkillAgents(id)
	if len(urls) == 0 {
		writeJSON(w, http.StatusNotFound, map[string]any{"success": false, "error": "skill not found: " + id})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"skill": skillEntry{
			ID:        id,
			AgentURLs: urls,
		},
	})
}

// handleSkillSearch searches skill IDs for the given query term.
// GET /api/skills/search?q=<query>
// handleSkillLoad loads a skill into the active working set.
// Stub — not yet implemented.
func (s *Server) handleSkillLoad(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, map[string]any{"success": false, "error": "not implemented"})
}

// handleSkillUnload removes a skill from the active working set.
// Stub — not yet implemented.
func (s *Server) handleSkillUnload(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, map[string]any{"success": false, "error": "not implemented"})
}

// handleSkillListLoaded returns currently loaded skills.
// Stub — not yet implemented.
func (s *Server) handleSkillListLoaded(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, map[string]any{"success": false, "error": "not implemented"})
}

func (s *Server) handleSkillSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	q := r.URL.Query().Get("q")
	if q == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing 'q' query parameter"})
		return
	}

	all := orchestration.GlobalSkillRegistry.ListAllSkillAgents()
	entries := make([]skillEntry, 0)
	query := strings.ToLower(q)

	for id, urls := range all {
		if strings.Contains(strings.ToLower(id), query) {
			entries = append(entries, skillEntry{
				ID:        id,
				AgentURLs: urls,
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"skills":  entries,
		"count":   len(entries),
	})
}
