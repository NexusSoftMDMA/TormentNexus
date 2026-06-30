package httpapi

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/tormentnexushq/tormentnexus-go/internal/controlplane"
	"github.com/tormentnexushq/tormentnexus-go/internal/memorystore"
	"github.com/tormentnexushq/tormentnexus-go/internal/tools"
)

func (s *Server) handleGetMemory(w http.ResponseWriter, r *http.Request) {
	s.handleMemoryList(w, r)
}

func (s *Server) handleExecuteCode(w http.ResponseWriter, r *http.Request) {
	s.handleCodeExec(w, r)
}

func (s *Server) handleMemorySearch(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("query"))
	if query == "" {
		query = strings.TrimSpace(r.URL.Query().Get("q"))
	}
	if query == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "missing query parameter",
		})
		return
	}
	limit := 5
	if limitParam := strings.TrimSpace(r.URL.Query().Get("limit")); limitParam != "" {
		if parsed, err := strconv.Atoi(limitParam); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	payload := map[string]any{"query": query, "limit": limit}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "memory.query", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "memory.query",
			},
		})
		return
	}

	results, localErr := s.localMemoryQueryResults(query, limit)
	if localErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   localErr.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    results,
		"bridge": map[string]any{
			"fallback":  "go-local-memory",
			"procedure": "memory.query",
			"reason":    "upstream unavailable; using local persisted memory search",
		},
	})
}

func (s *Server) handleMemoryContexts(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "memory.listContexts", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "memory.listContexts",
			},
		})
		return
	}

	contexts, localErr := s.localMemoryContexts()
	if localErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   localErr.Error(),
		})
		return
	}

	// Format contexts for compatibility
	formatted := make([]map[string]any, 0, len(contexts))
	for index, ctx := range contexts {
		metadata, _ := ctx["metadata"].(map[string]any)
		responseMetadata := cloneMap(metadata)
		responseMetadata["title"] = stringValue(ctx["title"])
		responseMetadata["source"] = stringValue(ctx["source"])
		responseMetadata["createdAt"] = ctx["createdAt"]
		responseMetadata["chunks"] = ctx["chunks"]
		formatted = append(formatted, map[string]any{
			"id":       localMemoryContextID(ctx, index+1),
			"content":  stringValue(ctx["content"]),
			"metadata": responseMetadata,
			"score":    1,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    formatted,
		"bridge": map[string]any{
			"fallback":  "go-local-memory",
			"procedure": "memory.listContexts",
			"reason":    "upstream unavailable; using local persisted contexts",
		},
	})
}

func (s *Server) handleMemorySectionedStatus(w http.ResponseWriter, r *http.Request) {
	status, err := memorystore.ReadStatus(s.cfg.WorkspaceRoot)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    status,
	})
}

func (s *Server) handleMemoryFTSearch(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		query = strings.TrimSpace(r.URL.Query().Get("query"))
	}
	if query == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing query"})
		return
	}
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if p, err := strconv.Atoi(l); err == nil && p > 0 && p <= 100 {
			limit = p
		}
	}
	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if p, err := strconv.Atoi(o); err == nil && p >= 0 {
			offset = p
		}
	}
	includeCold := r.URL.Query().Get("cold") == "true"

	if tools.GlobalVectorStore == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": "vector store not initialized"})
		return
	}

	fts, err := memorystore.NewFTSMemorySearch(tools.GlobalVectorStore.DB())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}

	results, err := fts.Search(r.Context(), query, includeCold, limit, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    results.Results,
		"total":   results.Total,
		"offset":  results.Offset,
		"limit":   results.Limit,
	})
}

func (s *Server) handleColdArchiveSearch(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing q query parameter"})
		return
	}
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if p, err := strconv.Atoi(l); err == nil && p > 0 && p <= 100 {
			limit = p
		}
	}
	dbPath := filepath.Join(s.cfg.ConfigDir, "l3_cold_archive.db")
	cold, err := memorystore.NewColdArchive(dbPath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}
	defer cold.Close()
	results, err := cold.SearchCold(r.Context(), query, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": results, "total": len(results)})
}

func (s *Server) handleColdArchiveCount(w http.ResponseWriter, r *http.Request) {
	dbPath := filepath.Join(s.cfg.ConfigDir, "l3_cold_archive.db")
	cold, err := memorystore.NewColdArchive(dbPath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}
	defer cold.Close()
	count, err := cold.Count(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "count": count})
}

func (s *Server) handleColdArchivePromote(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON"})
		return
	}
	if req.ID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing id"})
		return
	}
	dbPath := filepath.Join(s.cfg.ConfigDir, "l3_cold_archive.db")
	cold, err := memorystore.NewColdArchive(dbPath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}
	defer cold.Close()
	record, err := cold.Promote(r.Context(), req.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}
	if record != nil && tools.GlobalVectorStore != nil {
		_ = tools.GlobalVectorStore.Store(r.Context(), *record)
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": record})
}

func (s *Server) handleEnterpriseLicense(w http.ResponseWriter, r *http.Request) {
	if s.enterpriseWrapper == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    map[string]any{"valid": false, "licensedTo": "", "features": []string{}},
		})
		return
	}
	info := s.enterpriseWrapper.Info()
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": info})
}

func (s *Server) handleEnterpriseAudit(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if p, err := strconv.Atoi(l); err == nil && p > 0 && p <= 100 {
			limit = p
		}
	}
	if s.auditor == nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": []map[string]any{}})
		return
	}
	logs := s.auditor.Recent(limit)
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": logs})
}

func (s *Server) handleEnterpriseRoles(w http.ResponseWriter, r *http.Request) {
	if s.enterpriseWrapper == nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": []map[string]any{}})
		return
	}
	roles := s.enterpriseWrapper.GetRoles()
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": roles})
}

func (s *Server) handleMemoryArchiveSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{
			"success": false,
			"error":   "method not allowed",
		})
		return
	}

	var req struct {
		SessionID string   `json:"sessionId"`
		History   []string `json:"history"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "invalid JSON payload: " + err.Error(),
		})
		return
	}

	if req.SessionID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "missing sessionId",
		})
		return
	}

	if s.memoryArchiver == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"success": false,
			"error":   "memory archiver not initialized",
		})
		return
	}

	err := s.memoryArchiver.TakeSnapshot(r.Context(), req.SessionID, req.History)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   "failed to take session snapshot: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
	})
}

func (s *Server) handleLimboBury(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}
	if tools.GlobalVectorStore == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": "vector store not initialized"})
		return
	}
	limbo, err := memorystore.NewLimboVault(tools.GlobalVectorStore.DB())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}
	var req struct {
		ID     string `json:"id"`
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON"})
		return
	}
	rec := controlplane.L2VaultRecord{ID: req.ID}
	reason := memorystore.LimboReason(req.Reason)
	switch reason {
	case memorystore.LimboLost, memorystore.LimboForgotten, memorystore.LimboDiscarded, memorystore.LimboDecayed, memorystore.LimboReplaced:
	default:
		reason = memorystore.LimboDiscarded
	}
	if err := limbo.Bury(r.Context(), rec, reason); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

func (s *Server) handleLimboSearch(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if p, err := strconv.Atoi(l); err == nil && p > 0 && p <= 100 {
			limit = p
		}
	}
	if tools.GlobalVectorStore == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": "vector store not initialized"})
		return
	}
	limbo, err := memorystore.NewLimboVault(tools.GlobalVectorStore.DB())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}
	if query != "" {
		results, err := limbo.SearchLimbo(r.Context(), query, limit)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": results})
		return
	}
	stats, err := limbo.Stats(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "stats": stats})
}

func (s *Server) handleLimboResurrect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}
	if tools.GlobalVectorStore == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": "vector store not initialized"})
		return
	}
	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON"})
		return
	}
	limbo, err := memorystore.NewLimboVault(tools.GlobalVectorStore.DB())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}
	rec, err := limbo.Resurrect(r.Context(), req.ID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"success": false, "error": err.Error()})
		return
	}
	if err := tools.GlobalVectorStore.Commit(r.Context(), *rec); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": rec})
}
