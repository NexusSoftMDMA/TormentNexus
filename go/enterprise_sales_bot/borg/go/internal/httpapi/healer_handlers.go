package httpapi

import "net/http"

func (s *Server) handleHealerDiagnose(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "healer.diagnose")
}

func (s *Server) handleHealerHeal(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "healer.heal")
}

func (s *Server) handleHealerHistory(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "healer.getHistory", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "healer.getHistory",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    []map[string]any{},
		"bridge": map[string]any{
			"fallback":  "go-local-healer",
			"procedure": "healer.getHistory",
			"reason":    "upstream unavailable; healer history is empty without an active TypeScript healer runtime",
		},
	})
}
