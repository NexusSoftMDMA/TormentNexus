package httpapi

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleSkillsSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	results, err := s.skillDecision.SearchSkills(r.Context(), query)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    results,
	})
}

func (s *Server) handleSkillsListLoaded(w http.ResponseWriter, r *http.Request) {
	results := s.skillDecision.ListLoadedSkills()
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    results,
	})
}

func (s *Server) handleSkillsLoad(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	var payload struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON"})
		return
	}

	err := s.skillDecision.LoadSkill(payload.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

func (s *Server) handleSkillsUnload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	var payload struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON"})
		return
	}

	existed := s.skillDecision.UnloadSkill(payload.ID)
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "existed": existed})
}
