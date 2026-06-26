package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func (s *Server) handleNativeSupervisorList(w http.ResponseWriter, r *http.Request) {
	sessions := s.supervisorManager.ListSessions()
	writeJSON(w, http.StatusOK, map[string]any{
		"success":  true,
		"sessions": sessions,
		"count":    len(sessions),
	})
}

func (s *Server) handleNativeSupervisorCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "POST required"})
		return
	}

	var body struct {
		ID          string            `json:"id"`
		Command     string            `json:"command"`
		Args        []string          `json:"args"`
		Env         map[string]string `json:"env"`
		Cwd         string            `json:"cwd"`
		MaxRestarts int               `json:"maxRestarts"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid body"})
		return
	}

	if body.ID == "" || body.Command == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "id and command required"})
		return
	}

	if body.MaxRestarts <= 0 {
		body.MaxRestarts = 3
	}
	if body.Cwd == "" {
		body.Cwd = s.cfg.WorkspaceRoot
	}

	session, err := s.supervisorManager.CreateSession(body.ID, body.Command, body.Args, body.Env, body.Cwd, body.MaxRestarts)
	if err != nil {
		writeJSON(w, http.StatusConflict, map[string]any{"success": false, "error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "session": session})
}

func (s *Server) handleNativeSupervisorStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "POST required"})
		return
	}

	var body struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing session id"})
		return
	}

	// Detach from the request context so supervised processes are not tied to
	// the lifetime of a single HTTP request.
	if err := s.supervisorManager.StartSession(context.WithoutCancel(r.Context()), body.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "session started", "id": body.ID})
}

func (s *Server) handleNativeSupervisorStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "POST required"})
		return
	}

	var body struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing session id"})
		return
	}

	if err := s.supervisorManager.StopSession(body.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "session stopped", "id": body.ID})
}

func (s *Server) handleNativeSupervisorStatus(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing id"})
		return
	}

	sessions := s.supervisorManager.ListSessions()
	for _, session := range sessions {
		if session.ID == id {
			writeJSON(w, http.StatusOK, map[string]any{"success": true, "session": session})
			return
		}
	}

	writeJSON(w, http.StatusNotFound, map[string]any{"success": false, "error": "session not found"})
}
