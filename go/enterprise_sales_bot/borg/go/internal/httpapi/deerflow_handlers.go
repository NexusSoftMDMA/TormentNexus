package httpapi

import "net/http"

func (s *Server) handleDeerFlowStatus(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodGet, "deerFlow.status", nil)
}

func (s *Server) handleDeerFlowModels(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodGet, "deerFlow.models", nil)
}

func (s *Server) handleDeerFlowSkills(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodGet, "deerFlow.skills", nil)
}

func (s *Server) handleDeerFlowMemory(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodGet, "deerFlow.memory", nil)
}
