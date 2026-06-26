package httpapi

import (
	"net/http"
	"strings"
)

func (s *Server) handleCouncilFineTuneDatasets(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		payload := map[string]any{}
		if taskType := strings.TrimSpace(r.URL.Query().Get("taskType")); taskType != "" {
			payload["taskType"] = taskType
		}
		if len(payload) == 0 {
			s.handleTRPCBridgeCall(w, r, http.MethodGet, "council.fineTune.listDatasets", nil)
			return
		}
		s.handleTRPCBridgeCall(w, r, http.MethodGet, "council.fineTune.listDatasets", payload)
	case http.MethodPost:
		s.handleTRPCBridgeBodyCall(w, r, "council.fineTune.createDataset")
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
	}
}

func (s *Server) handleCouncilFineTuneDatasetGet(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing id query parameter"})
		return
	}
	s.handleTRPCBridgeCall(w, r, http.MethodGet, "council.fineTune.getDataset", map[string]any{"id": id})
}

func (s *Server) handleCouncilFineTuneJobs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleTRPCBridgeCall(w, r, http.MethodGet, "council.fineTune.listJobs", nil)
	case http.MethodPost:
		s.handleTRPCBridgeBodyCall(w, r, "council.fineTune.createJob")
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
	}
}

func (s *Server) handleCouncilFineTuneJobStart(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "council.fineTune.startJob")
}

func (s *Server) handleCouncilFineTuneModels(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleTRPCBridgeCall(w, r, http.MethodGet, "council.fineTune.listModels", nil)
	case http.MethodPost:
		s.handleTRPCBridgeBodyCall(w, r, "council.fineTune.registerModel")
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
	}
}

func (s *Server) handleCouncilFineTuneModelDeploy(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "council.fineTune.deployModel")
}

func (s *Server) handleCouncilFineTuneChat(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "council.fineTune.chat")
}

func (s *Server) handleCouncilFineTuneStats(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodGet, "council.fineTune.stats", nil)
}
