package httpapi

import "net/http"

func (s *Server) handleCouncilIDEStatus(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodGet, "council.ide.status", nil)
}

func (s *Server) handleCouncilIDESubmitTask(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "council.ide.submitTask")
}
