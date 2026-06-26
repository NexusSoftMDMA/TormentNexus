package httpapi

import "net/http"

func (s *Server) handleCouncilEvolutionStart(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodPost, "council.evolution.start", nil)
}

func (s *Server) handleCouncilEvolutionStop(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodPost, "council.evolution.stop", nil)
}

func (s *Server) handleCouncilEvolutionOptimize(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodPost, "council.evolution.optimize", nil)
}

func (s *Server) handleCouncilEvolutionEvolve(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "council.evolution.evolve")
}

func (s *Server) handleCouncilEvolutionTest(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodGet, "council.evolution.test", nil)
}
