package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/codeexec"
	"github.com/tormentnexushq/tormentnexus-go/internal/enterprise"
)

// handleCodeExec runs code in the process-based sandbox.
func (s *Server) handleCodeExec(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	var req struct {
		Language string `json:"language"`
		Code     string `json:"code"`
		Timeout  int    `json:"timeout,omitempty"` // seconds
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON"})
		return
	}

	if req.Code == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "code is required"})
		return
	}

	timeout := 30 * time.Second
	if req.Timeout > 0 && req.Timeout <= 120 {
		timeout = time.Duration(req.Timeout) * time.Second
	}

	executor := codeexec.NewCodeExecutor()
	result, err := executor.Execute(r.Context(), codeexec.ExecutionConfig{
		Language: codeexec.Language(req.Language),
		Code:     req.Code,
		Timeout:  timeout,
	})

	// Audit Code Execution (Enterprise Tier)
	if s.auditor != nil {
		status := "SUCCESS"
		if err != nil {
			status = "FAILURE: " + err.Error()
		}
		s.auditor.Log(enterprise.AuditEvent{
			UserID:   "system",
			Action:   "EXECUTE_CODE",
			Resource: req.Language,
			Result:   status,
			Metadata: map[string]string{"code_snippet": truncateString(req.Code, 100)},
		})
	}

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    result,
		"bridge": map[string]any{
			"source":    "go-native-code-exec",
			"sandboxed": false,
		},
	})
}

// handleWASMExec runs Go code in the WebAssembly sandbox.
func (s *Server) handleWASMExec(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	var req struct {
		Code       string `json:"code"`
		MaxMemory  int    `json:"maxMemoryMB,omitempty"`
		TimeoutSec int    `json:"timeoutSec,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON"})
		return
	}

	if req.Code == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "code is required"})
		return
	}

	cfg := codeexec.WASMSandboxConfig{
		MaxMemoryMB: req.MaxMemory,
	}
	if req.TimeoutSec > 0 {
		cfg.MaxTimeout = time.Duration(req.TimeoutSec) * time.Second
	}

	sandbox := codeexec.NewWASMSandbox(cfg)
	if !sandbox.IsAvailable() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"success": false,
			"error":   "WASM sandbox not available: need Go compiler (GOOS=js/GOARCH=wasm) and a WASM runtime (wasmtime/wasmer)",
			"hint":    "Install wasmtime: https://wasmtime.dev/ or wasmer: https://wasmer.io/",
		})
		return
	}

	result, err := sandbox.Execute(r.Context(), req.Code)

	// Audit WASM Execution (Enterprise Tier)
	if s.auditor != nil {
		status := "SUCCESS"
		if err != nil {
			status = "FAILURE: " + err.Error()
		}
		s.auditor.Log(enterprise.AuditEvent{
			UserID:   "system",
			Action:   "EXECUTE_WASM",
			Resource: "go",
			Result:   status,
			Metadata: map[string]string{"code_snippet": truncateString(req.Code, 100)},
		})
	}

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    result,
		"bridge": map[string]any{
			"source":    "go-wasm-sandbox",
			"sandboxed": true,
		},
	})
}

// handleWASMStatus reports the WASM sandbox availability and stats.
func (s *Server) handleWASMStatus(w http.ResponseWriter, r *http.Request) {
	sandbox := codeexec.NewWASMSandbox(codeexec.WASMSandboxConfig{})

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"available":           sandbox.IsAvailable(),
			"availableLanguages":  codeexec.ListAvailableLanguages(),
			"stats":               sandbox.Stats(),
		},
	})
}
