package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func (s *Server) handleDirectorMemorize(w http.ResponseWriter, r *http.Request) {
	// Complex LLM operation — try upstream, fall back to informative error
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "director.memorize", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": result, "bridge": map[string]any{"upstreamBase": upstreamBase, "procedure": "director.memorize"}})
		return
	}

	writeJSON(w, http.StatusServiceUnavailable, map[string]any{
		"success": false,
		"error":   "director.memorize requires TS runtime (LLM-dependent)",
		"bridge":  map[string]any{"fallback": "go-local", "procedure": "director.memorize", "reason": "upstream unavailable; LLM-dependent operation not supported in Go fallback"},
	})
}

func (s *Server) handleDirectorChat(w http.ResponseWriter, r *http.Request) {
	// Complex LLM operation — try upstream, fall back to informative error
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "director.chat", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": result, "bridge": map[string]any{"upstreamBase": upstreamBase, "procedure": "director.chat"}})
		return
	}

	writeJSON(w, http.StatusServiceUnavailable, map[string]any{
		"success": false,
		"error":   "director.chat requires TS runtime (LLM-dependent)",
		"bridge":  map[string]any{"fallback": "go-local", "procedure": "director.chat", "reason": "upstream unavailable; LLM-dependent operation not supported in Go fallback"},
	})
}

func (s *Server) handleDirectorStatus(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "director.status", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": result, "bridge": map[string]any{"upstreamBase": upstreamBase, "procedure": "director.status"}})
		return
	}

	// Native Go fallback: return director state from local instance
	status := map[string]any{
		"directorAvailable": s.goDirector != nil,
		"autoDrive":         s.lifecycleModes["autoDrive"],
		"autoDriveActive":   false,
		"status":            "available",
		"timestamp":         time.Now().UnixMilli(),
	}
	if v, ok := s.lifecycleModes["autoDriveActive"].(bool); ok {
		status["autoDriveActive"] = v
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    status,
		"bridge":  map[string]any{"fallback": "go-local-director", "procedure": "director.status", "reason": "upstream unavailable; using local director state"},
	})
}

func (s *Server) handleDirectorUpdateConfig(w http.ResponseWriter, r *http.Request) {
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "director.updateConfig", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": result, "bridge": map[string]any{"upstreamBase": upstreamBase, "procedure": "director.updateConfig"}})
		return
	}

	// Native Go fallback: write to local config file
	configPath := filepath.Join(s.cfg.WorkspaceRoot, ".tormentnexus", "config.json")
	os.MkdirAll(filepath.Dir(configPath), 0755)

	existing := localSettingsConfig(s.cfg.WorkspaceRoot)
	for k, v := range payload {
		existing[k] = v
	}

	raw, _ := json.MarshalIndent(existing, "", "  ")
	if err := os.WriteFile(configPath, raw, 0644); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": fmt.Sprintf("failed to write config: %v", err)})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    map[string]any{"ok": true, "updated": len(payload)},
		"bridge":  map[string]any{"fallback": "go-local-director", "procedure": "director.updateConfig", "reason": "upstream unavailable; updated local .tormentnexus/config.json"},
	})
}

func (s *Server) handleDirectorConfigGet(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "directorConfig.get", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": result, "bridge": map[string]any{"upstreamBase": upstreamBase, "procedure": "directorConfig.get"}})
		return
	}

	result = localSettingsConfig(s.cfg.WorkspaceRoot)
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    result,
		"bridge":  map[string]any{"fallback": "go-local-tormentnexus-config", "procedure": "directorConfig.get", "reason": "upstream unavailable; using local .tormentnexus/config.json"},
	})
}

func (s *Server) handleDirectorConfigTest(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "directorConfig.test", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": result, "bridge": map[string]any{"upstreamBase": upstreamBase, "procedure": "directorConfig.test"}})
		return
	}

	// Native Go fallback: validate the local config file exists and is parseable
	configPath := filepath.Join(s.cfg.WorkspaceRoot, ".tormentnexus", "config.json")
	cfg := localSettingsConfig(s.cfg.WorkspaceRoot)
	errors := []string{}
	warnings := []string{}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		errors = append(errors, "config file not found at .tormentnexus/config.json")
	}

	if cfg == nil || len(cfg) == 0 {
		warnings = append(warnings, "config file is empty or contains no settings")
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"valid":    len(errors) == 0,
			"errors":   errors,
			"warnings": warnings,
			"path":     configPath,
			"keys": func() []string {
				keys := make([]string, 0, len(cfg))
				for k := range cfg {
					keys = append(keys, k)
				}
				return keys
			}(),
		},
		"bridge": map[string]any{"fallback": "go-local-director", "procedure": "directorConfig.test", "reason": "upstream unavailable; validated local config"},
	})
}

func (s *Server) handleDirectorConfigUpdate(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Key   string `json:"key"`
		Value any    `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "directorConfig.update", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": result, "bridge": map[string]any{"upstreamBase": upstreamBase, "procedure": "directorConfig.update"}})
		return
	}

	// Native Go fallback: update single key in local config
	configPath := filepath.Join(s.cfg.WorkspaceRoot, ".tormentnexus", "config.json")
	os.MkdirAll(filepath.Dir(configPath), 0755)
	cfg := localSettingsConfig(s.cfg.WorkspaceRoot)
	if cfg == nil {
		cfg = map[string]any{}
	}
	cfg[payload.Key] = payload.Value

	raw, _ := json.MarshalIndent(cfg, "", "  ")
	if err := os.WriteFile(configPath, raw, 0644); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": fmt.Sprintf("failed to write config: %v", err)})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    map[string]any{"ok": true, "key": payload.Key},
		"bridge":  map[string]any{"fallback": "go-local-director", "procedure": "directorConfig.update", "reason": "upstream unavailable; updated local config"},
	})
}

func (s *Server) handleDirectorStopAutoDrive(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "director.stopAutoDrive", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": result, "bridge": map[string]any{"upstreamBase": upstreamBase, "procedure": "director.stopAutoDrive"}})
		return
	}

	// Native Go fallback: update lifecycle modes
	s.lifecycleModes["autoDrive"] = false
	s.lifecycleModes["autoDriveActive"] = false

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    map[string]any{"ok": true, "autoDrive": false},
		"bridge":  map[string]any{"fallback": "go-local-director", "procedure": "director.stopAutoDrive", "reason": "upstream unavailable; stopped auto-drive locally"},
	})
}

func (s *Server) handleDirectorStartAutoDrive(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "director.startAutoDrive", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": result, "bridge": map[string]any{"upstreamBase": upstreamBase, "procedure": "director.startAutoDrive"}})
		return
	}

	// Native Go fallback: update lifecycle modes and start autonomous task
	s.lifecycleModes["autoDrive"] = true
	s.lifecycleModes["autoDriveActive"] = true

	if s.goDirector != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
			defer cancel()
			_ = s.goDirector.StartAutonomousTask(ctx, "Auto-drive: continue current objective")
		}()
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    map[string]any{"ok": true, "autoDrive": true},
		"bridge":  map[string]any{"fallback": "go-local-director", "procedure": "director.startAutoDrive", "reason": "upstream unavailable; started auto-drive locally"},
	})
}

func (s *Server) handleDirectorNotesList(w http.ResponseWriter, r *http.Request) {
	// Try upstream first
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "director.getNotes", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge":  map[string]any{"upstreamBase": upstreamBase, "procedure": "director.getNotes"},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    s.directorNotes.GetNotes(),
		"bridge":  map[string]any{"fallback": "go-local-director-notes"},
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
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": result, "bridge": map[string]any{"upstreamBase": upstreamBase, "procedure": "director.synthesizeNote"}})
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
		"bridge":  map[string]any{"fallback": "go-local-director-notes"},
	})
}
