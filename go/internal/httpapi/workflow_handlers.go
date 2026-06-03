package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/tormentnexushq/tormentnexus-go/internal/workflow"
)

func (s *Server) handleNativeWorkflowList(w http.ResponseWriter, r *http.Request) {
	workflows := s.workflowEngine.List()
	writeJSON(w, http.StatusOK, map[string]any{
		"success":   true,
		"workflows": workflows,
	})
}

func (s *Server) handleNativeWorkflowGet(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing id"})
		return
	}

	wf, ok := s.workflowEngine.Get(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]any{"success": false, "error": "workflow not found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"success": true, "workflow": wf})
}

func (s *Server) handleNativeWorkflowRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "POST required"})
		return
	}

	var body struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing workflow id"})
		return
	}

	wf, ok := s.workflowEngine.Get(body.ID)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]any{"success": false, "error": "workflow not found"})
		return
	}

	// Run asynchronously using a detached context so execution is not canceled
	// when the originating HTTP request completes.
	runCtx := context.WithoutCancel(r.Context())
	go func() {
		_ = s.workflowEngine.RunWorkflow(runCtx, body.ID)
	}()

	writeJSON(w, http.StatusAccepted, map[string]any{
		"success":  true,
		"message":  "workflow started",
		"workflow": wf,
	})
}

func (s *Server) handleNativeWorkflowCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "POST required"})
		return
	}

	var body struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Steps       []struct {
			ID        string   `json:"id"`
			Name      string   `json:"name"`
			Command   string   `json:"command"`
			DependsOn []string `json:"dependsOn,omitempty"`
		} `json:"steps"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid body"})
		return
	}

	steps := make([]*workflow.Step, 0, len(body.Steps))
	for _, bs := range body.Steps {
		steps = append(steps, workflow.ShellStep(bs.ID, bs.Name, bs.Command, s.cfg.WorkspaceRoot, bs.DependsOn...))
	}

	wf := workflow.NewWorkflow(body.ID, body.Name, body.Description, steps)
	s.workflowEngine.Register(wf)

	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "workflow": wf})
}
