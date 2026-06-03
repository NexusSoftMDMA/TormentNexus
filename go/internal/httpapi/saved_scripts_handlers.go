package httpapi

/**
 * @file saved_scripts_handlers.go
 * @module go/internal/httpapi
 *
 * WHAT: Go-native handlers for the Saved Scripts subsystem.
 * Provides CRUD operations for user-created scripts and execution endpoints.
 *
 * WHY: The TS side has savedScriptsRouter.ts with list/get/create/update/delete/execute.
 * The Go sidecar needs native handlers so the dashboard works even when TS is unavailable.
 *
 * PORTING STATUS: Full parity with TS savedScriptsRouter.ts.
 */

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/config"
	"github.com/google/uuid"
)

// SavedScript represents a user-created script with metadata.
type SavedScript struct {
	UUID        string  `json:"uuid"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Code        string  `json:"code"`
	CreatedAt   string  `json:"createdAt,omitempty"`
	UpdatedAt   string  `json:"updatedAt,omitempty"`
}

// ScriptExecutionResult captures the result of running a saved script.
type ScriptExecutionResult struct {
	Success   bool            `json:"success"`
	Result    *string         `json:"result,omitempty"`
	Error     *string         `json:"error,omitempty"`
	Execution *ScriptExecMeta `json:"execution,omitempty"`
}

// ScriptExecMeta holds timing metadata for a script execution.
type ScriptExecMeta struct {
	ScriptUUID string `json:"scriptUuid"`
	ScriptName string `json:"scriptName"`
	StartedAt  string `json:"startedAt"`
	FinishedAt string `json:"finishedAt"`
	DurationMs int64  `json:"durationMs"`
}

// scriptStore manages saved scripts on disk.
// Scripts are stored as JSON files in the tormentnexus data directory.
type scriptStore struct {
	baseDir string
}

func newScriptStore(cfg config.Config) *scriptStore {
	dir := filepath.Join(cfg.MainConfigDir, "scripts")
	_ = os.MkdirAll(dir, 0755)
	return &scriptStore{baseDir: dir}
}

func (ss *scriptStore) list() ([]SavedScript, error) {
	entries, err := os.ReadDir(ss.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []SavedScript{}, nil
		}
		return nil, err
	}

	var scripts []SavedScript
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(ss.baseDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var script SavedScript
		if json.Unmarshal(data, &script) != nil {
			continue
		}
		scripts = append(scripts, script)
	}
	return scripts, nil
}

func (ss *scriptStore) get(scriptUUID string) (*SavedScript, error) {
	path := filepath.Join(ss.baseDir, scriptUUID+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("script not found: %s", scriptUUID)
	}
	var script SavedScript
	if err := json.Unmarshal(data, &script); err != nil {
		return nil, fmt.Errorf("invalid script data: %w", err)
	}
	return &script, nil
}

func (ss *scriptStore) save(script *SavedScript) error {
	if script.UUID == "" {
		script.UUID = uuid.New().String()
	}
	script.UpdatedAt = time.Now().Format(time.RFC3339)
	if script.CreatedAt == "" {
		script.CreatedAt = script.UpdatedAt
	}

	path := filepath.Join(ss.baseDir, script.UUID+".json")
	data, err := json.MarshalIndent(script, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal script: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func (ss *scriptStore) delete(scriptUUID string) error {
	path := filepath.Join(ss.baseDir, scriptUUID+".json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("script not found: %s", scriptUUID)
	}
	return os.Remove(path)
}

// registerSavedScriptRoutes is a legacy entry point — all /api/scripts/* routes
// are already registered in registerRoutes(). This function is kept for
// compatibility but does not add duplicate routes.
func (s *Server) registerSavedScriptRoutes() {
	// All /api/scripts/* routes are registered in registerRoutes().
}

func (s *Server) scriptStore() *scriptStore {
	return newScriptStore(s.cfg)
}

// handleScriptsList returns all saved scripts.
func (s *Server) handleScriptsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	// First try upstream TS server
	var upstreamData any
	_, err := s.callUpstreamJSON(r.Context(), "scripts.list", map[string]any{}, &upstreamData)
	if err == nil {
		writeJSON(w, http.StatusOK, upstreamData)
		return
	}

	// Fallback to local store
	store := s.scriptStore()
	scripts, err := store.list()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	if scripts == nil {
		scripts = []SavedScript{}
	}
	writeJSON(w, http.StatusOK, scripts)
}

// handleScriptsGet returns a single script by UUID.
func (s *Server) handleScriptsGet(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UUID string `json:"uuid"`
	}
	if !readRequestBody(w, r, &input) {
		return
	}

	// Try upstream first
	var upstreamData any
	_, err := s.callUpstreamJSON(r.Context(), "scripts.get", map[string]any{"uuid": input.UUID}, &upstreamData)
	if err == nil {
		writeJSON(w, http.StatusOK, upstreamData)
		return
	}

	// Fallback to local
	store := s.scriptStore()
	script, err := store.get(input.UUID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, script)
}

// handleScriptsCreate creates a new saved script.
func (s *Server) handleScriptsCreate(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
		Code        string  `json:"code"`
	}
	if !readRequestBody(w, r, &input) {
		return
	}

	if input.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}
	if input.Code == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "code is required"})
		return
	}

	// Try upstream first
	var upstreamData any
	_, err := s.callUpstreamJSON(r.Context(), "scripts.create", map[string]any{
		"name": input.Name, "description": input.Description, "code": input.Code,
	}, &upstreamData)
	if err == nil {
		writeJSON(w, http.StatusOK, upstreamData)
		return
	}

	// Fallback: save locally
	script := &SavedScript{
		Name:        input.Name,
		Description: input.Description,
		Code:        input.Code,
	}
	if err := s.scriptStore().save(script); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, script)
}

// handleScriptsUpdate updates an existing saved script.
func (s *Server) handleScriptsUpdate(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UUID        string  `json:"uuid"`
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Code        *string `json:"code"`
	}
	if !readRequestBody(w, r, &input) {
		return
	}

	// Try upstream first
	var upstreamData any
	_, err := s.callUpstreamJSON(r.Context(), "scripts.update", input, &upstreamData)
	if err == nil {
		writeJSON(w, http.StatusOK, upstreamData)
		return
	}

	// Fallback: update locally
	store := s.scriptStore()
	script, err := store.get(input.UUID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	if input.Name != nil {
		script.Name = *input.Name
	}
	if input.Description != nil {
		script.Description = input.Description
	}
	if input.Code != nil {
		script.Code = *input.Code
	}

	if err := store.save(script); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, script)
}

// handleScriptsDelete deletes a saved script by UUID.
func (s *Server) handleScriptsDelete(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UUID string `json:"uuid"`
	}
	if !readRequestBody(w, r, &input) {
		return
	}

	// Try upstream first
	var upstreamData any
	_, err := s.callUpstreamJSON(r.Context(), "scripts.delete", map[string]any{"uuid": input.UUID}, &upstreamData)
	if err == nil {
		writeJSON(w, http.StatusOK, upstreamData)
		return
	}

	// Fallback: delete locally
	if err := s.scriptStore().delete(input.UUID); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"success": "true", "message": "Script deleted"})
}

// handleScriptsExecute runs a saved script and returns the output.
func (s *Server) handleScriptsExecute(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UUID string `json:"uuid"`
	}
	if !readRequestBody(w, r, &input) {
		return
	}

	// Try upstream first
	var upstreamData any
	_, err := s.callUpstreamJSON(r.Context(), "scripts.execute", map[string]any{"uuid": input.UUID}, &upstreamData)
	if err == nil {
		writeJSON(w, http.StatusOK, upstreamData)
		return
	}

	// Fallback: execute locally
	store := s.scriptStore()
	script, err := store.get(input.UUID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	startedAt := time.Now()
	cmd := exec.CommandContext(r.Context(), "sh", "-c", script.Code)
	cmd.Dir = s.cfg.WorkspaceRoot

	output, execErr := cmd.CombinedOutput()
	finishedAt := time.Now()
	durationMs := finishedAt.Sub(startedAt).Milliseconds()

	result := ScriptExecutionResult{
		Execution: &ScriptExecMeta{
			ScriptUUID: script.UUID,
			ScriptName: script.Name,
			StartedAt:  startedAt.Format(time.RFC3339),
			FinishedAt: finishedAt.Format(time.RFC3339),
			DurationMs: durationMs,
		},
	}

	if execErr != nil {
		errMsg := execErr.Error()
		outputStr := string(output)
		if outputStr != "" {
			errMsg = outputStr + "\n" + errMsg
		}
		result.Success = false
		result.Error = &errMsg
	} else {
		result.Success = true
		outputStr := string(output)
		result.Result = &outputStr
	}

	statusCode := http.StatusOK
	if !result.Success {
		statusCode = http.StatusInternalServerError
	}
	writeJSON(w, statusCode, result)
}

// readRequestBody is a helper that reads and decodes JSON from the request body.
func readRequestBody(w http.ResponseWriter, r *http.Request, v any) bool {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		if r.Method == http.MethodGet {
			return true
		}
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return false
	}

	if r.Body == nil || r.ContentLength == 0 {
		return true
	}

	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		if err.Error() != "EOF" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
			return false
		}
	}
	return true
}

// intParam extracts an integer query parameter with a default value.
func intParam(r *http.Request, name string, defaultVal int) int {
	val := r.URL.Query().Get(name)
	if val == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return n
}
