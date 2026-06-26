package httpapi

import (
	"net/http"
	"strconv"
	"strings"
)

func (s *Server) handleSupervisorDecompose(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "supervisor.decompose")
}

func (s *Server) handleSupervisorSupervise(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "supervisor.supervise")
}

func (s *Server) handleSupervisorStatus(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodGet, "supervisor.status", nil)
}

func (s *Server) handleSupervisorListTasks(w http.ResponseWriter, r *http.Request) {
	payload := map[string]any{}
	if limit := strings.TrimSpace(r.URL.Query().Get("limit")); limit != "" {
		parsed, err := strconv.Atoi(limit)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid limit query parameter"})
			return
		}
		payload["limit"] = parsed
	}
	if status := strings.TrimSpace(r.URL.Query().Get("status")); status != "" {
		payload["status"] = status
	}
	if len(payload) == 0 {
		s.handleTRPCBridgeCall(w, r, http.MethodGet, "supervisor.listTasks", nil)
		return
	}
	s.handleTRPCBridgeCall(w, r, http.MethodGet, "supervisor.listTasks", payload)
}

func (s *Server) handleSupervisorCancel(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "supervisor.cancel")
}
