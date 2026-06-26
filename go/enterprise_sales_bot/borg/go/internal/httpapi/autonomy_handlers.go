package httpapi

import "net/http"

func (s *Server) handleAutonomyGetLevel(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodGet, "autonomy.getLevel", nil)
}

func (s *Server) handleAutonomySetLevel(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "autonomy.setLevel")
}

func (s *Server) handleAutonomyActivateFull(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodPost, "autonomy.activateFullAutonomy", nil)
}
