package httpapi

import "net/http"

func (s *Server) handleCouncilVisualSystemDiagram(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodGet, "council.visual.systemDiagram", nil)
}

func (s *Server) handleCouncilVisualPlanDiagram(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "council.visual.planDiagram")
}

func (s *Server) handleCouncilVisualParsePlan(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "council.visual.parsePlan")
}
