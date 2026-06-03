package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/tormentnexushq/tormentnexus-go/internal/memorystore"
)

// handleMemoryHydrate triggers the memory hydration engine to scan the
// workspace and populate the context store for autonomous operation.
func (s *Server) handleMemoryHydrate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	hs := memorystore.NewHydrationStore(s.cfg.WorkspaceRoot)
	report, err := hs.HydrateFromWorkspace(r.Context(), s.cfg.WorkspaceRoot)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    report,
	})
}

// handleMemoryHydrationStatus returns the current state of the hydration store.
func (s *Server) handleMemoryHydrationStatus(w http.ResponseWriter, r *http.Request) {
	hs := memorystore.NewHydrationStore(s.cfg.WorkspaceRoot)

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"totalEntries":  len(hs.All()),
			"sections":      hs.SectionNames(),
			"sectionCounts": hs.SectionCounts(),
		},
	})
}

// handleMemoryHydrationQuery searches the hydration store.
func (s *Server) handleMemoryHydrationQuery(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	section := r.URL.Query().Get("section")

	hs := memorystore.NewHydrationStore(s.cfg.WorkspaceRoot)

	if section != "" {
		entries := hs.Get(section, r.URL.Query().Get("key"))
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    entries,
		})
		return
	}

	if query != "" {
		entries := hs.Query(query)
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    entries,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    hs.All(),
	})
}

// handleMemoryHydrationAdd manually adds an entry to the hydration store.
func (s *Server) handleMemoryHydrationAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	var req struct {
		Section  string            `json:"section"`
		Key      string            `json:"key"`
		Content  string            `json:"content"`
		Tags     []string          `json:"tags"`
		Metadata map[string]string `json:"metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON"})
		return
	}

	hs := memorystore.NewHydrationStore(s.cfg.WorkspaceRoot)
	hs.Add(memorystore.HydrationEntry{
		Section:  req.Section,
		Key:      req.Key,
		Content:  req.Content,
		Source:   "manual",
		Tags:     req.Tags,
		Metadata: req.Metadata,
	})

	if err := hs.Save(); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Entry added to hydration store",
	})
}
