package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/tormentnexushq/tormentnexus-go/internal/mcp"
)

// handleNativeRouterSearch searches the Go-native MCP router (L1+L2).
func (s *Server) handleNativeRouterSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	profile := r.URL.Query().Get("profile")

	if query == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "query required"})
		return
	}

	if s.nativeRouter == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"success": false,
			"error":   "Go-native MCP router not initialized",
		})
		return
	}

	results, loaded, err := s.nativeRouter.SearchAndAutoLoad(r.Context(), query, profile)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"results":     results,
			"autoLoaded":  loaded,
			"workingSet":  s.nativeRouter.GetWorkingSet(),
		},
		"bridge": map[string]any{
			"source": "go-native-mcp-router",
			"layers": "L1+L2+L3",
		},
	})
}

// handleNativeRouterWorkingSet returns the current working set.
func (s *Server) handleNativeRouterWorkingSet(w http.ResponseWriter, r *http.Request) {
	if s.nativeRouter == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"success": false,
			"error":   "Go-native MCP router not initialized",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    s.nativeRouter.GetWorkingSet(),
	})
}

// handleNativeRouterLoad loads a tool into the working set.
func (s *Server) handleNativeRouterLoad(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON"})
		return
	}

	if s.nativeRouter == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": "Go-native MCP router not initialized"})
		return
	}

	entry, err := s.nativeRouter.LoadTool(r.Context(), req.Name)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"success": false, "error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    entry,
	})
}

// handleNativeRouterUnload removes a tool from the working set.
func (s *Server) handleNativeRouterUnload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON"})
		return
	}

	if s.nativeRouter == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": "Go-native MCP router not initialized"})
		return
	}

	if err := s.nativeRouter.UnloadTool(req.Name); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "tool unloaded"})
}

// handleNativeRouterState returns the full router state for diagnostics.
func (s *Server) handleNativeRouterState(w http.ResponseWriter, r *http.Request) {
	if s.nativeRouter == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"success": false,
			"error":   "Go-native MCP router not initialized",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    json.RawMessage(s.nativeRouter.MarshalState()),
	})
}

// handleNativeRouterRefreshCatalog refreshes the catalog from live inventory.
func (s *Server) handleNativeRouterRefreshCatalog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	if s.nativeRouter == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": "Go-native MCP router not initialized"})
		return
	}

	inventory, err := mcp.LoadInventory(s.cfg.WorkspaceRoot, s.cfg.MainConfigDir)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}

	count := s.nativeRouter.RefreshCatalog(inventory)

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"catalogEntries": count,
			"workingSet":     s.nativeRouter.GetWorkingSet(),
		},
	})
}
