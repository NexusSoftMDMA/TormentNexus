package httpapi

import "net/http"

func (s *Server) handleCouncilHooksList(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodGet, "council.hooks.list", nil)
}

func (s *Server) handleCouncilHooksRegister(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "council.hooks.register")
}

func (s *Server) handleCouncilHooksUnregister(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "council.hooks.unregister")
}

func (s *Server) handleCouncilHooksClear(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodPost, "council.hooks.clear", nil)
}
