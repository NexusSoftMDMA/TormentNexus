package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/tormentnexushq/tormentnexus-go/internal/gossip"
)

func (s *Server) handleGossipMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	var msg gossip.GossipMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "failed to decode gossip message: " + err.Error()})
		return
	}

	if s.gossipTransport != nil {
		s.gossipTransport.Receive(msg)
	}

	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}
