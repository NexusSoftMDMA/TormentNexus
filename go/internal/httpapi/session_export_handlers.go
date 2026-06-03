package httpapi

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/tormentnexushq/tormentnexus-go/internal/sessionimport"
)

func (s *Server) handleNativeSessionExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "POST required"})
		return
	}

	var body struct {
		OutputDir string `json:"outputDir"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	homeDir, _ := os.UserHomeDir()

	result, err := sessionimport.ExportSessions(s.cfg.WorkspaceRoot, homeDir, body.OutputDir)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    result,
	})
}
