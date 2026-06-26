package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/tormentnexushq/tormentnexus-go/internal/orchestration"
)

func (s *Server) handleCouncilBaseStatus(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodGet, "council.status", nil)
}

func (s *Server) handleCouncilBaseUpdateConfig(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "council.updateConfig")
}

func (s *Server) handleCouncilBaseAddSupervisors(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "council.addSupervisors")
}

func (s *Server) handleCouncilBaseClearSupervisors(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodPost, "council.clearSupervisors", nil)
}

func (s *Server) handleCouncilBaseDebate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	var payload struct {
		ID          string   `json:"id"`
		Objective   string   `json:"objective"`
		Description string   `json:"description"`
		Context     string   `json:"context"`
		Files       []string `json:"files"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	normalizedObjective := payload.Objective
	if normalizedObjective == "" {
		normalizedObjective = payload.Description
	}

	upstreamPayload := map[string]any{
		"id":          payload.ID,
		"objective":   normalizedObjective,
		"description": normalizedObjective,
		"context":     payload.Context,
		"files":       payload.Files,
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "council.debate", upstreamPayload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "council.debate",
			},
		})
		return
	}

	debateRes, fallbackErr := orchestration.RunDebate(r.Context(), s.debateHistory, normalizedObjective, payload.Context)
	if fallbackErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"success": false,
			"error":   fallbackErr.Error(),
			"detail":  fallbackErr.Error(),
		})
		return
	}

	savedRecord, saveErr := s.debateHistory.SaveNativeDebate(r.Context(), payload.ID, normalizedObjective, payload.Context, debateRes)
	if saveErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"success": false,
			"error":   saveErr.Error(),
			"detail":  saveErr.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"result": debateRes,
			"record": savedRecord,
		},
		"bridge": map[string]any{
			"fallback":  "go-local-council-debate",
			"procedure": "council.debate",
			"reason":    "upstream unavailable; executing native Go multi-agent debate loop with native debate-history persistence",
		},
	})
}

func (s *Server) handleCouncilBaseToggle(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodPost, "council.toggle", nil)
}

func (s *Server) handleCouncilBaseAddMock(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodPost, "council.addMock", nil)
}
