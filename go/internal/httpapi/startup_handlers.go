package httpapi

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/config"
	"github.com/tormentnexushq/tormentnexus-go/internal/interop"
	"github.com/tormentnexushq/tormentnexus-go/internal/memorystore"
	"github.com/tormentnexushq/tormentnexus-go/internal/mesh"
	"github.com/tormentnexushq/tormentnexus-go/internal/cache"
)

type StartupBlockingReason struct {
	Code   string `json:"code"`
	Detail string `json:"detail"`
}

type StartupStatus struct {
	Status          string                  `json:"status"`
	Ready           bool                    `json:"ready"`
	Summary         string                  `json:"summary"`
	BlockingReasons []StartupBlockingReason `json:"blockingReasons"`
	Checks          map[string]any          `json:"checks"`
}

func (s *Server) handleStartupStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	// Cache startup status for 5s (dashboard polls every 5s)
	val, err := cache.Cached(s.cacheService, "startup:status", func() (interface{}, error) {
		return s.buildStartupStatus(r.Context())
	}, 30000)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":   val,
	})
}

func (s *Server) buildStartupStatus(ctx context.Context) (StartupStatus, error) {
	// Run all potentially slow operations in parallel
	type upstreamCheck struct {
		ready   bool
		baseURL string
	}
	type meshResult struct {
		status mesh.Status
		err    error
	}
	type memoryResult struct {
		status memorystore.StoreStatus
		err    error
	}

	upstreamCh := make(chan upstreamCheck, 1)
	supervisorCh := make(chan upstreamCheck, 1)
	meshCh := make(chan meshResult, 1)
	memoryCh := make(chan memoryResult, 1)

	go func() {
		r, b := s.checkUpstreamProcedure(ctx, "health", nil)
		upstreamCh <- upstreamCheck{ready: r, baseURL: b}
	}()
	go func() {
		r, b := s.checkUpstreamProcedure(ctx, "session.list", nil)
		supervisorCh <- upstreamCheck{ready: r, baseURL: b}
	}()
	go func() {
		s, err := s.mesh.Status(ctx)
		meshCh <- meshResult{status: s, err: err}
	}()
	go func() {
		s, err := memorystore.ReadStatus(s.cfg.WorkspaceRoot)
		memoryCh <- memoryResult{status: s, err: err}
	}()

	// Collect results (all run in parallel, total time = slowest)
	upstreamResult := <-upstreamCh
	supervisorResult := <-supervisorCh
	meshRes := <-meshCh
	memoryRes := <-memoryCh

	upstreamReady, upstreamBase := upstreamResult.ready, upstreamResult.baseURL
	supervisorReady, supervisorBase := supervisorResult.ready, supervisorResult.baseURL

	if meshRes.err != nil {
		return StartupStatus{}, meshRes.err
	}
	if memoryRes.err != nil {
		return StartupStatus{}, memoryRes.err
	}

	configStatus := config.Snapshot(s.cfg)
	meshStatus := meshRes.status
	memoryStatus := memoryRes.status
	importedStats := s.importedSessionMaintenanceStats(ctx)

	blockingReasons := make([]StartupBlockingReason, 0, 4)
	if !configStatus.WorkspaceRoot.Exists {
		blockingReasons = append(blockingReasons, StartupBlockingReason{
			Code:   "workspace_root_missing",
			Detail: "Workspace root is not available to the Go sidecar.",
		})
	}
	if !configStatus.ConfigDir.Exists {
		blockingReasons = append(blockingReasons, StartupBlockingReason{
			Code:   "go_config_dir_missing",
			Detail: "Go sidecar config directory has not been created yet.",
		})
	}
	if !memoryStatus.Exists {
		blockingReasons = append(blockingReasons, StartupBlockingReason{
			Code:   "memory_store_not_ready",
			Detail: "Sectioned memory store is not available yet.",
		})
	}
	if !upstreamReady {
		blockingReasons = append(blockingReasons, StartupBlockingReason{
			Code:   "main_control_plane_unreachable",
			Detail: "The Go sidecar cannot currently reach the main TypeScript control plane.",
		})
	}
	if !supervisorReady {
		blockingReasons = append(blockingReasons, StartupBlockingReason{
			Code:   "session_supervisor_bridge_unavailable",
			Detail: "The supervised-session bridge is not reachable through the main control plane.",
		})
	}

	summary := "All Go startup checks passed."
	if len(blockingReasons) > 0 {
		summary = "Startup pending: "
		for i, reason := range blockingReasons {
			if i > 0 {
				summary += " "
			}
			summary += reason.Detail
		}
	}

	return StartupStatus{
		Status:          "running",
		Ready:           len(blockingReasons) == 0,
		Summary:         summary,
		BlockingReasons: blockingReasons,
		Checks: map[string]any{
			"config": map[string]any{
				"workspaceRootAvailable": configStatus.WorkspaceRoot.Exists,
				"goConfigDirAvailable":   configStatus.ConfigDir.Exists,
				"mainConfigDirAvailable": configStatus.MainConfigDir.Exists,
				"repoConfigAvailable":    configStatus.TormentNexusConfigFile.Exists,
				"mcpConfigAvailable":     configStatus.MCPConfigFile.Exists,
			},
			"memory": map[string]any{
				"ready":                   memoryStatus.Exists,
				"storePath":               memoryStatus.StorePath,
				"totalEntries":            memoryStatus.TotalEntries,
				"presentDefaultSections":  memoryStatus.PresentDefaultSectionCount,
				"expectedDefaultSections": memoryStatus.DefaultSectionCount,
				"missingSections":         memoryStatus.MissingSections,
			},
			"mainControlPlane": map[string]any{
				"ready":   upstreamReady,
				"baseUrl": upstreamBase,
			},
			"sessionSupervisorBridge": map[string]any{
				"ready":   supervisorReady,
				"baseUrl": supervisorBase,
			},
			"mesh": map[string]any{
				"nodeId":     meshStatus.NodeID,
				"peersCount": meshStatus.PeersCount,
			},
			"importedSessions": map[string]any{
				"totalSessions":                importedStats.TotalSessions,
				"inlineTranscriptCount":        importedStats.InlineTranscriptCount,
				"archivedTranscriptCount":      importedStats.ArchivedTranscriptCount,
				"missingRetentionSummaryCount": importedStats.MissingRetentionSummaryCount,
			},
		},
	}, nil
}

func (s *Server) importedSessionMaintenanceStats(ctx context.Context) ImportedSessionMaintenanceStats {
	// Fast path: use cached import scan results instead of calling TS core
	candidates, cacheErr := s.scanValidatedImportSources()
	if cacheErr == nil && len(candidates) > 0 {
		// Check archive cache first (avoids re-reading 6000+ gzipped files)
		if cached, ok := s.cacheService.Get("imported:archive:records"); ok {
			if typed, ok := cached.([]ImportedSessionRecord); ok && len(typed) > 0 {
				return archivedImportedSessionMaintenanceStats(typed)
			}
		}
		// Return quick estimate from import cache while archive loads in background
		stats := ImportedSessionMaintenanceStats{
			TotalSessions: len(candidates),
			InlineTranscriptCount: 0,
			ArchivedTranscriptCount: 0,
			MissingRetentionSummaryCount: 0,
		}
		// Kick off background archive load to populate cache for next call
		go func() {
			s.loadArchivedImportedSessionRecords()
		}()
		return stats
	}
	// Slow path: try upstream
	var stats ImportedSessionMaintenanceStats
	if _, err := s.callUpstreamJSON(ctx, "session.importedMaintenanceStats", nil, &stats); err == nil {
		return stats
	}
	// Ensure .tormentnexus/imported_sessions exists
	_ = os.MkdirAll(filepath.Join(s.cfg.WorkspaceRoot, ".tormentnexus", "imported_sessions"), 0755)
	return ImportedSessionMaintenanceStats{}
}

func (s *Server) checkUpstreamProcedure(ctx context.Context, procedure string, payload any) (bool, string) {
	checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	result, err := interop.CallTRPCProcedure(checkCtx, s.cfg.MainLockPath(), procedure, payload)
	if err != nil {
		return false, ""
	}

	return true, result.BaseURL
}
