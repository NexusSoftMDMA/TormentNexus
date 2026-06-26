package httpapi

import (
	"encoding/json"
	"net/http"
	"time"
)

func (s *Server) handleCloudOrchestratorPing() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	}
}

func (s *Server) handleCloudOrchestratorManifest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "tormentnexus-go-orchestrator",
			"name":    "tormentnexus Cloud Orchestrator (Go)",
			"version": "1.0.0",
			"capabilities": []string{
				"cloud_session_management",
				"autonomous_plan_approval",
				"semantic_rag_indexing",
				"council_supervisor_debate",
				"automatic_self_healing",
				"github_issue_conversion",
			},
			"endpoints": map[string]string{
				"sessions": "/api/sessions",
				"summary":  "/api/fleet/summary",
				"rag":      "/api/rag/query",
				"reindex":  "/api/rag/reindex",
			},
			"tormentnexusCompatible": true,
		})
	}
}

func (s *Server) handleCloudOrchestratorSessions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// First try upstream TS server
		var upstreamData any
		_, err := s.callUpstreamJSON(r.Context(), "cloudDev.list", map[string]any{}, &upstreamData)
		if err == nil {
			writeJSON(w, http.StatusOK, upstreamData)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		// Fallback local logic
		json.NewEncoder(w).Encode(map[string]interface{}{
			"sessions": []interface{}{},
			"bridge": map[string]interface{}{
				"status": "fallback",
				"reason": err.Error(),
			},
		})
	}
}

func (s *Server) handleCloudOrchestratorFleetSummary() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// First try upstream TS server
		var upstreamData any
		_, err := s.callUpstreamJSON(r.Context(), "cloudDev.stats", map[string]any{}, &upstreamData)
		if err == nil {
			writeJSON(w, http.StatusOK, upstreamData)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		// Fallback local logic
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "healthy",
			"nodes": 1,
			"bridge": map[string]interface{}{
				"status": "fallback",
				"reason": err.Error(),
			},
		})
	}
}

func (s *Server) RegisterCloudOrchestratorRoutes() {
	s.mux.HandleFunc("GET /api/ping", s.handleCloudOrchestratorPing())
	s.mux.HandleFunc("GET /api/manifest", s.handleCloudOrchestratorManifest())
	s.mux.HandleFunc("GET /api/sessions", s.handleCloudOrchestratorSessions())
	s.mux.HandleFunc("GET /api/fleet/summary", s.handleCloudOrchestratorFleetSummary())
	// Additional routes to be ported over from TS
}
