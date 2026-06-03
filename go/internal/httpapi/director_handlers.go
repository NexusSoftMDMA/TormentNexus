package httpapi

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleDirectorMemorize(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "director.memorize")
}

func (s *Server) handleDirectorChat(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "director.chat")
}

func (s *Server) handleDirectorStatus(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodGet, "director.status", nil)
}

func (s *Server) handleDirectorUpdateConfig(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "director.updateConfig")
}

func (s *Server) handleDirectorConfigGet(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "directorConfig.get", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "directorConfig.get",
			},
		})
		return
	}

	result = localSettingsConfig(s.cfg.WorkspaceRoot)
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    result,
		"bridge": map[string]any{
			"fallback":  "go-local-tormentnexus-config",
			"procedure": "directorConfig.get",
			"reason":    "upstream unavailable; using local .tormentnexus/config.json",
		},
	})
}

func (s *Server) handleDirectorConfigTest(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodGet, "directorConfig.test", nil)
}

func (s *Server) handleDirectorConfigUpdate(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "directorConfig.update")
}

func (s *Server) handleDirectorStopAutoDrive(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodPost, "director.stopAutoDrive", nil)
}

func (s *Server) handleDirectorStartAutoDrive(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodPost, "director.startAutoDrive", nil)
}

func (s *Server) handleDirectorNotesList(w http.ResponseWriter, r *http.Request) {
	// Try upstream first
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "director.getNotes", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "director.getNotes",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    s.directorNotes.GetNotes(),
		"bridge": map[string]any{
			"fallback": "go-local-director-notes",
		},
	})
}

func (s *Server) handleDirectorNotesSynthesize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	var payload struct {
		Objective  string `json:"objective"`
		Transcript string `json:"transcript"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	// Try upstream first
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "director.synthesizeNote", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "director.synthesizeNote",
			},
		})
		return
	}

	// Fallback to local
	note, fallbackErr := s.directorNotes.SynthesizeSessionNote(r.Context(), payload.Objective, payload.Transcript)
	if fallbackErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": fallbackErr.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    note,
		"bridge": map[string]any{
			"fallback": "go-local-director-notes",
		},
	})
}
