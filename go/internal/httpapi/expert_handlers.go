package httpapi

import (
	"encoding/json"
	"net/http"
	"github.com/tormentnexushq/tormentnexus-go/internal/ctxharvester"
)

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
