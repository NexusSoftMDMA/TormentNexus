package httpapi

import (
	"net/http"
	"strings"
)

func (s *Server) handleCouncilRotationList(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodGet, "council.rotation.list", nil)
}

func (s *Server) handleCouncilRotationGet(w http.ResponseWriter, r *http.Request) {
	roomID := strings.TrimSpace(r.URL.Query().Get("roomId"))
	if roomID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing roomId query parameter"})
		return
	}
	s.handleTRPCBridgeCall(w, r, http.MethodGet, "council.rotation.get", map[string]any{"roomId": roomID})
}

func (s *Server) handleCouncilRotationCreate(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "council.rotation.create")
}

func (s *Server) handleCouncilRotationAddParticipant(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "council.rotation.addParticipant")
}

func (s *Server) handleCouncilRotationPostMessage(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "council.rotation.postMessage")
}

func (s *Server) handleCouncilRotationSetAgreement(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "council.rotation.setAgreement")
}

func (s *Server) handleCouncilRotationAdvanceTurn(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "council.rotation.advanceTurn")
}

func (s *Server) handleCouncilRotationConfigureSupervisor(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "council.rotation.configureSupervisor")
}

func (s *Server) handleCouncilRotationRunSupervisorCheck(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "council.rotation.runSupervisorCheck")
}

func (s *Server) handleCouncilRotationUpdateSharedContext(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "council.rotation.updateSharedContext")
}

func (s *Server) handleCouncilRotationPause(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "council.rotation.pause")
}

func (s *Server) handleCouncilRotationResume(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "council.rotation.resume")
}

func (s *Server) handleCouncilRotationStartExecution(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "council.rotation.startExecution")
}

func (s *Server) handleCouncilRotationComplete(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "council.rotation.complete")
}
