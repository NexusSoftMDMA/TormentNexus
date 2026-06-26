package httpapi

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/eventbus"
	"github.com/tormentnexushq/tormentnexus-go/internal/supervisor"
)

// handleTormentNexusProtocol handles inbound tormentnexus:// deep links passed from the OS
// or client dispatchers.
func (s *Server) handleTormentNexusProtocol(w http.ResponseWriter, r *http.Request) {
	var rawURL string

	if r.Method == http.MethodPost {
		var req struct {
			URL string `json:"url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			rawURL = req.URL
		}
	} else {
		rawURL = r.URL.Query().Get("url")
		if rawURL == "" {
			rawURL = r.URL.Query().Get("uri")
		}
	}

	if rawURL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "missing url or uri parameter or JSON body",
		})
		return
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "invalid deep link URL: " + err.Error(),
		})
		return
	}

	// Validate scheme
	if u.Scheme != "tormentnexus" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "invalid URL scheme; expected tormentnexus://",
		})
		return
	}

	// Determine action from host or path
	action := u.Host
	if action == "" {
		action = strings.TrimPrefix(u.Path, "/")
	}
	action = strings.ToLower(action)

	queryParams := u.Query()
	sessionID := queryParams.Get("session")
	if sessionID == "" {
		sessionID = queryParams.Get("id")
	}

	switch action {
	case "attach":
		if sessionID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{
				"success": false,
				"error":   "missing session parameter for attach action",
			})
			return
		}

		if s.supervisorManager == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{
				"success": false,
				"error":  "session supervisor not available",
			})
			return
		}
		session, ok := s.supervisorManager.GetSession(sessionID)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]any{
				"success": false,
				"error":   "session " + sessionID + " not found in supervisor registry",
			})
			return
		}

		// Emit attachment event to SSE and EventBus so frontends focus the session
		if s.eventBus != nil {
			s.eventBus.EmitEvent(eventbus.SystemEventType("session:attach"), "protocol", map[string]any{
				"sessionId":        session.ID,
				"cliType":          session.CliType,
				"workingDirectory": session.WorkingDirectory,
				"timestamp":        time.Now().UnixMilli(),
			})
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data": map[string]any{
				"action":  "attach",
				"session": session,
				"message": "attachment focus signal broadcast successfully",
			},
		})

	case "create":
		cliType := queryParams.Get("cliType")
		workingDir := queryParams.Get("workingDirectory")
		if workingDir == "" {
			workingDir = queryParams.Get("cwd")
		}

		if cliType == "" || workingDir == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{
				"success": false,
				"error":   "missing cliType or workingDirectory parameters",
			})
			return
		}

		// Fallback command mapping
		cmd, args := defaultLocalSessionCommand(cliType, nil)

		if s.supervisorManager == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{
				"success": false,
				"error":  "session supervisor not available",
			})
			return
		}
		session, createErr := s.supervisorManager.CreateSessionWithOptions(supervisor.CreateSessionOptions{
			CliType:             cliType,
			Command:             cmd,
			Args:                args,
			WorkingDirectory:    workingDir,
			RequestedWorkingDir: workingDir,
			AutoRestart:         true,
		})
		if createErr != nil {
			writeJSON(w, http.StatusConflict, map[string]any{
				"success": false,
				"error":   "failed to create session: " + createErr.Error(),
			})
			return
		}

		// Start session automatically
		_ = s.supervisorManager.StartSession(r.Context(), session.ID)

		// Broadcast new session event
		if s.eventBus != nil {
			s.eventBus.EmitEvent(eventbus.SystemEventType("session:attach"), "protocol", map[string]any{
				"sessionId":        session.ID,
				"cliType":          session.CliType,
				"workingDirectory": session.WorkingDirectory,
				"timestamp":        time.Now().UnixMilli(),
			})
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data": map[string]any{
				"action":  "create",
				"session": session,
				"message": "supervised session created and dispatch focus broadcast",
			},
		})

	default:
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "unknown action '" + action + "'; supported: attach, create",
		})
	}
}
