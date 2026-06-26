package httpapi

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/supervisor"
)

func (s *Server) handleSupervisorSessionList(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "session.list", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "session.list",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    s.supervisorManager.ListSessions(),
		"bridge": map[string]any{
			"fallback":  "go-local-supervisor",
			"procedure": "session.list",
			"reason":    "upstream unavailable; using native Go supervised session inventory",
		},
	})
}

func (s *Server) handleSupervisorSessionGet(w http.ResponseWriter, r *http.Request) {
	sessionID := strings.TrimSpace(r.URL.Query().Get("id"))
	if sessionID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "missing id query parameter",
		})
		return
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "session.get", map[string]any{"id": sessionID}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "session.get",
			},
		})
		return
	}

	session, ok := s.supervisorManager.GetSession(sessionID)
	if !ok {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    nil,
			"bridge": map[string]any{
				"fallback":  "go-local-supervisor",
				"procedure": "session.get",
				"reason":    "upstream unavailable; session not present in native Go supervised session inventory",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    session,
		"bridge": map[string]any{
			"fallback":  "go-local-supervisor",
			"procedure": "session.get",
			"reason":    "upstream unavailable; using native Go supervised session snapshot",
		},
	})
}

func (s *Server) handleSupervisorSessionCreate(w http.ResponseWriter, r *http.Request) {
	var payload map[string]any
	if err := decodeJSONBody(r, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "session.create", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "session.create",
			},
		})
		return
	}

	cliType := strings.TrimSpace(stringValue(payload["cliType"]))
	workingDirectory := strings.TrimSpace(stringValue(payload["workingDirectory"]))
	if cliType == "" || workingDirectory == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing cliType or workingDirectory"})
		return
	}

	command := strings.TrimSpace(stringValue(payload["command"]))
	args := stringArray(payload["args"])
	if command == "" {
		command, args = defaultLocalSessionCommand(cliType, args)
	}
	session, createErr := s.supervisorManager.CreateSessionWithOptions(supervisor.CreateSessionOptions{
		ID:                  strings.TrimSpace(stringValue(payload["id"])),
		Name:                strings.TrimSpace(stringValue(payload["name"])),
		CliType:             cliType,
		Command:             command,
		Args:                args,
		Env:                 stringMap(payload["env"]),
		RequestedWorkingDir: workingDirectory,
		WorkingDirectory:    workingDirectory,
		ExecutionProfile:    strings.TrimSpace(stringValue(payload["executionProfile"])),
		AutoRestart:         localOptionalBool(payload, "autoRestart", true),
		IsolateWorktree:     localOptionalBool(payload, "isolateWorktree", false),
		Metadata:            mapValue(payload["metadata"]),
		MaxRestarts:         localMaxInt(0, intNumber(payload["maxRestartAttempts"])),
	})
	if createErr != nil {
		writeJSON(w, http.StatusConflict, map[string]any{"success": false, "error": createErr.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    session,
		"bridge": map[string]any{
			"fallback":  "go-local-supervisor",
			"procedure": "session.create",
			"reason":    "upstream unavailable; using native Go in-memory supervised session manager",
		},
	})
}

func (s *Server) handleSupervisorSessionStart(w http.ResponseWriter, r *http.Request) {
	var payload map[string]any
	if err := decodeJSONBody(r, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "session.start", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "session.start",
			},
		})
		return
	}

	sessionID := strings.TrimSpace(stringValue(payload["id"]))
	if sessionID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing id"})
		return
	}
	if err := s.supervisorManager.StartSession(context.WithoutCancel(r.Context()), sessionID); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": err.Error()})
		return
	}
	session, _ := s.supervisorManager.GetSession(sessionID)
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    session,
		"bridge": map[string]any{
			"fallback":  "go-local-supervisor",
			"procedure": "session.start",
			"reason":    "upstream unavailable; using native Go supervised session runtime",
		},
	})
}

func (s *Server) handleSupervisorSessionStop(w http.ResponseWriter, r *http.Request) {
	var payload map[string]any
	if err := decodeJSONBody(r, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "session.stop", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "session.stop",
			},
		})
		return
	}

	sessionID := strings.TrimSpace(stringValue(payload["id"]))
	if sessionID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing id"})
		return
	}
	if err := s.supervisorManager.StopSessionWithOptions(sessionID, localOptionalBool(payload, "force", false)); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": err.Error()})
		return
	}
	session, _ := s.supervisorManager.GetSession(sessionID)
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    session,
		"bridge": map[string]any{
			"fallback":  "go-local-supervisor",
			"procedure": "session.stop",
			"reason":    "upstream unavailable; using native Go supervised session runtime",
		},
	})
}

func (s *Server) handleSupervisorSessionRestart(w http.ResponseWriter, r *http.Request) {
	var payload map[string]any
	if err := decodeJSONBody(r, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "session.restart", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "session.restart",
			},
		})
		return
	}

	sessionID := strings.TrimSpace(stringValue(payload["id"]))
	if sessionID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing id"})
		return
	}
	if err := s.supervisorManager.RestartSession(context.WithoutCancel(r.Context()), sessionID); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": err.Error()})
		return
	}
	session, _ := s.supervisorManager.GetSession(sessionID)
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    session,
		"bridge": map[string]any{
			"fallback":  "go-local-supervisor",
			"procedure": "session.restart",
			"reason":    "upstream unavailable; using native Go supervised session restart flow",
		},
	})
}

func (s *Server) handleSupervisorSessionLogs(w http.ResponseWriter, r *http.Request) {
	sessionID := strings.TrimSpace(r.URL.Query().Get("id"))
	if sessionID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "missing id query parameter",
		})
		return
	}

	payload := map[string]any{"id": sessionID}
	parsedLimit := 0
	if limit := strings.TrimSpace(r.URL.Query().Get("limit")); limit != "" {
		parsed, err := strconv.Atoi(limit)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{
				"success": false,
				"error":   "invalid limit query parameter",
			})
			return
		}
		payload["limit"] = parsed
		parsedLimit = parsed
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "session.logs", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "session.logs",
			},
		})
		return
	}

	logs, localErr := s.supervisorManager.GetSessionLogs(sessionID, parsedLimit)
	if localErr != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    []supervisor.SessionLogEntry{},
			"bridge": map[string]any{
				"fallback":  "go-local-supervisor",
				"procedure": "session.logs",
				"reason":    "upstream unavailable; session not present in native Go supervised session log buffer",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    logs,
		"bridge": map[string]any{
			"fallback":  "go-local-supervisor",
			"procedure": "session.logs",
			"reason":    "upstream unavailable; using native Go supervised session log buffer",
		},
	})
}

func (s *Server) handleSupervisorSessionExecuteShell(w http.ResponseWriter, r *http.Request) {
	var payload map[string]any
	if err := decodeJSONBody(r, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "session.executeShell", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "session.executeShell",
			},
		})
		return
	}

	sessionID := strings.TrimSpace(stringValue(payload["id"]))
	commandText := strings.TrimSpace(stringValue(payload["command"]))
	if sessionID == "" || commandText == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing id or command"})
		return
	}
	session, ok := s.supervisorManager.GetSession(sessionID)
	if !ok {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": "session not found in native Go supervisor state"})
		return
	}

	startedAt := time.Now()
	timeoutMs := intNumber(payload["timeoutMs"])
	if timeoutMs <= 0 {
		timeoutMs = 30_000
	}
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	shellCommand, shellArgs, shellFamily, shellPath := localShellCommand(commandText, session.ExecutionPolicy)
	cmd := exec.CommandContext(ctx, shellCommand, shellArgs...)
	cmd.Dir = session.WorkingDirectory
	cmd.Env = append([]string{}, os.Environ()...)
	for key, value := range session.Env {
		cmd.Env = append(cmd.Env, key+"="+value)
	}
	output, runErr := cmd.CombinedOutput()
	durationMs := time.Since(startedAt).Milliseconds()
	exitCode := 0
	succeeded := runErr == nil
	stderr := ""
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
			stderr = strings.TrimSpace(string(exitErr.Stderr))
		} else if ctx.Err() != nil {
			exitCode = -1
			stderr = ctx.Err().Error()
		} else {
			exitCode = -1
			stderr = runErr.Error()
		}
	}
	stdout := strings.TrimSpace(string(output))
	combined := strings.TrimSpace(stdout)
	if combined == "" {
		combined = stderr
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"command":     commandText,
			"cwd":         session.WorkingDirectory,
			"shellFamily": shellFamily,
			"shellPath":   shellPath,
			"stdout":      stdout,
			"stderr":      stderr,
			"output":      combined,
			"exitCode":    exitCode,
			"durationMs":  durationMs,
			"succeeded":   succeeded,
		},
		"bridge": map[string]any{
			"fallback":  "go-local-supervisor",
			"procedure": "session.executeShell",
			"reason":    "upstream unavailable; using native Go one-shot shell execution in the supervised session working directory",
		},
	})
}

func (s *Server) handleSupervisorSessionAttachInfo(w http.ResponseWriter, r *http.Request) {
	sessionID := strings.TrimSpace(r.URL.Query().Get("id"))
	if sessionID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "missing id query parameter",
		})
		return
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "session.attachInfo", map[string]any{"id": sessionID}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "session.attachInfo",
			},
		})
		return
	}

	attachInfo, localErr := s.supervisorManager.GetAttachInfo(sessionID)
	if localErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    attachInfo,
		"bridge": map[string]any{
			"fallback":  "go-local-supervisor",
			"procedure": "session.attachInfo",
			"reason":    "upstream unavailable; using native Go attach-readiness snapshot",
		},
	})
}

func (s *Server) handleSupervisorSessionHealth(w http.ResponseWriter, r *http.Request) {
	sessionID := strings.TrimSpace(r.URL.Query().Get("id"))
	if sessionID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "missing id query parameter",
		})
		return
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "session.health", map[string]any{"id": sessionID}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "session.health",
			},
		})
		return
	}

	health, localErr := s.supervisorManager.GetSessionHealth(sessionID)
	if localErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    health,
		"bridge": map[string]any{
			"fallback":  "go-local-supervisor",
			"procedure": "session.health",
			"reason":    "upstream unavailable; using native Go supervised session health snapshot",
		},
	})
}

func (s *Server) handleSupervisorSessionState(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "session.getState", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "session.getState",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    s.sessionState.snapshot(),
		"bridge": map[string]any{
			"fallback":  "go-local-session-state",
			"procedure": "session.getState",
			"reason":    "upstream unavailable; using native Go session-state snapshot",
		},
	})
}

func (s *Server) handleSupervisorSessionUpdateState(w http.ResponseWriter, r *http.Request) {
	var payload map[string]any
	if err := decodeJSONBody(r, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "session.updateState", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "session.updateState",
			},
		})
		return
	}

	nextState := s.sessionState.update(payload)
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"success":            true,
			"toolAdvertisements": []string{},
			"memoryBootstrap":    nil,
			"state":              nextState,
		},
		"bridge": map[string]any{
			"fallback":  "go-local-session-state",
			"procedure": "session.updateState",
			"reason":    "upstream unavailable; using native Go session-state persistence without TS memory/bootstrap enrichment",
		},
	})
}

func (s *Server) handleSupervisorSessionClear(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "session.clear", map[string]any{}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "session.clear",
			},
		})
		return
	}

	s.sessionState.clear()
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    map[string]any{"success": true},
		"bridge": map[string]any{
			"fallback":  "go-local-session-state",
			"procedure": "session.clear",
			"reason":    "upstream unavailable; using native Go session-state reset",
		},
	})
}

func (s *Server) handleSupervisorSessionHeartbeat(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "session.heartbeat", map[string]any{}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "session.heartbeat",
			},
		})
		return
	}

	state := s.sessionState.touch()
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"alive":     true,
			"timestamp": state["lastHeartbeat"],
		},
		"bridge": map[string]any{
			"fallback":  "go-local-session-state",
			"procedure": "session.heartbeat",
			"reason":    "upstream unavailable; using native Go session-state heartbeat",
		},
	})
}

func (s *Server) handleSupervisorSessionRestore(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "session.restore", map[string]any{}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "session.restore",
			},
		})
		return
	}

	restored := s.supervisorManager.RestoreSessions()
	status := s.supervisorManager.GetRestoreStatus()
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"restoredCount":    len(restored),
			"sessions":         restored,
			"autoResumeCount":  status.AutoResumeCount,
			"lastRestoreAt":    status.LastRestoreAt,
			"restoredSessions": restored,
		},
		"bridge": map[string]any{
			"fallback":  "go-local-supervisor",
			"procedure": "session.restore",
			"reason":    "upstream unavailable; reloaded native Go persisted supervisor sessions",
		},
	})
}

func defaultLocalSessionCommand(cliType string, providedArgs []string) (string, []string) {
	switch strings.TrimSpace(cliType) {
	case "factory-droid":
		return "droid", append([]string(nil), providedArgs...)
	default:
		return strings.TrimSpace(cliType), append([]string(nil), providedArgs...)
	}
}

func localOptionalBool(payload map[string]any, key string, fallback bool) bool {
	value, ok := payload[key]
	if !ok {
		return fallback
	}
	boolValue, ok := value.(bool)
	if !ok {
		return fallback
	}
	return boolValue
}

func localShellCommand(command string, policy *supervisor.ExecutionPolicy) (string, []string, string, any) {
	if policy != nil && policy.ShellPath != nil && policy.ShellFamily != nil {
		shellPath := strings.TrimSpace(*policy.ShellPath)
		shellFamily := strings.TrimSpace(*policy.ShellFamily)
		if shellPath != "" && shellFamily != "" {
			switch shellFamily {
			case "powershell":
				return shellPath, []string{"-NoProfile", "-Command", command}, shellFamily, shellPath
			case "cmd":
				return shellPath, []string{"/C", command}, shellFamily, shellPath
			case "wsl":
				return shellPath, []string{"bash", "-lc", command}, shellFamily, shellPath
			default:
				return shellPath, []string{"-lc", command}, shellFamily, shellPath
			}
		}
	}
	if runtime.GOOS == "windows" {
		comspec := strings.TrimSpace(os.Getenv("COMSPEC"))
		if comspec == "" {
			comspec = "cmd"
		}
		return comspec, []string{"/C", command}, "cmd", comspec
	}
	return "sh", []string{"-lc", command}, "posix", "sh"
}

func mapValue(value any) map[string]any {
	if value == nil {
		return map[string]any{}
	}
	mapped, ok := value.(map[string]any)
	if !ok {
		return map[string]any{}
	}
	cloned := make(map[string]any, len(mapped))
	for key, entry := range mapped {
		cloned[key] = entry
	}
	return cloned
}

func localMaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
