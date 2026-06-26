package httpapi

import "net/http"

func (s *Server) handleCouncilSmartPilotStatus(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodGet, "council.smartPilot.status", nil)
}

func (s *Server) handleCouncilSmartPilotConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleTRPCBridgeCall(w, r, http.MethodGet, "council.smartPilot.getConfig", nil)
	case http.MethodPost:
		s.handleTRPCBridgeBodyCall(w, r, "council.smartPilot.updateConfig")
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
	}
}

func (s *Server) handleCouncilSmartPilotTrigger(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "council.smartPilot.trigger")
}

func (s *Server) handleCouncilSmartPilotResetCount(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "council.smartPilot.resetCount")
}

func (s *Server) handleCouncilSmartPilotResetAll(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodPost, "council.smartPilot.resetAllCounts", nil)
}
