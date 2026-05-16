package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

func (s *Server) handleSkillList(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "skills.list", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "skills.list",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    s.skillRegistry.List(),
		"bridge": map[string]any{
			"fallback":  "go-local-skills",
			"procedure": "skills.list",
			"reason":    "upstream unavailable",
		},
	})
}

func (s *Server) handleSkillGet(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "id required"})
		return
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "skills.get", map[string]string{"id": id}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "skills.get",
			},
		})
		return
	}

	skill, ok := s.skillRegistry.Get(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]any{"success": false, "error": "skill not found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    skill,
		"bridge": map[string]any{
			"fallback": "go-local-skills",
			"reason":   "upstream unavailable",
		},
	})
}

func (s *Server) handleSkillSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "skills.search", map[string]string{"query": query}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "skills.search",
			},
		})
		return
	}

	// Use native progressive disclosure search
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	results, err := s.skillDecision.SearchSkills(ctx, query)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    results,
		"bridge": map[string]any{
			"fallback":  "go-local-skills",
			"procedure": "skills.search",
			"reason":    "upstream unavailable; using Go decision system",
		},
	})
}

func (s *Server) handleSkillLoad(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	err := s.skillDecision.LoadSkill(ctx, req.ID, false)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"success": false, "error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"success": true, "status": "loaded"})
}

func (s *Server) handleSkillUnload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON"})
		return
	}

	existed := s.skillDecision.UnloadSkill(req.ID)
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "existed": existed})
}

func (s *Server) handleSkillListLoaded(w http.ResponseWriter, r *http.Request) {
	loaded := s.skillDecision.ListLoadedSkills()
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    loaded,
		"count":   len(loaded),
	})
}

func (s *Server) handleSkillSummary(w http.ResponseWriter, r *http.Request) {
	skills := s.skillRegistry.List()
	summary := make([]SkillSummary, len(skills))
	for i, sk := range skills {
		summary[i] = SkillSummary{
			ID:     sk.ID,
			Name:   sk.Name,
			Folder: sk.Category,
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    summary,
	})
}
