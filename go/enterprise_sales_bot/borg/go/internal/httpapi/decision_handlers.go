package httpapi

/**
 * @file decision_handlers.go
 * @module go/internal/httpapi
 *
 * WHAT: HTTP API handlers for the MCP Decision System and Go-native services.
 * ADDED: v1.0.0-alpha.32
 */

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/cache"
	"github.com/tormentnexushq/tormentnexus-go/internal/ctxharvester"
	"github.com/tormentnexushq/tormentnexus-go/internal/eventbus"
	"github.com/tormentnexushq/tormentnexus-go/internal/healer"
	"github.com/tormentnexushq/tormentnexus-go/internal/mcp"
	"github.com/tormentnexushq/tormentnexus-go/internal/metrics"
	processmanager "github.com/tormentnexushq/tormentnexus-go/internal/process"
	"github.com/tormentnexushq/tormentnexus-go/internal/session"
	"github.com/tormentnexushq/tormentnexus-go/internal/toolregistry"
	"github.com/tormentnexushq/tormentnexus-go/internal/workspaces"
)

// ==================== MCP Decision System ====================

func (s *Server) handleDecisionSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Query string `json:"query"`
		Limit int    `json:"limit,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	results, err := s.mcpDecision.SearchTools(ctx, req.Query)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	if req.Limit > 0 && len(results) > req.Limit {
		results = results[:req.Limit]
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"query":   req.Query,
		"count":   len(results),
		"results": results,
	})
}

func (s *Server) handleDecisionSearchAndCall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Query     string                 `json:"query"`
		Arguments map[string]interface{} `json:"arguments,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	result, err := s.mcpDecision.SearchAndCall(ctx, req.Query, req.Arguments)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"result":  result,
	})
}

func (s *Server) handleDecisionLoad(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		ToolName string `json:"toolName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	lt, err := s.mcpDecision.LoadTool(ctx, req.ToolName)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"tool":    lt,
	})
}

func (s *Server) handleDecisionCall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		ToolName  string                 `json:"toolName"`
		Arguments map[string]interface{} `json:"arguments,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	result, err := s.mcpDecision.CallTool(ctx, req.ToolName, req.Arguments)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleDecisionListLoaded(w http.ResponseWriter, r *http.Request) {
	tools := s.mcpDecision.ListLoadedTools()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"count": len(tools),
		"tools": tools,
	})
}

func (s *Server) handleDecisionUnload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		ToolName string `json:"toolName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if err := s.mcpDecision.UnloadTool(req.ToolName); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "unloaded"})
}

func (s *Server) handleDecisionListAll(w http.ResponseWriter, r *http.Request) {
	overview := s.mcpDecision.ListAllTools()
	writeJSON(w, http.StatusOK, overview)
}

func (s *Server) handleDecisionEvents(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	events := s.mcpDecision.GetEvents(limit)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"count":  len(events),
		"events": events,
	})
}

func (s *Server) handleDecisionCatalogRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}

	inv, err := mcp.LoadInventory(s.cfg.WorkspaceRoot, s.cfg.MainConfigDir)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	s.mcpDecision.RefreshFromInventory(inv)
	s.mcpDecision.AddCatalogEntries(mcp.BuiltinTools())

	count := 0
	overview := s.mcpDecision.ListAllTools()
	if overview != nil {
		count = overview.TotalKnown
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":    true,
		"totalKnown": count,
	})
}

func (s *Server) handleDecisionCatalogSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}

	path := s.cfg.MainConfigDir + "/mcp-catalog.json"
	if err := s.mcpDecision.SaveCatalog(path); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "saved", "path": path})
}

// ==================== Cache ====================

func (s *Server) handleCacheGet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "key required"})
		return
	}

	val, ok := s.cacheService.Get(key)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"key":    key,
		"value":  val,
		"exists": true,
	})
}

func (s *Server) handleCacheSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Key   string      `json:"key"`
		Value interface{} `json:"value"`
		TTL   int64       `json:"ttl,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if req.TTL > 0 {
		s.cacheService.SetTTL(req.Key, req.Value, req.TTL)
	} else {
		s.cacheService.Set(req.Key, req.Value)
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "set"})
}

func (s *Server) handleCacheInvalidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Key string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	s.cacheService.Delete(req.Key)
	writeJSON(w, http.StatusOK, map[string]string{"status": "invalidated"})
}

func (s *Server) handleCacheStats(w http.ResponseWriter, r *http.Request) {
	size, maxSize := s.cacheService.Stats()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"size":    size,
		"maxSize": maxSize,
	})
}

// ==================== Git (native) ====================
// NOTE: /api/git/* routes are handled by existing handlers in server.go.
// These /api/native/git/* routes delegate to the Go-native gitservice.

func (s *Server) handleNativeGitLog(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}

	entries, err := s.gitService.GetLog(limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"count":   len(entries),
		"entries": entries,
	})
}

func (s *Server) handleNativeGitStatus(w http.ResponseWriter, r *http.Request) {
	status, err := s.gitService.GetStatus()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func (s *Server) handleNativeGitDiff(w http.ResponseWriter, r *http.Request) {
	staged := r.URL.Query().Get("staged") == "true"

	diff, err := s.gitService.Diff(staged)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"count": len(diff),
		"diff":  diff,
	})
}

func (s *Server) handleNativeGitBranches(w http.ResponseWriter, r *http.Request) {
	branches, err := s.gitService.ListBranches()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"count":    len(branches),
		"branches": branches,
	})
}

// ==================== Session Manager (native) ====================

func (s *Server) handleSessionList(w http.ResponseWriter, r *http.Request) {
	sessions := s.sessionManager.List()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"count":    len(sessions),
		"sessions": sessions,
	})
}

func (s *Server) handleSessionCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		ID      string `json:"id"`
		CLIType string `json:"cliType"`
		WorkDir string `json:"workDir"`
		Task    string `json:"task"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	sess := s.sessionManager.Create(req.ID, req.CLIType, req.WorkDir, req.Task)
	if err := s.sessionManager.Start(sess.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, sess)
}

func (s *Server) handleSessionGet(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id required"})
		return
	}

	sess, ok := s.sessionManager.Get(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "session not found"})
		return
	}

	writeJSON(w, http.StatusOK, sess)
}

// ==================== Workspaces ====================

func (s *Server) handleWorkspacesList(w http.ResponseWriter, r *http.Request) {
	wsList, err := s.workspaceTracker.ListWorkspaces()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"count":      len(wsList),
		"workspaces": wsList,
	})
}

func (s *Server) handleWorkspacesRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if req.Path == "" {
		req.Path = s.cfg.WorkspaceRoot
	}

	if err := s.workspaceTracker.RegisterWorkspace(req.Path); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "registered", "path": req.Path})
}

// ==================== Metrics ====================

func (s *Server) handleMetricsPrometheus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(s.metricsService.ExportPrometheus()))
}

func (s *Server) handleMetricsCounters(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"metrics": "see prometheus endpoint for full export",
	})
}

// ==================== Tool Registry ====================

func (s *Server) handleNativeToolSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "q parameter required"})
		return
	}

	results := s.toolRegistry.Search(query, 10)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"query":   query,
		"count":   len(results),
		"results": results,
	})
}

func (s *Server) handleNativeToolList(w http.ResponseWriter, r *http.Request) {
	alwaysOn := r.URL.Query().Get("alwaysOn") == "true"
	if alwaysOn {
		allTools := s.toolRegistry.ListAlwaysOn()
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"count": len(allTools),
			"tools": allTools,
		})
		return
	}

	allTools := s.toolRegistry.List()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"count": len(allTools),
		"tools": allTools,
	})
}

func (s *Server) handleNativeToolRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		ServerName  string   `json:"serverName"`
		AlwaysOn    bool     `json:"alwaysOn,omitempty"`
		Tags        []string `json:"tags,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	_ = s.toolRegistry.Register(toolregistry.ToolInfo{
		Name:        req.Name,
		Description: req.Description,
		ServerName:  req.ServerName,
		AlwaysOn:    req.AlwaysOn,
		Tags:        req.Tags,
	})

	writeJSON(w, http.StatusOK, map[string]string{"status": "registered"})
}

// ==================== Healer (native) ====================
// NOTE: /api/healer/* routes are handled by existing handlers in healer_handlers.go.
// These /api/native/healer/* routes delegate to the Go-native healer service.

func (s *Server) handleNativeHealerDiagnose(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Error   string `json:"error"`
		Context string `json:"context,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	result, err := s.healerService.AnalyzeError(ctx, req.Error, req.Context)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleNativeHealerHistory(w http.ResponseWriter, r *http.Request) {
	history := s.healerService.GetHistory()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"count":   len(history),
		"history": history,
	})
}

func (s *Server) handleNativeHealerHeal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Error   string `json:"error"`
		Context string `json:"context,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	success, err := s.healerService.Heal(ctx, req.Error, req.Context)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": success,
	})
}

func (s *Server) handleNativeHealerVault(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		var parsed int
		if _, err := fmt.Sscanf(limitStr, "%d", &parsed); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if s.memoryReactor == nil || s.memoryReactor.VectorStore() == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"count":   0,
			"records": []any{},
			"error":   "vector store not initialized",
		})
		return
	}

	records, err := s.memoryReactor.VectorStore().GetAllVaultRecords(r.Context(), limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	totalCount, _ := s.memoryReactor.VectorStore().GetVaultRecordCount(r.Context())

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"count":      len(records),
		"totalCount": totalCount,
		"records":    records,
	})
}

// ==================== Context Harvester ====================

func (s *Server) handleHarvesterAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Content  string                 `json:"content"`
		Source   string                 `json:"source"`
		Metadata map[string]interface{} `json:"metadata,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	chunks := s.contextHarvester.Harvest(ctxharvester.ContextSource(req.Source), req.Content, req.Metadata)
	ids := make([]string, len(chunks))
	for i, c := range chunks {
		ids[i] = c.ID
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "harvested",
		"count":  len(chunks),
		"ids":    ids,
	})
}

func (s *Server) handleHarvesterSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "q parameter required"})
		return
	}

	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}

	results := s.contextHarvester.Retrieve(query, limit)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"query":   query,
		"count":   len(results),
		"results": results,
	})
}

func (s *Server) handleHarvesterReport(w http.ResponseWriter, r *http.Request) {
	report := s.contextHarvester.GetReport()
	writeJSON(w, http.StatusOK, report)
}

// ==================== Process Manager ====================

func (s *Server) handleProcessSpawn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Command string            `json:"command"`
		Args    []string          `json:"args,omitempty"`
		Env     map[string]string `json:"env,omitempty"`
		Cwd     string            `json:"cwd,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	pid, err := s.processManager.Spawn(processmanager.ProcessConfig{
		Command: req.Command,
		Args:    req.Args,
		Env:     req.Env,
		Cwd:     req.Cwd,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"pid":    pid,
		"status": "spawned",
	})
}

func (s *Server) handleProcessList(w http.ResponseWriter, r *http.Request) {
	sessions := s.processManager.ListActiveSessions()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"count":    len(sessions),
		"sessions": sessions,
		"active":   s.processManager.ActiveCount(),
	})
}

func (s *Server) handleProcessKill(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		SessionID string `json:"sessionId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if !s.processManager.Kill(req.SessionID) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "process not found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "killed"})
}

// ==================== Ensure types compile ====================

var (
	_ = cache.Cache{}
	_ = ctxharvester.ContextHarvester{}
	_ = eventbus.EventBus{}
	_ = healer.HealerService{}
	_ = metrics.MetricsService{}
	_ = processmanager.ProcessManager{}
	_ = session.SessionManager{}
	_ = toolregistry.ToolRegistry{}
	_ = workspaces.WorkspaceTracker{}
	_ = mcp.DecisionSystem{}
)
