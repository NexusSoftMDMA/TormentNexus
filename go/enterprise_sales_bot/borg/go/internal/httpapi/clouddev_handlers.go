package httpapi

import (
	"net/http"
	"strconv"
	"strings"
)

func (s *Server) handleCloudDevListProviders(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodGet, "cloudDev.listProviders", nil)
}

func (s *Server) handleCloudDevCreateSession(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "cloudDev.createSession")
}

func (s *Server) handleCloudDevListSessions(w http.ResponseWriter, r *http.Request) {
	payload := map[string]any{}
	if provider := strings.TrimSpace(r.URL.Query().Get("provider")); provider != "" {
		payload["provider"] = provider
	}
	if status := strings.TrimSpace(r.URL.Query().Get("status")); status != "" {
		payload["status"] = status
	}
	if len(payload) == 0 {
		s.handleTRPCBridgeCall(w, r, http.MethodGet, "cloudDev.listSessions", nil)
		return
	}
	s.handleTRPCBridgeCall(w, r, http.MethodGet, "cloudDev.listSessions", payload)
}

func (s *Server) handleCloudDevGetSession(w http.ResponseWriter, r *http.Request) {
	sessionID := strings.TrimSpace(r.URL.Query().Get("sessionId"))
	if sessionID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing sessionId query parameter"})
		return
	}
	s.handleTRPCBridgeCall(w, r, http.MethodGet, "cloudDev.getSession", map[string]any{"sessionId": sessionID})
}

func (s *Server) handleCloudDevUpdateSessionStatus(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "cloudDev.updateSessionStatus")
}

func (s *Server) handleCloudDevDeleteSession(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "cloudDev.deleteSession")
}

func (s *Server) handleCloudDevSendMessage(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "cloudDev.sendMessage")
}

func (s *Server) handleCloudDevBroadcastMessage(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "cloudDev.broadcastMessage")
}

func (s *Server) handleCloudDevPreviewBroadcastRecipients(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "cloudDev.previewBroadcastRecipients")
}

func (s *Server) handleCloudDevAcceptPlan(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "cloudDev.acceptPlan")
}

func (s *Server) handleCloudDevSetAutoAcceptPlan(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "cloudDev.setAutoAcceptPlan")
}

func (s *Server) handleCloudDevGetMessages(w http.ResponseWriter, r *http.Request) {
	sessionID := strings.TrimSpace(r.URL.Query().Get("sessionId"))
	if sessionID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing sessionId query parameter"})
		return
	}
	payload := map[string]any{"sessionId": sessionID}
	if limit := strings.TrimSpace(r.URL.Query().Get("limit")); limit != "" {
		if parsed, err := strconv.Atoi(limit); err == nil {
			payload["limit"] = parsed
		}
	}
	s.handleTRPCBridgeCall(w, r, http.MethodGet, "cloudDev.getMessages", payload)
}

func (s *Server) handleCloudDevGetLogs(w http.ResponseWriter, r *http.Request) {
	sessionID := strings.TrimSpace(r.URL.Query().Get("sessionId"))
	if sessionID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing sessionId query parameter"})
		return
	}
	payload := map[string]any{"sessionId": sessionID}
	if limit := strings.TrimSpace(r.URL.Query().Get("limit")); limit != "" {
		if parsed, err := strconv.Atoi(limit); err == nil {
			payload["limit"] = parsed
		}
	}
	s.handleTRPCBridgeCall(w, r, http.MethodGet, "cloudDev.getLogs", payload)
}

func (s *Server) handleCloudDevStats(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodGet, "cloudDev.stats", nil)
}
