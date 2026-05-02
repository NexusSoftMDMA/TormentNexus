package httpapi

import (
	"encoding/json"
	"net/http"
	"github.com/borghq/borg-go/internal/ctxharvester"
)

func (s *Server) handleExpertResearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	var payload struct {
		Topic string `json:"topic"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	// Try upstream first
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "expert.research", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "expert.research",
			},
		})
		return
	}

	// Fallback to local
	res, fallbackErr := s.expertManager.ExpertResearch(r.Context(), payload.Topic)
	if fallbackErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": fallbackErr.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    res,
		"bridge": map[string]any{
			"fallback": "go-local-expert",
		},
	})
}

func (s *Server) handleExpertCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	var payload struct {
		Instruction string `json:"instruction"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	// Try upstream first
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "expert.code", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "expert.code",
			},
		})
		return
	}

	// Fallback to local
	res, fallbackErr := s.expertManager.ExpertCode(r.Context(), payload.Instruction)
	if fallbackErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": fallbackErr.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    res,
		"bridge": map[string]any{
			"fallback": "go-local-expert",
		},
	})
}

func (s *Server) handleExpertStatus(w http.ResponseWriter, r *http.Request) {
	// Try upstream first
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "expert.getStatus", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "expert.getStatus",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"researcher": "online",
			"coder":      "online",
		},
		"bridge": map[string]any{
			"fallback": "go-local-expert",
		},
	})
}

func (s *Server) handleExpertPredict(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		History string `json:"history"`
		Goal    string `json:"goal"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	predicted, err := s.expertManager.PredictTools(r.Context(), payload.History, payload.Goal)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    predicted,
	})
}

func (s *Server) handleExpertGroom(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Messages  []ctxharvester.ChatMessage `json:"messages"`
		MaxTokens int                        `json:"maxTokens"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	groomer := ctxharvester.NewContextGroomer(payload.MaxTokens)
	groomed := groomer.CompressContext(payload.Messages)

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    groomed,
	})
}
