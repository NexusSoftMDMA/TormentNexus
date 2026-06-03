package httpapi

import (
	"context"
	"net/http"

	"github.com/tormentnexushq/tormentnexus-go/internal/git"
)

func (s *Server) handleSubmoduleUpdateAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	// Try upstream TS server first for robust logging/status tracking
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "submodule.updateAll", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "submodule.updateAll",
			},
		})
		return
	}

	// Fallback to ultra-fast native Go concurrent submodule update
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	report, fallbackErr := git.UpdateAll(ctx, s.cfg.WorkspaceRoot)
	if fallbackErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"success": false,
			"error":   fallbackErr.Error(),
			"detail":  fallbackErr.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    report,
		"bridge": map[string]any{
			"fallback":  "go-local-git-orchestration",
			"procedure": "submodule.updateAll",
			"reason":    "upstream unavailable; executing native Go concurrent submodule synchronization",
		},
	})
}
