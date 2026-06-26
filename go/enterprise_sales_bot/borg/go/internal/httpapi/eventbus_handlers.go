package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/tormentnexushq/tormentnexus-go/internal/eventbus"
)

func (s *Server) handleEventBusPublish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	var payload struct {
		Type    eventbus.SystemEventType `json:"type"`
		Source  string                   `json:"source"`
		Payload interface{}              `json:"payload"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	s.eventBus.EmitEvent(payload.Type, payload.Source, payload.Payload)
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

func (s *Server) handleEventBusHistory(w http.ResponseWriter, r *http.Request) {
	limit := intParam(r, "limit", 100)
	history := s.eventBus.GetHistory(limit)
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": history})
}
