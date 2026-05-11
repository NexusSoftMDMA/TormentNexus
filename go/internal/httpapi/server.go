package httpapi

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/borghq/borg-go/internal/codeexec"
	"github.com/borghq/borg-go/internal/memorystore"
	"io"
	"io/fs"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/borghq/borg-go/internal/buildinfo"
	"github.com/borghq/borg-go/internal/config"
	"github.com/borghq/borg-go/internal/controlplane"
	"github.com/borghq/borg-go/internal/harnesses"
	"github.com/borghq/borg-go/internal/hsync"
	"github.com/borghq/borg-go/internal/interop"
	"github.com/borghq/borg-go/internal/mcp"
	"github.com/borghq/borg-go/internal/mesh"
	"github.com/borghq/borg-go/internal/orchestration"
	"github.com/borghq/borg-go/internal/providers"
	"github.com/borghq/borg-go/internal/sessionimport"
	"github.com/borghq/borg-go/internal/supervisor"
	"github.com/borghq/borg-go/internal/tools"
	"github.com/borghq/borg-go/internal/workflow"

	"github.com/borghq/borg-go/internal/cache"
	"github.com/borghq/borg-go/internal/ctxharvester"
	"github.com/borghq/borg-go/internal/eventbus"
	"github.com/borghq/borg-go/internal/gitservice"
	"github.com/borghq/borg-go/internal/healer"
	"github.com/borghq/borg-go/internal/metrics"
	processmanager "github.com/borghq/borg-go/internal/process"
	"github.com/borghq/borg-go/internal/repograph"
	"github.com/borghq/borg-go/internal/session"
	"github.com/borghq/borg-go/internal/skillregistry"
	"github.com/borghq/borg-go/internal/toolregistry"
	"github.com/borghq/borg-go/internal/workspaces"
	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

var sessionExportKnownFormats = []map[string]any{
	{"id": "claude-code", "type": "claude-code", "paths": []string{".claude", ".claude/sessions"}},
	{"id": "cursor", "type": "cursor", "paths": []string{".cursor", ".cursor/sessions"}},
	{"id": "opencode", "type": "opencode", "paths": []string{".docs/ai-logs"}},
	{"id": "aider", "type": "aider", "paths": []string{".aider.chat.history.md", ".aider.tags.cache"}},
	{"id": "windsurf", "type": "windsurf", "paths": []string{".windsurf", ".docs/ai-logs"}},
	{"id": "borg", "type": "borg", "paths": []string{".borg", ".borg/sessions"}},
	{"id": "continue", "type": "continue", "paths": []string{".continue", ".continue/sessions"}},
	{"id": "copilot", "type": "copilot", "paths": []string{".github/copilot"}},
}

type Server struct {
	memoryManager     *memorystore.Manager
	codeExecutor      *codeexec.Sandbox
	cfg               config.Config
	detector          controlplane.ToolProvider
	mesh              *mesh.Service
	startedAt         time.Time
	mux               *http.ServeMux
	lifecycleModes    map[string]any
	fallbackBuffer    *providerFallbackBuffer
	autoDev           *localAutoDevManager
	squad             *localSquadManager
	swarm             *localSwarmManager
	debateHistory     *orchestration.DebateHistoryStore
	darwinState       *localDarwinStateManager
	runtimeServers    *runtimeServerRegistry
	supervisorManager *supervisor.Manager
	sessionState      *localSessionStateManager
	workflowEngine    *workflow.Engine
	toolsRegistry     *tools.Registry
	mcpAggregator     *mcp.Aggregator
	mcpPredictor      *mcp.ToolPredictor
	mcpDecision       *mcp.DecisionSystem
	nativeRouter      *mcp.NativeMCPRouter
	a2aLogger         *orchestration.A2ALogger
	a2aBroker         *orchestration.A2ABroker
	taskQueue         *orchestration.TaskQueue
	swarmController   *orchestration.SwarmController
	coderAgent        *orchestration.CoderAgent
	goDirector        *orchestration.Director
	highValueIngestor *hsync.HighValueIngestor
	memoryReactor     *memorystore.MemoryReactor
	mcpConfig         *mcp.ConfigManager
	skillStore        *harnesses.SkillStore
	skillRegistry     *skillregistry.SkillRegistry
	skillDecision     *skillregistry.SkillDecisionSystem
	pairOrchestrator  *orchestration.PairOrchestrator
	directorNotes     *orchestration.DirectorNotesManager
	expertManager     *hsync.ExpertManager
	memoryArchiver    *memorystore.MemoryArchiver

	// --- New Go-native services (alpha.32+) ---
	eventBus         *eventbus.EventBus
	metricsService   *metrics.MetricsService
	sessionManager   *session.SessionManager
	toolRegistry     *toolregistry.ToolRegistry
	gitService       *gitservice.GitService
	contextHarvester *ctxharvester.ContextHarvester
	workspaceTracker *workspaces.WorkspaceTracker
	processManager   *processmanager.ProcessManager
	healerService    *healer.HealerService
	cacheService     *cache.Cache
	repoGraph        *repograph.RepoGraphService
}

type providerFallbackEvent struct {
	ID                int64  `json:"id"`
	Timestamp         int64  `json:"timestamp"`
	RequestedProvider string `json:"requestedProvider,omitempty"`
	SelectedProvider  string `json:"selectedProvider"`
	SelectedModelID   string `json:"selectedModelId"`
	TaskType          string `json:"taskType"`
	Strategy          string `json:"strategy"`
	Reason            string `json:"reason"`
	CauseCode         string `json:"causeCode"`
}

type providerFallbackBuffer struct {
	mu      sync.Mutex
	nextID  int64
	maxSize int
	events  []providerFallbackEvent
}

func newProviderFallbackBuffer(maxSize int) *providerFallbackBuffer {
	return &providerFallbackBuffer{maxSize: maxSize}
}

func (b *providerFallbackBuffer) append(event providerFallbackEvent) providerFallbackEvent {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.nextID++
	event.ID = b.nextID
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().UnixMilli()
	}
	b.events = append(b.events, event)
	if len(b.events) > b.maxSize {
		b.events = append([]providerFallbackEvent(nil), b.events[len(b.events)-b.maxSize:]...)
	}
	return event
}

func (b *providerFallbackBuffer) list(limit int) []providerFallbackEvent {
	b.mu.Lock()
	defer b.mu.Unlock()

	if limit <= 0 {
		limit = 20
	}
	if limit > len(b.events) {
		limit = len(b.events)
	}
	result := make([]providerFallbackEvent, 0, limit)
	for i := len(b.events) - 1; i >= len(b.events)-limit; i-- {
		result = append(result, b.events[i])
	}
	return result
}

func (b *providerFallbackBuffer) clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.events = nil
}

type Session struct {
	ID             string   `json:"id"`
	CLIType        string   `json:"cliType"`
	Status         string   `json:"status"`
	Task           string   `json:"task,omitempty"`
	StartedAt      string   `json:"startedAt,omitempty"`
	SourcePath     string   `json:"sourcePath,omitempty"`
	SessionFormat  string   `json:"sessionFormat,omitempty"`
	Valid          bool     `json:"valid"`
	DetectedModels []string `json:"detectedModels,omitempty"`
}

type RuntimeStatus struct {
	Service              string                       `json:"service"`
	Version              string                       `json:"version"`
	BaseURL              string                       `json:"baseUrl"`
	UptimeSec            int                          `json:"uptimeSec"`
	Locks                []interop.ControlPlaneStatus `json:"locks"`
	LockSummary          LockRuntimeSummary           `json:"lockSummary"`
	Config               ConfigRuntimeSummary         `json:"config"`
	CLI                  CLIRuntimeSummary            `json:"cli"`
	Providers            ProviderRuntimeSummary       `json:"providers"`
	Memory               MemoryRuntimeSummary         `json:"memory"`
	Sessions             SessionRuntimeSummary        `json:"sessions"`
	ImportedInstructions ImportedInstructionsSummary  `json:"importedInstructions"`
	ImportRoots          ImportRootsSummary           `json:"importRoots"`
	ImportSources        ImportSourcesSummary         `json:"importSources"`
}

type ProviderRuntimeSummary struct {
	ProviderCount           int                `json:"providerCount"`
	ConfiguredCount         int                `json:"configuredCount"`
	AuthenticatedCount      int                `json:"authenticatedCount"`
	ExecutableCount         int                `json:"executableCount"`
	RoutingPreviewAvailable bool               `json:"routingPreviewAvailable"`
	ByAuthMethod            []SummaryBucket    `json:"byAuthMethod,omitempty"`
	ByPreferredTask         []SummaryBucket    `json:"byPreferredTask,omitempty"`
	Statuses                []providers.Status `json:"statuses"`
}

type LockRuntimeSummary struct {
	VisibleCount int `json:"visibleCount"`
	RunningCount int `json:"runningCount"`
}

type ConfigRuntimeSummary struct {
	WorkspaceRootAvailable bool `json:"workspaceRootAvailable"`
	ConfigDirAvailable     bool `json:"configDirAvailable"`
	MainConfigDirAvailable bool `json:"mainConfigDirAvailable"`
	RepoConfigAvailable    bool `json:"repoConfigAvailable"`
	MCPConfigAvailable     bool `json:"mcpConfigAvailable"`
	BorgSubmoduleAvailable bool `json:"borgSubmoduleAvailable"`
}

type MemoryRuntimeSummary struct {
	Available                  bool                        `json:"available"`
	StorePath                  string                      `json:"storePath"`
	TotalEntries               int                         `json:"totalEntries"`
	SectionCount               int                         `json:"sectionCount"`
	DefaultSectionCount        int                         `json:"defaultSectionCount"`
	PopulatedSectionCount      int                         `json:"populatedSectionCount"`
	PresentDefaultSectionCount int                         `json:"presentDefaultSectionCount"`
	MissingSections            []string                    `json:"missingSections,omitempty"`
	Sections                   []memorystore.SectionStatus `json:"sections,omitempty"`
	LastUpdatedAt              string                      `json:"lastUpdatedAt,omitempty"`
}

type SessionRuntimeSummary struct {
	DiscoveredCount           int             `json:"discoveredCount"`
	ValidCount                int             `json:"validCount"`
	SupervisorBridgeAvailable bool            `json:"supervisorBridgeAvailable"`
	SupervisorBridgeBase      string          `json:"supervisorBridgeBase,omitempty"`
	ByCLIType                 []SummaryBucket `json:"byCliType,omitempty"`
	ByFormat                  []SummaryBucket `json:"byFormat,omitempty"`
	ByTask                    []SummaryBucket `json:"byTask,omitempty"`
	ByModelHint               []SummaryBucket `json:"byModelHint,omitempty"`
}

type SessionSummary struct {
	Count       int             `json:"count"`
	ValidCount  int             `json:"validCount"`
	ByCLIType   []SummaryBucket `json:"byCliType"`
	ByFormat    []SummaryBucket `json:"byFormat"`
	ByTask      []SummaryBucket `json:"byTask"`
	ByModelHint []SummaryBucket `json:"byModelHint"`
}

type CLIRuntimeSummary struct {
	ToolCount                   int    `json:"toolCount"`
	AvailableToolCount          int    `json:"availableToolCount"`
	HarnessCount                int    `json:"harnessCount"`
	InstalledHarnessCount       int    `json:"installedHarnessCount"`
	SourceBackedHarnessCount    int    `json:"sourceBackedHarnessCount"`
	MetadataOnlyHarnessCount    int    `json:"metadataOnlyHarnessCount"`
	OperatorDefinedHarnessCount int    `json:"operatorDefinedHarnessCount"`
	SourceBackedToolCount       int    `json:"sourceBackedToolCount"`
	PrimaryHarness              string `json:"primaryHarness,omitempty"`
}

type CLISummary struct {
	ToolCount                   int                    `json:"toolCount"`
	AvailableToolCount          int                    `json:"availableToolCount"`
	HarnessCount                int                    `json:"harnessCount"`
	InstalledHarnessCount       int                    `json:"installedHarnessCount"`
	SourceBackedHarnessCount    int                    `json:"sourceBackedHarnessCount"`
	MetadataOnlyHarnessCount    int                    `json:"metadataOnlyHarnessCount"`
	OperatorDefinedHarnessCount int                    `json:"operatorDefinedHarnessCount"`
	SourceBackedToolCount       int                    `json:"sourceBackedToolCount"`
	PrimaryHarness              string                 `json:"primaryHarness,omitempty"`
	AvailableTools              []controlplane.Tool    `json:"availableTools"`
	InstalledHarnesses          []harnesses.Definition `json:"installedHarnesses"`
}

type SkillSummary struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Folder string `json:"folder"`
}

type ImportedInstructionsSummary struct {
	Path       string `json:"path"`
	Available  bool   `json:"available"`
	ModifiedAt string `json:"modifiedAt,omitempty"`
	Size       int64  `json:"size,omitempty"`
}

type ImportedSessionMemory struct {
	ID                string         `json:"id"`
	ImportedSessionID string         `json:"importedSessionId"`
	Kind              string         `json:"kind"`
	Content           string         `json:"content"`
	Tags              []string       `json:"tags"`
	Source            string         `json:"source"`
	Metadata          map[string]any `json:"metadata"`
	CreatedAt         int64          `json:"createdAt"`
}

type ImportedSessionRecord struct {
	ID                string                  `json:"id"`
	SourceTool        string                  `json:"sourceTool"`
	SourcePath        string                  `json:"sourcePath"`
	ExternalSessionID *string                 `json:"externalSessionId"`
	Title             *string                 `json:"title"`
	SessionFormat     string                  `json:"sessionFormat"`
	Transcript        string                  `json:"transcript"`
	Excerpt           *string                 `json:"excerpt"`
	WorkingDirectory  *string                 `json:"workingDirectory"`
	TranscriptHash    string                  `json:"transcriptHash"`
	NormalizedSession map[string]any          `json:"normalizedSession"`
	Metadata          map[string]any          `json:"metadata"`
	DiscoveredAt      int64                   `json:"discoveredAt"`
	ImportedAt        int64                   `json:"importedAt"`
	LastModifiedAt    *int64                  `json:"lastModifiedAt"`
	CreatedAt         int64                   `json:"createdAt"`
	UpdatedAt         int64                   `json:"updatedAt"`
	ParsedMemories    []ImportedSessionMemory `json:"parsedMemories"`
}

type ImportedSessionMaintenanceStats struct {
	TotalSessions                int `json:"totalSessions"`
	InlineTranscriptCount        int `json:"inlineTranscriptCount"`
	ArchivedTranscriptCount      int `json:"archivedTranscriptCount"`
	MissingRetentionSummaryCount int `json:"missingRetentionSummaryCount"`
}

type importedSessionArchiveSidecar struct {
	SessionID               string         `json:"sessionId"`
	SourceTool              string         `json:"sourceTool"`
	SourcePath              string         `json:"sourcePath"`
	SessionFormat           string         `json:"sessionFormat"`
	TranscriptHash          string         `json:"transcriptHash"`
	Title                   *string        `json:"title"`
	WorkingDirectory        *string        `json:"workingDirectory"`
	TranscriptLength        int            `json:"transcriptLength"`
	Excerpt                 *string        `json:"excerpt"`
	DurableMemoryCount      int            `json:"durableMemoryCount"`
	DurableInstructionCount int            `json:"durableInstructionCount"`
	MemoryTags              []string       `json:"memoryTags"`
	RetentionSummary        map[string]any `json:"retentionSummary"`
	ArchivedAt              int64          `json:"archivedAt"`
}

type localAuditFilter struct {
	level   string
	agentID string
	action  string
	limit   int
}

type ImportRootsSummary struct {
	Count         int                        `json:"count"`
	ExistingCount int                        `json:"existingCount"`
	Roots         []sessionimport.RootStatus `json:"roots"`
}

type ImportSourcesSummary struct {
	Count              int                       `json:"count"`
	ValidCount         int                       `json:"validCount"`
	InvalidCount       int                       `json:"invalidCount"`
	TotalEstimatedSize int64                     `json:"totalEstimatedSize"`
	Candidates         []sessionimport.Candidate `json:"candidates"`
	BySourceTool       []SummaryBucket           `json:"bySourceTool,omitempty"`
	BySourceType       []SummaryBucket           `json:"bySourceType,omitempty"`
	ByFormat           []SummaryBucket           `json:"byFormat,omitempty"`
	ByModelHint        []SummaryBucket           `json:"byModelHint,omitempty"`
	ByError            []SummaryBucket           `json:"byError,omitempty"`
}

type APIIndex struct {
	Service string      `json:"service"`
	BaseURL string      `json:"baseUrl"`
	Routes  []RouteInfo `json:"routes"`
}

type RouteInfo struct {
	Path        string `json:"path"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

type SummaryBucket struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
}

func New(cfg config.Config, detector controlplane.ToolProvider) *Server {
	memoryManager := memorystore.NewManager(filepath.Join(cfg.ConfigDir, "memory.json"))
	codeExecutor := codeexec.NewSandbox(filepath.Join(cfg.WorkspaceRoot, ".borg", "sandbox"))
	server := &Server{
		cfg:           cfg,
		memoryManager: memoryManager,
		codeExecutor:  codeExecutor,
		detector:      detector,
		mesh:          mesh.New(cfg),
		startedAt:     time.Now().UTC(),
		mux:           http.NewServeMux(),
		lifecycleModes: map[string]any{
			"lazySessionMode":        false,
			"singleActiveServerMode": false,
		},
		fallbackBuffer:    newProviderFallbackBuffer(50),
		autoDev:           newLocalAutoDevManager(cfg.WorkspaceRoot),
		squad:             newLocalSquadManager(cfg.WorkspaceRoot),
		swarm:             newLocalSwarmManager(cfg.WorkspaceRoot),
		debateHistory:     orchestration.NewDebateHistoryStore(filepath.Join(cfg.WorkspaceRoot, "debate_history.db")),
		darwinState:       newLocalDarwinStateManager(filepath.Join(cfg.WorkspaceRoot, "darwin_state.json")),
		runtimeServers:    newRuntimeServerRegistry(),
		supervisorManager: supervisor.NewManager(supervisor.ManagerOptions{WorktreeRoot: cfg.WorkspaceRoot, PersistencePath: filepath.Join(cfg.ConfigDir, "session-supervisor.json")}),
		sessionState:      newLocalSessionStateManager(filepath.Join(cfg.WorkspaceRoot, ".borg-session.json")),
		workflowEngine:    workflow.NewEngine(),
		toolsRegistry:     tools.NewRegistry(),
		mcpAggregator:     mcp.NewAggregator(),
		a2aLogger:         orchestration.NewA2ALogger(cfg.WorkspaceRoot),
	}
	server.a2aBroker = orchestration.NewA2ABroker(server.a2aLogger)
	server.skillStore = harnesses.NewSkillStore(cfg.MainConfigDir)
	server.skillRegistry = skillregistry.NewSkillRegistry()
	server.skillDecision = skillregistry.NewSkillDecisionSystem(skillregistry.DefaultSkillDecisionConfig(), server.skillRegistry)
	server.pairOrchestrator = orchestration.NewPairOrchestrator()
	server.pairOrchestrator.SetupFrontierSquad()
	server.directorNotes = orchestration.NewDirectorNotesManager()
	server.expertManager = hsync.NewExpertManager(server.goDirector, server.mcpPredictor)

	// Populate skill registry from store
	if skillIDs, err := server.skillStore.ListSkills(); err == nil {
		for _, id := range skillIDs {
			if s, err := server.skillStore.GetSkill(id); err == nil {
				server.skillRegistry.Register(skillregistry.SkillInfo{
					ID:          s.ID,
					Name:        s.Name,
					Description: s.Description,
					Content:     s.Content,
					Path:        s.Path,
				})
			}
		}
	}

	server.coderAgent = orchestration.NewCoderAgent(server.a2aBroker, cfg.WorkspaceRoot)
	server.coderAgent.Start(context.Background())
	server.goDirector = orchestration.NewDirector(server.swarmController, server.coderAgent, server.a2aBroker)
	server.mcpConfig = mcp.NewConfigManager(cfg.MainConfigDir)
	server.memoryReactor = memorystore.NewMemoryReactor(cfg.WorkspaceRoot)
	server.highValueIngestor = hsync.NewHighValueIngestor(filepath.Join(cfg.MainConfigDir, "metamcp.db"), server.skillStore, server.mcpConfig)
	server.memoryArchiver = memorystore.NewMemoryArchiver(cfg.WorkspaceRoot)
	server.swarmController = orchestration.NewSwarmController(server.a2aBroker)
	server.mcpPredictor = mcp.NewToolPredictor(server.mcpAggregator)
	server.supervisorManager.SetPredictor(server.mcpPredictor)

	// --- Initialize MCP Decision System ---
	decisionCfg := mcp.DefaultDecisionConfig()
	decisionCfg.CatalogDBPath = filepath.Join(cfg.ConfigDir, "mcp-catalog.json")
	server.mcpDecision = mcp.NewDecisionSystem(decisionCfg, server.mcpAggregator)
	server.mcpDecision.AddCatalogEntries(mcp.BuiltinTools())
	// Load persisted catalog if available
	_ = server.mcpDecision.LoadCatalog(decisionCfg.CatalogDBPath)
	// Refresh from live inventory
	if inv, err := mcp.LoadInventory(cfg.WorkspaceRoot, cfg.MainConfigDir); err == nil {
		server.mcpDecision.RefreshFromInventory(inv)
		// Initialize Go-native MCP Router
		server.nativeRouter = mcp.NewNativeMCPRouter(server.mcpDecision, nil, mcp.DefaultRouterConfig())
		if inv != nil {
			server.nativeRouter.RefreshCatalog(inv)
		}
	}

	// --- Initialize new Go-native services ---
	server.eventBus = eventbus.New(1000)
	server.eventBus.OnGlobal(func(ev eventbus.SystemEvent) {
		if data, err := json.Marshal(ev); err == nil {
			GlobalSSEBroker.Broadcast(data)
		}
	})
	server.metricsService = metrics.NewMetricsService()
	server.sessionManager = session.NewSessionManager(100)
	server.toolRegistry = toolregistry.NewToolRegistry()
	server.gitService = gitservice.NewGitService(cfg.WorkspaceRoot)
	server.contextHarvester = ctxharvester.NewContextHarvester(nil)
	server.workspaceTracker = workspaces.NewWorkspaceTracker("")
	server.processManager = processmanager.NewProcessManager()
	server.healerService = healer.NewHealerService(nil, "") // LLM provider wired later
	server.cacheService = cache.New(cache.CacheOptions{MaxSize: 500, DefaultTTL: 60000})
	server.repoGraph = repograph.NewRepoGraphService(cfg.WorkspaceRoot)

	// Register workspace on startup
	_ = server.workspaceTracker.RegisterWorkspace(cfg.WorkspaceRoot)

	server.squad.load()
	server.swarm.load()
	server.registerRoutes()
	return server
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	httpServer := &http.Server{
		Addr:              s.cfg.Host + ":" + jsonNumber(s.cfg.Port),
		Handler:           s.mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/api/memory/list", s.handleMemoryList)
	s.mux.HandleFunc("/api/memory/add", s.handleMemoryAdd)
	s.mux.HandleFunc("/api/memory/add-history", s.handleMemoryAddHistory)
	s.mux.HandleFunc("/api/code/exec", s.handleCodeExec)

	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/version", s.handleVersion)
	s.mux.HandleFunc("/api/index", s.handleAPIIndex)
	s.mux.HandleFunc("/api/health", s.handleHealth)
	s.mux.HandleFunc("/api/health/server", s.handleHealth)
	s.mux.HandleFunc("/api/config/status", s.handleConfigStatus)
	s.mux.HandleFunc("/api/config/list", s.handleConfigList)
	s.mux.HandleFunc("/api/config/get", s.handleConfigGet)
	s.mux.HandleFunc("/api/config/upsert", s.handleConfigUpsert)
	s.mux.HandleFunc("/api/config/delete", s.handleConfigDelete)
	s.mux.HandleFunc("/api/config/update", s.handleConfigUpdate)
	s.mux.HandleFunc("/api/config/mcp-timeout", s.handleConfigGetMCPTimeout)
	s.mux.HandleFunc("/api/config/mcp-timeout/set", s.handleConfigSetMCPTimeout)
	s.mux.HandleFunc("/api/config/mcp-max-attempts", s.handleConfigGetMCPMaxAttempts)
	s.mux.HandleFunc("/api/config/mcp-max-attempts/set", s.handleConfigSetMCPMaxAttempts)
	s.mux.HandleFunc("/api/config/mcp-max-total-timeout", s.handleConfigGetMCPMaxTotalTimeout)
	s.mux.HandleFunc("/api/config/mcp-max-total-timeout/set", s.handleConfigSetMCPMaxTotalTimeout)
	s.mux.HandleFunc("/api/config/mcp-reset-timeout-on-progress", s.handleConfigGetMCPResetTimeoutOnProgress)
	s.mux.HandleFunc("/api/config/mcp-reset-timeout-on-progress/set", s.handleConfigSetMCPResetTimeoutOnProgress)
	s.mux.HandleFunc("/api/config/session-lifetime", s.handleConfigGetSessionLifetime)
	s.mux.HandleFunc("/api/config/session-lifetime/set", s.handleConfigSetSessionLifetime)
	s.mux.HandleFunc("/api/config/signup-disabled", s.handleConfigGetSignupDisabled)
	s.mux.HandleFunc("/api/config/signup-disabled/set", s.handleConfigSetSignupDisabled)
	s.mux.HandleFunc("/api/config/sso-signup-disabled", s.handleConfigGetSSOSignupDisabled)
	s.mux.HandleFunc("/api/config/sso-signup-disabled/set", s.handleConfigSetSSOSignupDisabled)
	s.mux.HandleFunc("/api/config/basic-auth-disabled", s.handleConfigGetBasicAuthDisabled)
	s.mux.HandleFunc("/api/config/basic-auth-disabled/set", s.handleConfigSetBasicAuthDisabled)
	s.mux.HandleFunc("/api/config/auth-providers", s.handleConfigGetAuthProviders)
	s.mux.HandleFunc("/api/config/always-visible-tools", s.handleConfigGetAlwaysVisibleTools)
	s.mux.HandleFunc("/api/config/always-visible-tools/set", s.handleConfigSetAlwaysVisibleTools)
	s.mux.HandleFunc("/api/providers/status", s.handleProviderStatus)
	s.mux.HandleFunc("/api/providers/catalog", s.handleProviderCatalog)
	s.mux.HandleFunc("/api/providers/summary", s.handleProviderSummary)
	s.mux.HandleFunc("/api/providers/routing-summary", s.handleRoutingSummary)
	s.mux.HandleFunc("/api/sessions", s.handleSessions)
	s.mux.HandleFunc("/api/sessions/summary", s.handleSessionSummary)
	s.mux.HandleFunc("/api/sessions/context", s.handleSessionContext)
	s.mux.HandleFunc("/api/sessions/supervisor/catalog", s.handleSupervisorSessionCatalog)
	s.mux.HandleFunc("/api/sessions/supervisor/list", s.handleSupervisorSessionList)
	s.mux.HandleFunc("/api/sessions/supervisor/get", s.handleSupervisorSessionGet)
	s.mux.HandleFunc("/api/sessions/supervisor/create", s.handleSupervisorSessionCreate)
	s.mux.HandleFunc("/api/sessions/supervisor/start", s.handleSupervisorSessionStart)
	s.mux.HandleFunc("/api/sessions/supervisor/stop", s.handleSupervisorSessionStop)
	s.mux.HandleFunc("/api/sessions/supervisor/restart", s.handleSupervisorSessionRestart)
	s.mux.HandleFunc("/api/sessions/supervisor/logs", s.handleSupervisorSessionLogs)
	s.mux.HandleFunc("/api/sessions/supervisor/execute-shell", s.handleSupervisorSessionExecuteShell)
	s.mux.HandleFunc("/api/sessions/supervisor/attach-info", s.handleSupervisorSessionAttachInfo)
	s.mux.HandleFunc("/api/sessions/supervisor/health", s.handleSupervisorSessionHealth)
	s.mux.HandleFunc("/api/sessions/supervisor/state", s.handleSupervisorSessionState)
	s.mux.HandleFunc("/api/sessions/supervisor/update-state", s.handleSupervisorSessionUpdateState)
	s.mux.HandleFunc("/api/sessions/supervisor/clear", s.handleSupervisorSessionClear)
	s.mux.HandleFunc("/api/sessions/supervisor/heartbeat", s.handleSupervisorSessionHeartbeat)
	s.mux.HandleFunc("/api/sessions/supervisor/restore", s.handleSupervisorSessionRestore)
	s.mux.HandleFunc("/api/sessions/imported/list", s.handleImportedSessionList)
	s.mux.HandleFunc("/api/sessions/imported/get", s.handleImportedSessionGet)
	s.mux.HandleFunc("/api/sessions/imported/scan", s.handleImportedSessionScan)
	s.mux.HandleFunc("/api/sessions/imported/instruction-docs", s.handleImportedSessionInstructionDocs)
	s.mux.HandleFunc("/api/sessions/imported/maintenance-stats", s.handleImportedSessionMaintenanceStats)
	s.mux.HandleFunc("/api/billing/status", s.handleBillingStatus)
	s.mux.HandleFunc("/api/billing/provider-quotas", s.handleBillingProviderQuotas)
	s.mux.HandleFunc("/api/billing/cost-history", s.handleBillingCostHistory)
	s.mux.HandleFunc("/api/billing/model-pricing", s.handleBillingModelPricing)
	s.mux.HandleFunc("/api/billing/fallback-chain", s.handleBillingFallbackChain)
	s.mux.HandleFunc("/api/billing/task-routing-rules", s.handleBillingTaskRoutingRules)
	s.mux.HandleFunc("/api/billing/routing-strategy", s.handleBillingSetRoutingStrategy)
	s.mux.HandleFunc("/api/billing/task-routing-rule", s.handleBillingSetTaskRoutingRule)
	s.mux.HandleFunc("/api/billing/depleted-models", s.handleBillingDepletedModels)
	s.mux.HandleFunc("/api/billing/fallback-history", s.handleBillingFallbackHistory)
	s.mux.HandleFunc("/api/billing/fallback-history/clear", s.handleBillingClearFallbackHistory)
	s.mux.HandleFunc("/api/mcp/status", s.handleMCPStatus)
	s.mux.HandleFunc("/api/system/overview", s.handleSystemOverview)
	s.mux.HandleFunc("/api/mcp/servers/runtime", s.handleMCPRuntimeServers)
	s.mux.HandleFunc("/api/mcp/servers", s.handleMCPServersList)
	s.mux.HandleFunc("/api/mcp/servers/configured", s.handleMCPConfiguredServers)
	s.mux.HandleFunc("/api/mcp/servers/get", s.handleMCPConfiguredServerGet)
	s.mux.HandleFunc("/api/mcp/servers/create", s.handleMCPConfiguredServerCreate)
	s.mux.HandleFunc("/api/mcp/servers/update", s.handleMCPConfiguredServerUpdate)
	s.mux.HandleFunc("/api/mcp/servers/delete", s.handleMCPConfiguredServerDelete)
	s.mux.HandleFunc("/api/mcp/servers/bulk-import", s.handleMCPConfiguredServerBulkImport)
	s.mux.HandleFunc("/api/mcp/servers/reload-metadata", s.handleMCPConfiguredServerReloadMetadata)
	s.mux.HandleFunc("/api/mcp/servers/clear-metadata-cache", s.handleMCPConfiguredServerClearMetadataCache)
	s.mux.HandleFunc("/api/mcp/servers/registry-snapshot", s.handleMCPRegistrySnapshot)
	s.mux.HandleFunc("/api/mcp/servers/sync-targets", s.handleMCPSyncTargets)
	s.mux.HandleFunc("/api/mcp/servers/export-client-config", s.handleMCPExportClientConfig)
	s.mux.HandleFunc("/api/mcp/servers/sync-client-config", s.handleMCPSyncClientConfig)
	s.mux.HandleFunc("/api/mcp/tools", s.handleMCPTools)
	s.mux.HandleFunc("/api/mcp/tools/search", s.handleMCPSearchTools)
	s.mux.HandleFunc("/api/mcp/tools/predict", s.handleMCPPredictTools)
	s.mux.HandleFunc("/api/mcp/tools/call", s.handleMCPCallTool)
	s.mux.HandleFunc("/api/mcp/tools/auto-call", s.handleMCPAutoCallTool)
	s.mux.HandleFunc("/api/mcp/tool-ads", s.handleMCPToolAdvertisements)
	s.mux.HandleFunc("/api/mcp/sync", s.handleMCPSync)
	s.mux.HandleFunc("/api/service/connectivity", s.handleServiceConnectivity)
	s.mux.HandleFunc("/api/mcp/client-sync", s.handleMCPClientSync)

	// --- MCP Decision System (unified search/call/load) ---
	s.mux.HandleFunc("/api/mcp/decision/search", s.handleDecisionSearch)
	s.mux.HandleFunc("/api/mcp/decision/search-and-call", s.handleDecisionSearchAndCall)
	s.mux.HandleFunc("/api/mcp/native/search", s.handleNativeRouterSearch)
	s.mux.HandleFunc("/api/mcp/native/working-set", s.handleNativeRouterWorkingSet)
	s.mux.HandleFunc("/api/mcp/native/load", s.handleNativeRouterLoad)
	s.mux.HandleFunc("/api/mcp/native/unload", s.handleNativeRouterUnload)
	s.mux.HandleFunc("/api/mcp/native/state", s.handleNativeRouterState)
	s.mux.HandleFunc("/api/mcp/native/refresh-catalog", s.handleNativeRouterRefreshCatalog)
	s.mux.HandleFunc("/api/mcp/decision/load", s.handleDecisionLoad)
	s.mux.HandleFunc("/api/mcp/decision/call", s.handleDecisionCall)
	s.mux.HandleFunc("/api/mcp/decision/list-loaded", s.handleDecisionListLoaded)
	s.mux.HandleFunc("/api/mcp/decision/unload", s.handleDecisionUnload)
	s.mux.HandleFunc("/api/mcp/decision/list-all", s.handleDecisionListAll)
	s.mux.HandleFunc("/api/mcp/decision/events", s.handleDecisionEvents)
	s.mux.HandleFunc("/api/mcp/decision/catalog/refresh", s.handleDecisionCatalogRefresh)
	s.mux.HandleFunc("/api/mcp/decision/catalog/save", s.handleDecisionCatalogSave)

	// --- Go-native service endpoints ---
	s.mux.HandleFunc("/api/native/eventbus/publish", s.handleEventBusPublish)
	s.mux.HandleFunc("/api/native/eventbus/history", s.handleEventBusHistory)
	s.mux.HandleFunc("/api/native/cache/get", s.handleCacheGet)
	s.mux.HandleFunc("/api/native/cache/set", s.handleCacheSet)
	s.mux.HandleFunc("/api/native/cache/invalidate", s.handleCacheInvalidate)
	s.mux.HandleFunc("/api/native/cache/stats", s.handleCacheStats)
	s.mux.HandleFunc("/api/native/git/log", s.handleNativeGitLog)
	s.mux.HandleFunc("/api/native/git/status", s.handleNativeGitStatus)
	s.mux.HandleFunc("/api/native/git/diff", s.handleNativeGitDiff)
	s.mux.HandleFunc("/api/native/git/branches", s.handleNativeGitBranches)
	s.mux.HandleFunc("/api/native/session/list", s.handleSessionList)
	s.mux.HandleFunc("/api/native/session/create", s.handleSessionCreate)
	s.mux.HandleFunc("/api/native/session/get", s.handleSessionGet)
	s.mux.HandleFunc("/api/native/workspaces/list", s.handleWorkspacesList)
	s.mux.HandleFunc("/api/native/workspaces/register", s.handleWorkspacesRegister)
	s.mux.HandleFunc("/api/native/metrics/prometheus", s.handleMetricsPrometheus)
	s.mux.HandleFunc("/api/native/metrics/counters", s.handleMetricsCounters)
	s.mux.HandleFunc("/api/native/tools/search", s.handleNativeToolSearch)
	s.mux.HandleFunc("/api/native/tools/list", s.handleNativeToolList)
	s.mux.HandleFunc("/api/native/tools/register", s.handleNativeToolRegister)
	s.mux.HandleFunc("/api/native/healer/diagnose", s.handleNativeHealerDiagnose)
	s.mux.HandleFunc("/api/native/healer/history", s.handleNativeHealerHistory)
	s.mux.HandleFunc("/api/native/harvester/add", s.handleHarvesterAdd)
	s.mux.HandleFunc("/api/native/harvester/search", s.handleHarvesterSearch)
	s.mux.HandleFunc("/api/native/harvester/report", s.handleHarvesterReport)
	s.mux.HandleFunc("/api/native/process/spawn", s.handleProcessSpawn)
	s.mux.HandleFunc("/api/native/process/list", s.handleProcessList)
	s.mux.HandleFunc("/api/native/process/kill", s.handleProcessKill)
	s.mux.HandleFunc("/api/native/memory/get", s.handleGetMemory)
	s.mux.HandleFunc("/api/native/codeexec/execute", s.handleExecuteCode)
	s.mux.HandleFunc("/api/mcp/tools/schema", s.handleMCPToolSchema)
	s.mux.HandleFunc("/api/mcp/preferences", s.handleMCPToolPreferences)
	s.mux.HandleFunc("/api/mcp/traffic", s.handleMCPTraffic)
	s.mux.HandleFunc("/api/mcp/tool-selection-telemetry", s.handleMCPToolSelectionTelemetry)
	s.mux.HandleFunc("/api/mcp/tool-selection-telemetry/clear", s.handleMCPClearToolSelectionTelemetry)
	s.mux.HandleFunc("/api/mcp/server-test", s.handleMCPServerTest)
	s.mux.HandleFunc("/api/mcp/lifecycle-modes", s.handleMCPSetLifecycleModes)
	s.mux.HandleFunc("/api/mcp/runtime-servers/add", s.handleMCPAddServer)
	s.mux.HandleFunc("/api/mcp/runtime-servers/remove", s.handleMCPRemoveServer)
	s.mux.HandleFunc("/api/mcp/config/jsonc", s.handleMCPJsoncConfig)
	s.mux.HandleFunc("/api/mcp/working-set", s.handleMCPWorkingSet)
	s.mux.HandleFunc("/api/mcp/working-set/evictions", s.handleMCPWorkingSetEvictions)
	s.mux.HandleFunc("/api/mcp/working-set/evictions/clear", s.handleMCPClearWorkingSetEvictions)
	s.mux.HandleFunc("/api/mcp/working-set/load", s.handleMCPLoadTool)
	s.mux.HandleFunc("/api/mcp/working-set/unload", s.handleMCPUnloadTool)
	s.mux.HandleFunc("/api/memory/search", s.handleMemorySearch)
	s.mux.HandleFunc("/api/memory/contexts", s.handleMemoryContexts)
	s.mux.HandleFunc("/api/memory/context/save", s.handleMemoryContextSave)
	s.mux.HandleFunc("/api/memory/context/get", s.handleMemoryContextGet)
	s.mux.HandleFunc("/api/memory/context/delete", s.handleMemoryContextDelete)
	s.mux.HandleFunc("/api/memory/agent-stats", s.handleMemoryAgentStats)
	s.mux.HandleFunc("/api/memory/agent-search", s.handleMemoryAgentSearch)
	s.mux.HandleFunc("/api/memory/facts/add", s.handleMemoryAddFact)
	s.mux.HandleFunc("/api/memory/observations/record", s.handleMemoryRecordObservation)
	s.mux.HandleFunc("/api/memory/observations/recent", s.handleMemoryRecentObservations)
	s.mux.HandleFunc("/api/memory/observations/search", s.handleMemorySearchObservations)
	s.mux.HandleFunc("/api/memory/user-prompts/capture", s.handleMemoryCaptureUserPrompt)
	s.mux.HandleFunc("/api/memory/user-prompts/recent", s.handleMemoryRecentUserPrompts)
	s.mux.HandleFunc("/api/memory/user-prompts/search", s.handleMemorySearchUserPrompts)
	s.mux.HandleFunc("/api/memory/pivot/search", s.handleMemorySearchPivot)
	s.mux.HandleFunc("/api/memory/timeline/window", s.handleMemoryTimelineWindow)
	s.mux.HandleFunc("/api/memory/cross-session-links", s.handleMemoryCrossSessionLinks)
	s.mux.HandleFunc("/api/memory/session-bootstrap", s.handleMemorySessionBootstrap)
	s.mux.HandleFunc("/api/memory/tool-context", s.handleMemoryToolContext)
	s.mux.HandleFunc("/api/memory/session-summaries/capture", s.handleMemoryCaptureSessionSummary)
	s.mux.HandleFunc("/api/memory/session-summaries/recent", s.handleMemoryRecentSessionSummaries)
	s.mux.HandleFunc("/api/memory/session-summaries/search", s.handleMemorySearchSessionSummaries)
	s.mux.HandleFunc("/api/memory/sectioned-status", s.handleMemorySectionedStatus)
	s.mux.HandleFunc("/api/memory/interchange-formats", s.handleMemoryInterchangeFormats)
	s.mux.HandleFunc("/api/memory/export", s.handleMemoryExport)
	s.mux.HandleFunc("/api/memory/import", s.handleMemoryImport)
	s.mux.HandleFunc("/api/memory/convert", s.handleMemoryConvert)
	s.mux.HandleFunc("/api/agent-memory/search", s.handleAgentMemorySearch)
	s.mux.HandleFunc("/api/agent-memory/add", s.handleAgentMemoryAdd)
	s.mux.HandleFunc("/api/agent-memory/recent", s.handleAgentMemoryRecent)
	s.mux.HandleFunc("/api/agent-memory/by-type", s.handleAgentMemoryByType)
	s.mux.HandleFunc("/api/agent-memory/by-namespace", s.handleAgentMemoryByNamespace)
	s.mux.HandleFunc("/api/agent-memory/delete", s.handleAgentMemoryDelete)
	s.mux.HandleFunc("/api/agent-memory/clear-session", s.handleAgentMemoryClearSession)
	s.mux.HandleFunc("/api/agent-memory/export", s.handleAgentMemoryExport)
	s.mux.HandleFunc("/api/agent-memory/handoff", s.handleAgentMemoryHandoff)
	s.mux.HandleFunc("/api/agent-memory/pickup", s.handleAgentMemoryPickup)
	s.mux.HandleFunc("/api/agent-memory/stats", s.handleAgentMemoryStats)
	s.mux.HandleFunc("/api/graph", s.handleGraphGet)
	s.mux.HandleFunc("/api/graph/rebuild", s.handleGraphRebuild)
	s.mux.HandleFunc("/api/graph/consumers", s.handleGraphConsumers)
	s.mux.HandleFunc("/api/graph/dependencies", s.handleGraphDependencies)
	s.mux.HandleFunc("/api/graph/symbols", s.handleGraphSymbols)
	s.mux.HandleFunc("/api/context/list", s.handleContextList)
	s.mux.HandleFunc("/api/context/add", s.handleContextAdd)
	s.mux.HandleFunc("/api/context/remove", s.handleContextRemove)
	s.mux.HandleFunc("/api/context/clear", s.handleContextClear)
	s.mux.HandleFunc("/api/context/prompt", s.handleContextPrompt)
	s.mux.HandleFunc("/api/git/modules", s.handleGitModules)
	s.mux.HandleFunc("/api/git/log", s.handleGitLog)
	s.mux.HandleFunc("/api/git/status", s.handleGitStatus)
	s.mux.HandleFunc("/api/git/revert", s.handleGitRevert)
	s.mux.HandleFunc("/api/tests/status", s.handleTestsStatus)
	s.mux.HandleFunc("/api/tests/start", s.handleTestsStart)
	s.mux.HandleFunc("/api/tests/stop", s.handleTestsStop)
	s.mux.HandleFunc("/api/tests/run", s.handleTestsRun)
	s.mux.HandleFunc("/api/tests/results", s.handleTestsResults)
	s.mux.HandleFunc("/api/autodev/start-loop", s.handleAutoDevStartLoop)
	s.mux.HandleFunc("/api/autodev/cancel-loop", s.handleAutoDevCancelLoop)
	s.mux.HandleFunc("/api/autodev/loops", s.handleAutoDevGetLoops)
	s.mux.HandleFunc("/api/autodev/loop", s.handleAutoDevGetLoop)
	s.mux.HandleFunc("/api/autodev/clear-completed", s.handleAutoDevClearCompleted)
	s.mux.HandleFunc("/api/darwin/evolve", s.handleDarwinEvolve)
	s.mux.HandleFunc("/api/darwin/experiment", s.handleDarwinExperiment)
	s.mux.HandleFunc("/api/darwin/status", s.handleDarwinStatus)
	s.mux.HandleFunc("/api/squad", s.handleSquadList)
	s.mux.HandleFunc("/api/squad/spawn", s.handleSquadSpawn)
	s.mux.HandleFunc("/api/squad/kill", s.handleSquadKill)
	s.mux.HandleFunc("/api/squad/chat", s.handleSquadChat)
	s.mux.HandleFunc("/api/squad/indexer/toggle", s.handleSquadToggleIndexer)
	s.mux.HandleFunc("/api/squad/indexer/status", s.handleSquadIndexerStatus)
	s.mux.HandleFunc("/api/supervisor/decompose", s.handleSupervisorDecompose)
	s.mux.HandleFunc("/api/supervisor/supervise", s.handleSupervisorSupervise)
	s.mux.HandleFunc("/api/supervisor/status", s.handleSupervisorStatus)
	s.mux.HandleFunc("/api/supervisor/tasks", s.handleSupervisorListTasks)
	s.mux.HandleFunc("/api/supervisor/cancel", s.handleSupervisorCancel)
	s.mux.HandleFunc("/api/autonomy/get-level", s.handleAutonomyGetLevel)
	s.mux.HandleFunc("/api/autonomy/set-level", s.handleAutonomySetLevel)
	s.mux.HandleFunc("/api/autonomy/activate-full", s.handleAutonomyActivateFull)
	s.mux.HandleFunc("/api/director/memorize", s.handleDirectorMemorize)
	s.mux.HandleFunc("/api/director/chat", s.handleDirectorChat)
	s.mux.HandleFunc("/api/director/status", s.handleDirectorStatus)
	s.mux.HandleFunc("/api/director/config/update", s.handleDirectorUpdateConfig)
	s.mux.HandleFunc("/api/director-config", s.handleDirectorConfigGet)
	s.mux.HandleFunc("/api/director-config/test", s.handleDirectorConfigTest)
	s.mux.HandleFunc("/api/director-config/update", s.handleDirectorConfigUpdate)
	s.mux.HandleFunc("/api/director/auto-drive/stop", s.handleDirectorStopAutoDrive)
	s.mux.HandleFunc("/api/director/auto-drive/start", s.handleDirectorStartAutoDrive)
	s.mux.HandleFunc("/api/council/members", s.handleCouncilMembers)
	s.mux.HandleFunc("/api/council/members/update", s.handleCouncilUpdateMembers)
	s.mux.HandleFunc("/api/council/status", s.handleCouncilBaseStatus)
	s.mux.HandleFunc("/api/council/config/update", s.handleCouncilBaseUpdateConfig)
	s.mux.HandleFunc("/api/council/supervisors/add", s.handleCouncilBaseAddSupervisors)
	s.mux.HandleFunc("/api/council/supervisors/clear", s.handleCouncilBaseClearSupervisors)
	s.mux.HandleFunc("/api/council/debate", s.handleCouncilBaseDebate)
	s.mux.HandleFunc("/api/council/toggle", s.handleCouncilBaseToggle)
	s.mux.HandleFunc("/api/council/mock/add", s.handleCouncilBaseAddMock)
	s.mux.HandleFunc("/api/council/sessions", s.handleCouncilSessionsList)
	s.mux.HandleFunc("/api/council/sessions/active", s.handleCouncilSessionsActive)
	s.mux.HandleFunc("/api/council/sessions/stats", s.handleCouncilSessionsStats)
	s.mux.HandleFunc("/api/council/sessions/get", s.handleCouncilSessionsGet)
	s.mux.HandleFunc("/api/council/sessions/start", s.handleCouncilSessionsStart)
	s.mux.HandleFunc("/api/council/sessions/bulk-start", s.handleCouncilSessionsBulkStart)
	s.mux.HandleFunc("/api/council/sessions/bulk-stop", s.handleCouncilSessionsBulkStop)
	s.mux.HandleFunc("/api/council/sessions/bulk-resume", s.handleCouncilSessionsBulkResume)
	s.mux.HandleFunc("/api/council/sessions/stop", s.handleCouncilSessionsStop)
	s.mux.HandleFunc("/api/council/sessions/resume", s.handleCouncilSessionsResume)
	s.mux.HandleFunc("/api/council/sessions/delete", s.handleCouncilSessionsDelete)
	s.mux.HandleFunc("/api/council/sessions/guidance", s.handleCouncilSessionsGuidance)
	s.mux.HandleFunc("/api/council/sessions/logs", s.handleCouncilSessionsLogs)
	s.mux.HandleFunc("/api/council/sessions/templates", s.handleCouncilSessionsTemplates)
	s.mux.HandleFunc("/api/council/sessions/from-template", s.handleCouncilSessionsStartFromTemplate)
	s.mux.HandleFunc("/api/council/sessions/persisted", s.handleCouncilSessionsPersisted)
	s.mux.HandleFunc("/api/council/sessions/by-tag", s.handleCouncilSessionsByTag)
	s.mux.HandleFunc("/api/council/sessions/by-template", s.handleCouncilSessionsByTemplate)
	s.mux.HandleFunc("/api/council/sessions/by-cli", s.handleCouncilSessionsByCLI)
	s.mux.HandleFunc("/api/council/sessions/tags/update", s.handleCouncilSessionsUpdateTags)
	s.mux.HandleFunc("/api/council/sessions/tags/add", s.handleCouncilSessionsAddTag)
	s.mux.HandleFunc("/api/council/sessions/tags/remove", s.handleCouncilSessionsRemoveTag)
	s.mux.HandleFunc("/api/council/quota/status", s.handleCouncilQuotaStatus)
	s.mux.HandleFunc("/api/council/quota/config", s.handleCouncilQuotaConfig)
	s.mux.HandleFunc("/api/council/quota/enabled", s.handleCouncilQuotaEnabled)
	s.mux.HandleFunc("/api/council/quota/check", s.handleCouncilQuotaCheck)
	s.mux.HandleFunc("/api/council/quota/stats", s.handleCouncilQuotaStats)
	s.mux.HandleFunc("/api/council/quota/limits", s.handleCouncilQuotaLimits)
	s.mux.HandleFunc("/api/council/quota/reset", s.handleCouncilQuotaReset)
	s.mux.HandleFunc("/api/council/quota/unthrottle", s.handleCouncilQuotaUnthrottle)
	s.mux.HandleFunc("/api/council/quota/record-request", s.handleCouncilQuotaRecordRequest)
	s.mux.HandleFunc("/api/council/quota/rate-limit-error", s.handleCouncilQuotaRecordRateLimitError)
	s.mux.HandleFunc("/api/council/history/status", s.handleCouncilHistoryStatus)
	s.mux.HandleFunc("/api/council/history/config", s.handleCouncilHistoryConfig)
	s.mux.HandleFunc("/api/council/history/toggle", s.handleCouncilHistoryToggle)
	s.mux.HandleFunc("/api/council/history/stats", s.handleCouncilHistoryStats)
	s.mux.HandleFunc("/api/council/history/list", s.handleCouncilHistoryList)
	s.mux.HandleFunc("/api/council/history/get", s.handleCouncilHistoryGet)
	s.mux.HandleFunc("/api/council/history/delete", s.handleCouncilHistoryDelete)
	s.mux.HandleFunc("/api/council/history/supervisor", s.handleCouncilHistorySupervisor)
	s.mux.HandleFunc("/api/council/history/clear", s.handleCouncilHistoryClear)
	s.mux.HandleFunc("/api/council/history/initialize", s.handleCouncilHistoryInitialize)
	s.mux.HandleFunc("/api/council/smart-pilot/status", s.handleCouncilSmartPilotStatus)
	s.mux.HandleFunc("/api/council/smart-pilot/config", s.handleCouncilSmartPilotConfig)
	s.mux.HandleFunc("/api/council/smart-pilot/trigger", s.handleCouncilSmartPilotTrigger)
	s.mux.HandleFunc("/api/council/smart-pilot/reset-count", s.handleCouncilSmartPilotResetCount)
	s.mux.HandleFunc("/api/council/smart-pilot/reset-all", s.handleCouncilSmartPilotResetAll)
	s.mux.HandleFunc("/api/council/hooks", s.handleCouncilHooksList)
	s.mux.HandleFunc("/api/council/hooks/register", s.handleCouncilHooksRegister)
	s.mux.HandleFunc("/api/council/hooks/unregister", s.handleCouncilHooksUnregister)
	s.mux.HandleFunc("/api/council/hooks/clear", s.handleCouncilHooksClear)
	s.mux.HandleFunc("/api/council/ide/status", s.handleCouncilIDEStatus)
	s.mux.HandleFunc("/api/council/ide/submit-task", s.handleCouncilIDESubmitTask)
	s.mux.HandleFunc("/api/council/evolution/start", s.handleCouncilEvolutionStart)
	s.mux.HandleFunc("/api/council/evolution/stop", s.handleCouncilEvolutionStop)
	s.mux.HandleFunc("/api/council/evolution/optimize", s.handleCouncilEvolutionOptimize)
	s.mux.HandleFunc("/api/council/evolution/evolve", s.handleCouncilEvolutionEvolve)
	s.mux.HandleFunc("/api/council/evolution/test", s.handleCouncilEvolutionTest)
	s.mux.HandleFunc("/api/council/fine-tune/datasets", s.handleCouncilFineTuneDatasets)
	s.mux.HandleFunc("/api/council/fine-tune/datasets/get", s.handleCouncilFineTuneDatasetGet)
	s.mux.HandleFunc("/api/council/fine-tune/jobs", s.handleCouncilFineTuneJobs)
	s.mux.HandleFunc("/api/council/fine-tune/jobs/start", s.handleCouncilFineTuneJobStart)
	s.mux.HandleFunc("/api/council/fine-tune/models", s.handleCouncilFineTuneModels)
	s.mux.HandleFunc("/api/council/fine-tune/models/deploy", s.handleCouncilFineTuneModelDeploy)
	s.mux.HandleFunc("/api/council/fine-tune/chat", s.handleCouncilFineTuneChat)
	s.mux.HandleFunc("/api/council/fine-tune/stats", s.handleCouncilFineTuneStats)
	s.mux.HandleFunc("/api/council/rotation", s.handleCouncilRotationList)
	s.mux.HandleFunc("/api/council/rotation/get", s.handleCouncilRotationGet)
	s.mux.HandleFunc("/api/council/rotation/create", s.handleCouncilRotationCreate)
	s.mux.HandleFunc("/api/council/rotation/add-participant", s.handleCouncilRotationAddParticipant)
	s.mux.HandleFunc("/api/council/rotation/post-message", s.handleCouncilRotationPostMessage)
	s.mux.HandleFunc("/api/council/rotation/set-agreement", s.handleCouncilRotationSetAgreement)
	s.mux.HandleFunc("/api/council/rotation/advance-turn", s.handleCouncilRotationAdvanceTurn)
	s.mux.HandleFunc("/api/council/rotation/configure-supervisor", s.handleCouncilRotationConfigureSupervisor)
	s.mux.HandleFunc("/api/council/rotation/run-supervisor-check", s.handleCouncilRotationRunSupervisorCheck)
	s.mux.HandleFunc("/api/council/rotation/update-shared-context", s.handleCouncilRotationUpdateSharedContext)
	s.mux.HandleFunc("/api/council/rotation/pause", s.handleCouncilRotationPause)
	s.mux.HandleFunc("/api/council/rotation/resume", s.handleCouncilRotationResume)
	s.mux.HandleFunc("/api/council/rotation/start-execution", s.handleCouncilRotationStartExecution)
	s.mux.HandleFunc("/api/council/rotation/complete", s.handleCouncilRotationComplete)
	s.mux.HandleFunc("/api/council/visual/system-diagram", s.handleCouncilVisualSystemDiagram)
	s.mux.HandleFunc("/api/council/visual/plan-diagram", s.handleCouncilVisualPlanDiagram)
	s.mux.HandleFunc("/api/council/visual/parse-plan", s.handleCouncilVisualParsePlan)
	s.mux.HandleFunc("/api/deerflow/status", s.handleDeerFlowStatus)
	s.mux.HandleFunc("/api/deerflow/models", s.handleDeerFlowModels)
	s.mux.HandleFunc("/api/deerflow/skills", s.handleDeerFlowSkills)
	s.mux.HandleFunc("/api/deerflow/memory", s.handleDeerFlowMemory)
	s.mux.HandleFunc("/api/healer/diagnose", s.handleHealerDiagnose)
	s.mux.HandleFunc("/api/healer/heal", s.handleHealerHeal)
	s.mux.HandleFunc("/api/healer/history", s.handleHealerHistory)
	s.mux.HandleFunc("/api/clouddev/providers", s.handleCloudDevListProviders)
	s.mux.HandleFunc("/api/clouddev/sessions/create", s.handleCloudDevCreateSession)
	s.mux.HandleFunc("/api/clouddev/sessions", s.handleCloudDevListSessions)
	s.mux.HandleFunc("/api/clouddev/sessions/get", s.handleCloudDevGetSession)
	s.mux.HandleFunc("/api/clouddev/sessions/status", s.handleCloudDevUpdateSessionStatus)
	s.mux.HandleFunc("/api/clouddev/sessions/delete", s.handleCloudDevDeleteSession)
	s.mux.HandleFunc("/api/clouddev/messages/send", s.handleCloudDevSendMessage)
	s.mux.HandleFunc("/api/clouddev/messages/broadcast", s.handleCloudDevBroadcastMessage)
	s.mux.HandleFunc("/api/clouddev/messages/preview-recipients", s.handleCloudDevPreviewBroadcastRecipients)
	s.mux.HandleFunc("/api/clouddev/plan/accept", s.handleCloudDevAcceptPlan)
	s.mux.HandleFunc("/api/clouddev/plan/auto-accept", s.handleCloudDevSetAutoAcceptPlan)
	s.mux.HandleFunc("/api/clouddev/messages/get", s.handleCloudDevGetMessages)
	s.mux.HandleFunc("/api/clouddev/logs", s.handleCloudDevGetLogs)
	s.mux.HandleFunc("/api/clouddev/stats", s.handleCloudDevStats)
	s.mux.HandleFunc("/api/metrics/stats", s.handleMetricsStats)
	s.mux.HandleFunc("/api/metrics/track", s.handleMetricsTrack)
	s.mux.HandleFunc("/api/metrics/system-snapshot", s.handleMetricsSystemSnapshot)
	s.mux.HandleFunc("/api/metrics/timeline", s.handleMetricsTimeline)
	s.mux.HandleFunc("/api/metrics/provider-breakdown", s.handleMetricsProviderBreakdown)
	s.mux.HandleFunc("/api/metrics/monitoring", s.handleMetricsMonitoring)
	s.mux.HandleFunc("/api/metrics/routing-history", s.handleMetricsRoutingHistory)
	s.mux.HandleFunc("/api/logs", s.handleLogsList)
	s.mux.HandleFunc("/api/logs/summary", s.handleLogsSummary)
	s.mux.HandleFunc("/api/logs/clear", s.handleLogsClear)
	s.mux.HandleFunc("/api/server-health/check", s.handleServerHealthCheck)
	s.mux.HandleFunc("/api/server-health/reset", s.handleServerHealthReset)
	s.mux.HandleFunc("/api/settings", s.handleSettingsGet)
	s.mux.HandleFunc("/api/settings/update", s.handleSettingsUpdate)
	s.mux.HandleFunc("/api/settings/providers", s.handleSettingsProviders)
	s.mux.HandleFunc("/api/settings/test-connection", s.handleSettingsTestConnection)
	s.mux.HandleFunc("/api/settings/environment", s.handleSettingsEnvironment)
	s.mux.HandleFunc("/api/settings/mcp-servers", s.handleSettingsMCPServers)
	s.mux.HandleFunc("/api/settings/provider-key", s.handleSettingsProviderKey)
	s.mux.HandleFunc("/api/skills/get", s.handleSkillGet)
	s.mux.HandleFunc("/api/skills/list", s.handleSkillList)
	s.mux.HandleFunc("/api/skills/search", s.handleSkillSearch)
	s.mux.HandleFunc("/api/skills/load", s.handleSkillLoad)
	s.mux.HandleFunc("/api/skills/unload", s.handleSkillUnload)
	s.mux.HandleFunc("/api/skills/list-loaded", s.handleSkillListLoaded)
	s.mux.HandleFunc("/api/skills/summary", s.handleSkillsSummary)
	s.mux.HandleFunc("/api/tools", s.handleToolsList)
	s.mux.HandleFunc("/api/tools/by-server", s.handleToolsByServer)
	s.mux.HandleFunc("/api/tools/search", s.handleToolsSearch)
	s.mux.HandleFunc("/api/tools/context", s.handleToolsContext)
	s.mux.HandleFunc("/api/tools/detect-cli-harnesses", s.handleToolsDetectCLIHarnesses)
	s.mux.HandleFunc("/api/tools/detect-execution-environment", s.handleToolsDetectExecutionEnvironment)
	s.mux.HandleFunc("/api/tools/detect-install-surfaces", s.handleToolsDetectInstallSurfaces)
	s.mux.HandleFunc("/api/tools/get", s.handleToolsGet)
	s.mux.HandleFunc("/api/tools/create", s.handleToolsCreate)
	s.mux.HandleFunc("/api/tools/upsert-batch", s.handleToolsUpsertBatch)
	s.mux.HandleFunc("/api/tools/delete", s.handleToolsDelete)
	s.mux.HandleFunc("/api/tools/always-on", s.handleToolsAlwaysOn)
	s.mux.HandleFunc("/api/tool-sets", s.handleToolSetsList)
	s.mux.HandleFunc("/api/tool-sets/get", s.handleToolSetsGet)
	s.mux.HandleFunc("/api/tool-sets/create", s.handleToolSetsCreate)
	s.mux.HandleFunc("/api/tool-sets/update", s.handleToolSetsUpdate)
	s.mux.HandleFunc("/api/tool-sets/delete", s.handleToolSetsDelete)
	s.mux.HandleFunc("/api/project/context", s.handleProjectContext)
	s.mux.HandleFunc("/api/director/notes", s.handleDirectorNotesList)
	s.mux.HandleFunc("/api/director/notes/synthesize", s.handleDirectorNotesSynthesize)
	s.mux.HandleFunc("/api/project/context/update", s.handleProjectContextUpdate)
	s.mux.HandleFunc("/api/project/handoffs", s.handleProjectHandoffs)
	s.mux.HandleFunc("/api/shell/log", s.handleShellLog)
	s.mux.HandleFunc("/api/shell/history/query", s.handleShellQueryHistory)
	s.mux.HandleFunc("/api/shell/history/system", s.handleShellSystemHistory)
	s.mux.HandleFunc("/api/agent/tool", s.handleAgentRunTool)
	s.mux.HandleFunc("/api/agent/chat", s.handleAgentChat)
	s.mux.HandleFunc("/api/agent/a2a/agents", s.handleA2AListAgents)
	s.mux.HandleFunc("/api/agent/a2a/messages", s.handleA2AGetMessages)
	s.mux.HandleFunc("/api/agent/a2a/logs", s.handleA2AGetLogs)
	s.mux.HandleFunc("/api/agent/a2a/broadcast", s.handleA2ABroadcast)
	s.mux.HandleFunc("/api/agent/swarm/start", s.handleAgentSwarmStart)
	s.mux.HandleFunc("/api/agent/swarm/transcript", s.handleAgentSwarmTranscript)
	s.mux.HandleFunc("/api/agent/supervisor/evaluate", s.handleAgentSupervisorEvaluate)
	s.mux.HandleFunc("/api/agent/director/start", s.handleGoDirectorStart)
	s.mux.HandleFunc("/api/memory/archive-session", s.handleMemoryArchiveSession)
	s.mux.HandleFunc("/api/memory/hydrate", s.handleMemoryHydrate)
	s.mux.HandleFunc("/api/memory/hydration/status", s.handleMemoryHydrationStatus)
	s.mux.HandleFunc("/api/memory/hydration/query", s.handleMemoryHydrationQuery)
	s.mux.HandleFunc("/api/memory/hydration/add", s.handleMemoryHydrationAdd)
	s.mux.HandleFunc("/api/commands/execute", s.handleCommandsExecute)
	s.mux.HandleFunc("/api/commands", s.handleCommandsList)
	s.mux.HandleFunc("/api/code/wasm/exec", s.handleWASMExec)
	s.mux.HandleFunc("/api/code/wasm/status", s.handleWASMStatus)
	s.mux.HandleFunc("/api/skills", s.handleSkillsList)
	s.mux.HandleFunc("/api/skills/read", s.handleSkillsRead)
	s.mux.HandleFunc("/api/skills/create", s.handleSkillsCreate)
	s.mux.HandleFunc("/api/skills/save", s.handleSkillsSave)
	s.mux.HandleFunc("/api/skills/assimilate", s.handleSkillsAssimilate)
	s.mux.HandleFunc("/api/workflows", s.handleWorkflowList)
	s.mux.HandleFunc("/api/workflows/graph", s.handleWorkflowGraph)
	s.mux.HandleFunc("/api/workflows/start", s.handleWorkflowStart)
	s.mux.HandleFunc("/api/workflows/executions", s.handleWorkflowExecutions)
	s.mux.HandleFunc("/api/workflows/execution", s.handleWorkflowExecution)
	s.mux.HandleFunc("/api/workflows/history", s.handleWorkflowHistory)
	s.mux.HandleFunc("/api/workflows/resume", s.handleWorkflowResume)
	s.mux.HandleFunc("/api/workflows/pause", s.handleWorkflowPause)
	s.mux.HandleFunc("/api/workflows/approve", s.handleWorkflowApprove)
	s.mux.HandleFunc("/api/workflows/reject", s.handleWorkflowReject)
	s.mux.HandleFunc("/api/workflows/canvases", s.handleWorkflowCanvases)
	s.mux.HandleFunc("/api/workflows/canvas", s.handleWorkflowCanvas)
	s.mux.HandleFunc("/api/workflows/canvas/save", s.handleWorkflowCanvasSave)
	s.mux.HandleFunc("/api/symbols", s.handleSymbolsList)
	s.mux.HandleFunc("/api/symbols/find", s.handleSymbolsFind)
	s.mux.HandleFunc("/api/symbols/pin", s.handleSymbolsPin)
	s.mux.HandleFunc("/api/symbols/unpin", s.handleSymbolsUnpin)
	s.mux.HandleFunc("/api/symbols/priority", s.handleSymbolsUpdatePriority)
	s.mux.HandleFunc("/api/symbols/notes", s.handleSymbolsAddNotes)
	s.mux.HandleFunc("/api/symbols/clear", s.handleSymbolsClear)
	s.mux.HandleFunc("/api/symbols/file", s.handleSymbolsForFile)
	s.mux.HandleFunc("/api/lsp/find-symbol", s.handleLSPFindSymbol)
	s.mux.HandleFunc("/api/lsp/find-references", s.handleLSPFindReferences)
	s.mux.HandleFunc("/api/lsp/symbols", s.handleLSPGetSymbols)
	s.mux.HandleFunc("/api/lsp/search", s.handleLSPSearchSymbols)
	s.mux.HandleFunc("/api/lsp/index", s.handleLSPIndexProject)
	s.mux.HandleFunc("/api/api-keys", s.handleAPIKeysList)
	s.mux.HandleFunc("/api/api-keys/get", s.handleAPIKeysGet)
	s.mux.HandleFunc("/api/api-keys/create", s.handleAPIKeysCreate)
	s.mux.HandleFunc("/api/api-keys/update", s.handleAPIKeysUpdate)
	s.mux.HandleFunc("/api/api-keys/delete", s.handleAPIKeysDelete)
	s.mux.HandleFunc("/api/api-keys/validate", s.handleAPIKeysValidate)
	s.mux.HandleFunc("/api/audit", s.handleAuditList)
	s.mux.HandleFunc("/api/audit/query", s.handleAuditQuery)
	s.mux.HandleFunc("/api/scripts", s.handleSavedScriptsList)
	s.mux.HandleFunc("/api/scripts/get", s.handleSavedScriptsGet)
	s.mux.HandleFunc("/api/scripts/create", s.handleSavedScriptsCreate)
	s.mux.HandleFunc("/api/scripts/update", s.handleSavedScriptsUpdate)
	s.mux.HandleFunc("/api/scripts/delete", s.handleSavedScriptsDelete)
	s.mux.HandleFunc("/api/scripts/execute", s.handleSavedScriptsExecute)
	s.mux.HandleFunc("/api/links-backlog", s.handleLinksBacklogList)
	s.mux.HandleFunc("/api/links-backlog/stats", s.handleLinksBacklogStats)
	s.mux.HandleFunc("/api/links-backlog/get", s.handleLinksBacklogGet)
	s.mux.HandleFunc("/api/links-backlog/sync", s.handleLinksBacklogSync)
	s.mux.HandleFunc("/api/infrastructure", s.handleInfrastructureStatus)
	s.mux.HandleFunc("/api/infrastructure/doctor", s.handleInfrastructureDoctor)
	s.mux.HandleFunc("/api/infrastructure/apply", s.handleInfrastructureApply)
	s.mux.HandleFunc("/api/expert/research", s.handleExpertResearch)
	s.mux.HandleFunc("/api/expert/code", s.handleExpertCode)
	s.mux.HandleFunc("/api/expert/status", s.handleExpertStatus)
	s.mux.HandleFunc("/api/expert/predict", s.handleExpertPredict)
	s.mux.HandleFunc("/api/expert/groom", s.handleExpertGroom)
	s.mux.HandleFunc("/api/policies", s.handlePoliciesList)
	s.mux.HandleFunc("/api/policies/get", s.handlePoliciesGet)
	s.mux.HandleFunc("/api/policies/create", s.handlePoliciesCreate)
	s.mux.HandleFunc("/api/policies/update", s.handlePoliciesUpdate)
	s.mux.HandleFunc("/api/policies/delete", s.handlePoliciesDelete)
	s.mux.HandleFunc("/api/secrets", s.handleSecretsList)
	s.mux.HandleFunc("/api/secrets/set", s.handleSecretsSet)
	s.mux.HandleFunc("/api/secrets/delete", s.handleSecretsDelete)
	s.mux.HandleFunc("/api/marketplace", s.handleMarketplaceList)
	s.mux.HandleFunc("/api/marketplace/install", s.handleMarketplaceInstall)
	s.mux.HandleFunc("/api/marketplace/publish", s.handleMarketplacePublish)
	s.mux.HandleFunc("/api/catalog", s.handleCatalogList)
	s.mux.HandleFunc("/api/catalog/get", s.handleCatalogGet)
	s.mux.HandleFunc("/api/catalog/runs", s.handleCatalogRuns)
	s.mux.HandleFunc("/api/catalog/ingest", s.handleCatalogIngest)
	s.mux.HandleFunc("/api/catalog/validate", s.handleCatalogValidate)
	s.mux.HandleFunc("/api/catalog/install", s.handleCatalogInstall)
	s.mux.HandleFunc("/api/catalog/validate-batch", s.handleCatalogValidateBatch)
	s.mux.HandleFunc("/api/catalog/stats", s.handleCatalogStats)
	s.mux.HandleFunc("/api/catalog/linked-servers", s.handleCatalogLinkedServers)
	s.mux.HandleFunc("/api/oauth/clients/create", s.handleOAuthClientCreate)
	s.mux.HandleFunc("/api/oauth/clients/get", s.handleOAuthClientGet)
	s.mux.HandleFunc("/api/oauth/sessions/upsert", s.handleOAuthSessionUpsert)
	s.mux.HandleFunc("/api/oauth/sessions/by-server", s.handleOAuthSessionGetByServer)
	s.mux.HandleFunc("/api/oauth/exchange", s.handleOAuthExchange)
	s.mux.HandleFunc("/api/research/conduct", s.handleResearchConduct)
	s.mux.HandleFunc("/api/research/ingest", s.handleResearchIngest)
	s.mux.HandleFunc("/api/research/recursive", s.handleResearchRecursive)
	s.mux.HandleFunc("/api/research/queries", s.handleResearchQueries)
	s.mux.HandleFunc("/api/research/queue", s.handleResearchQueue)
	s.mux.HandleFunc("/api/research/retry-failed", s.handleResearchRetryFailed)
	s.mux.HandleFunc("/api/research/retry-all-failed", s.handleResearchRetryAllFailed)
	s.mux.HandleFunc("/api/research/enqueue", s.handleResearchEnqueuePending)
	s.mux.HandleFunc("/api/pulse/events", s.handlePulseEvents)
	s.mux.HandleFunc("/api/pulse/status", s.handlePulseStatus)
	s.mux.HandleFunc("/api/pulse/providers", s.handlePulseProviders)
	s.mux.HandleFunc("/api/session-export/export", s.handleSessionExport)
	s.mux.HandleFunc("/api/session-export/import", s.handleSessionImport)
	s.mux.HandleFunc("/api/session-export/detect-format", s.handleSessionExportDetectFormat)
	s.mux.HandleFunc("/api/session-export/formats", s.handleSessionExportKnownFormats)
	s.mux.HandleFunc("/api/session-export/history", s.handleSessionExportHistory)
	s.mux.HandleFunc("/api/browser/status", s.handleBrowserStatus)
	s.mux.HandleFunc("/api/browser/close-page", s.handleBrowserClosePage)
	s.mux.HandleFunc("/api/browser/close-all", s.handleBrowserCloseAll)
	s.mux.HandleFunc("/api/browser/search-history", s.handleBrowserSearchHistory)
	s.mux.HandleFunc("/api/browser/scrape", s.handleBrowserScrapePage)
	s.mux.HandleFunc("/api/browser/screenshot", s.handleBrowserScreenshot)
	s.mux.HandleFunc("/api/browser/debug", s.handleBrowserDebug)
	s.mux.HandleFunc("/api/browser/proxy-fetch", s.handleBrowserProxyFetch)
	s.mux.HandleFunc("/api/browser-extension/save-memory", s.handleBrowserExtensionSaveMemory)
	s.mux.HandleFunc("/api/browser-extension/parse-dom", s.handleBrowserExtensionParseDOM)
	s.mux.HandleFunc("/api/browser-extension/memories", s.handleBrowserExtensionListMemories)
	s.mux.HandleFunc("/api/browser-extension/delete-memory", s.handleBrowserExtensionDeleteMemory)
	s.mux.HandleFunc("/api/browser-extension/stats", s.handleBrowserExtensionStats)
	s.mux.HandleFunc("/api/open-webui/status", s.handleOpenWebUIStatus)
	s.mux.HandleFunc("/api/open-webui/embed-url", s.handleOpenWebUIEmbedURL)
	s.mux.HandleFunc("/api/code-mode/status", s.handleCodeModeStatus)
	s.mux.HandleFunc("/api/code-mode/enable", s.handleCodeModeEnable)
	s.mux.HandleFunc("/api/code-mode/disable", s.handleCodeModeDisable)
	s.mux.HandleFunc("/api/code-mode/execute", s.handleCodeModeExecute)
	s.mux.HandleFunc("/api/submodules", s.handleSubmoduleList)
	s.mux.HandleFunc("/api/submodules/update-all", s.handleSubmoduleUpdateAll)
	s.mux.HandleFunc("/api/submodules/install-dependencies", s.handleSubmoduleInstallDependencies)
	s.mux.HandleFunc("/api/submodules/build", s.handleSubmoduleBuild)
	s.mux.HandleFunc("/api/submodules/enable", s.handleSubmoduleEnable)
	s.mux.HandleFunc("/api/submodules/capabilities", s.handleSubmoduleCapabilities)
	s.mux.HandleFunc("/api/suggestions", s.handleSuggestionsList)
	s.mux.HandleFunc("/api/suggestions/resolve", s.handleSuggestionsResolve)
	s.mux.HandleFunc("/api/suggestions/clear", s.handleSuggestionsClear)
	s.mux.HandleFunc("/api/plan/mode", s.handlePlanMode)
	s.mux.HandleFunc("/api/plan/diffs", s.handlePlanDiffs)
	s.mux.HandleFunc("/api/plan/approve-diff", s.handlePlanApproveDiff)
	s.mux.HandleFunc("/api/plan/reject-diff", s.handlePlanRejectDiff)
	s.mux.HandleFunc("/api/plan/apply-all", s.handlePlanApplyAll)
	s.mux.HandleFunc("/api/plan/summary", s.handlePlanSummary)
	s.mux.HandleFunc("/api/plan/checkpoints", s.handlePlanCheckpoints)
	s.mux.HandleFunc("/api/plan/create-checkpoint", s.handlePlanCreateCheckpoint)
	s.mux.HandleFunc("/api/plan/rollback", s.handlePlanRollback)
	s.mux.HandleFunc("/api/plan/clear", s.handlePlanClear)
	s.mux.HandleFunc("/api/knowledge/graph", s.handleKnowledgeGraph)
	s.mux.HandleFunc("/api/knowledge/stats", s.handleKnowledgeStats)
	s.mux.HandleFunc("/api/knowledge/ingest", s.handleKnowledgeIngest)
	s.mux.HandleFunc("/api/knowledge/resources", s.handleKnowledgeResources)
	s.mux.HandleFunc("/api/rag/file", s.handleRAGIngestFile)
	s.mux.HandleFunc("/api/rag/text", s.handleRAGIngestText)
	s.mux.HandleFunc("/api/directory", s.handleUnifiedDirectoryList)
	s.mux.HandleFunc("/api/directory/stats", s.handleUnifiedDirectoryStats)
	s.mux.HandleFunc("/api/directory/high-value-ingest", s.handleHighValueIngest)
	s.mux.HandleFunc("/api/tool-chains/aliases", s.handleToolChainAliases)
	s.mux.HandleFunc("/api/tool-chains/aliases/create", s.handleToolChainCreateAlias)
	s.mux.HandleFunc("/api/tool-chains/aliases/remove", s.handleToolChainRemoveAlias)
	s.mux.HandleFunc("/api/tool-chains/aliases/resolve", s.handleToolChainResolveAlias)
	s.mux.HandleFunc("/api/tool-chains", s.handleToolChainsList)
	s.mux.HandleFunc("/api/tool-chains/get", s.handleToolChainsGet)
	s.mux.HandleFunc("/api/tool-chains/create", s.handleToolChainsCreate)
	s.mux.HandleFunc("/api/tool-chains/execute", s.handleToolChainsExecute)
	s.mux.HandleFunc("/api/tool-chains/delete", s.handleToolChainsDelete)
	s.mux.HandleFunc("/api/tool-chains/lazy", s.handleToolChainsLazyStates)
	s.mux.HandleFunc("/api/tool-chains/lazy/register", s.handleToolChainsRegisterLazy)
	s.mux.HandleFunc("/api/tool-chains/lazy/mark-loaded", s.handleToolChainsMarkLoaded)
	s.mux.HandleFunc("/api/browser-controls/scrape", s.handleBrowserControlsScrape)
	s.mux.HandleFunc("/api/browser-controls/history/push", s.handleBrowserControlsPushHistory)
	s.mux.HandleFunc("/api/browser-controls/history/query", s.handleBrowserControlsQueryHistory)
	s.mux.HandleFunc("/api/browser-controls/logs/push", s.handleBrowserControlsPushLogs)
	s.mux.HandleFunc("/api/browser-controls/logs/query", s.handleBrowserControlsQueryLogs)
	s.mux.HandleFunc("/api/browser-controls/stats", s.handleBrowserControlsStats)
	s.mux.HandleFunc("/api/agent/pair/run", s.handlePairSessionRun)
	s.mux.HandleFunc("/api/agent/pair/status", s.handlePairSessionStatus)
	s.mux.HandleFunc("/api/agent/pair/rotate", s.handlePairSessionRotate)
	s.mux.HandleFunc("/api/swarm/start", s.handleSwarmStart)
	s.mux.HandleFunc("/api/swarm/resume", s.handleSwarmResumeMission)
	s.mux.HandleFunc("/api/swarm/approve-task", s.handleSwarmApproveTask)
	s.mux.HandleFunc("/api/swarm/decompose-task", s.handleSwarmDecomposeTask)
	s.mux.HandleFunc("/api/swarm/update-task-priority", s.handleSwarmUpdateTaskPriority)
	s.mux.HandleFunc("/api/swarm/debate", s.handleSwarmExecuteDebate)
	s.mux.HandleFunc("/api/swarm/consensus", s.handleSwarmSeekConsensus)
	s.mux.HandleFunc("/api/swarm/missions", s.handleSwarmMissionHistory)
	s.mux.HandleFunc("/api/swarm/risk/summary", s.handleSwarmMissionRiskSummary)
	s.mux.HandleFunc("/api/swarm/risk/rows", s.handleSwarmMissionRiskRows)
	s.mux.HandleFunc("/api/swarm/risk/facets", s.handleSwarmMissionRiskFacets)
	s.mux.HandleFunc("/api/swarm/mesh-capabilities", s.handleSwarmMeshCapabilities)
	s.mux.HandleFunc("/api/swarm/direct-message", s.handleSwarmSendDirectMessage)
	s.mux.HandleFunc("/api/cli/tools", s.handleCLITools)
	s.mux.HandleFunc("/api/cli/harnesses", s.handleHarnesses)
	s.mux.HandleFunc("/api/cli/summary", s.handleCLISummary)
	s.mux.HandleFunc("/api/memory/borg-memory/status", s.handleMemoryStatus)
	s.mux.HandleFunc("/api/import/sources", s.handleImportSources)
	s.mux.HandleFunc("/api/import/roots", s.handleImportRoots)
	s.mux.HandleFunc("/api/import/validate", s.handleImportValidate)
	s.mux.HandleFunc("/api/import/candidates", s.handleImportCandidates)
	s.mux.HandleFunc("/api/import/manifest", s.handleImportManifest)
	s.mux.HandleFunc("/api/import/summary", s.handleImportSummary)
	s.mux.HandleFunc("/api/runtime/locks", s.handleRuntimeLocks)
	s.mux.HandleFunc("/api/runtime/status", s.handleRuntimeStatus)
	s.mux.HandleFunc("/api/runtime/imported-instructions", s.handleImportedInstructions)
	s.mux.HandleFunc("/api/startup/status", s.handleStartupStatus)
	s.mux.HandleFunc("/api/mesh/status", s.handleMeshStatus)
	s.mux.HandleFunc("/api/mesh/peers", s.handleMeshPeers)
	s.mux.HandleFunc("/api/mesh/capabilities", s.handleMeshCapabilities)
	s.mux.HandleFunc("/api/mesh/query-capabilities", s.handleMeshQueryCapabilities)
	s.mux.HandleFunc("/api/mesh/find-peer", s.handleMeshFindPeer)
	s.mux.HandleFunc("/api/mesh/broadcast", s.handleMeshBroadcast)

	// --- Repograph Routes ---
	s.mux.HandleFunc("/api/repograph/build", s.handleRepoGraphBuild)
	s.mux.HandleFunc("/api/repograph/graph", s.handleRepoGraphGet)
	s.mux.HandleFunc("/api/repograph/references", s.handleRepoGraphReferences)
	s.mux.HandleFunc("/api/repograph/dependents", s.handleRepoGraphDependents)
	s.mux.HandleFunc("/api/repograph/search", s.handleRepoGraphSearch)

	s.mux.HandleFunc("/api/sse", s.handleSSE)

	// --- New Go-native handlers (alpha.11+) ---
	s.registerSavedScriptRoutes()
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":        true,
		"service":   "borg-go",
		"version":   buildinfo.Version,
		"uptimeSec": int(time.Since(s.startedAt).Seconds()),
		"baseUrl":   s.cfg.BaseURL(),
	})
}

func (s *Server) handleVersion(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"version": buildinfo.Version,
	})
}

func (s *Server) handleAPIIndex(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": APIIndex{
			Service: "borg-go",
			BaseURL: s.cfg.BaseURL(),
			Routes: []RouteInfo{
				{Path: "/health", Category: "meta", Description: "Basic service health check."},
				{Path: "/version", Category: "meta", Description: "Build version for the Go sidecar."},
				{Path: "/api/index", Category: "meta", Description: "Self-describing index of the Go sidecar API surface."},
				{Path: "/api/health/server", Category: "meta", Description: "Health check alias for API consumers."},
				{Path: "/api/config/status", Category: "config", Description: "Path and config visibility snapshot for the sidecar and main workspace."},
				{Path: "/api/config/list", Category: "config", Description: "List config key/value entries through the TypeScript config router."},
				{Path: "/api/config/get", Category: "config", Description: "Read one config key through the TypeScript config router."},
				{Path: "/api/config/upsert", Category: "config", Description: "Upsert a config key through the TypeScript config router."},
				{Path: "/api/config/delete", Category: "config", Description: "Delete a config key through the TypeScript config router."},
				{Path: "/api/config/update", Category: "config", Description: "Update a config key through the TypeScript config router."},
				{Path: "/api/config/mcp-timeout", Category: "config", Description: "Read MCP timeout through the TypeScript config router."},
				{Path: "/api/config/mcp-timeout/set", Category: "config", Description: "Update MCP timeout through the TypeScript config router."},
				{Path: "/api/config/mcp-max-attempts", Category: "config", Description: "Read MCP max attempts through the TypeScript config router."},
				{Path: "/api/config/mcp-max-attempts/set", Category: "config", Description: "Update MCP max attempts through the TypeScript config router."},
				{Path: "/api/config/mcp-max-total-timeout", Category: "config", Description: "Read MCP max total timeout through the TypeScript config router."},
				{Path: "/api/config/mcp-max-total-timeout/set", Category: "config", Description: "Update MCP max total timeout through the TypeScript config router."},
				{Path: "/api/config/mcp-reset-timeout-on-progress", Category: "config", Description: "Read the MCP reset-on-progress flag through the TypeScript config router."},
				{Path: "/api/config/mcp-reset-timeout-on-progress/set", Category: "config", Description: "Update the MCP reset-on-progress flag through the TypeScript config router."},
				{Path: "/api/config/session-lifetime", Category: "config", Description: "Read session lifetime through the TypeScript config router."},
				{Path: "/api/config/session-lifetime/set", Category: "config", Description: "Update session lifetime through the TypeScript config router."},
				{Path: "/api/config/signup-disabled", Category: "config", Description: "Read signup-disabled state through the TypeScript config router."},
				{Path: "/api/config/signup-disabled/set", Category: "config", Description: "Update signup-disabled state through the TypeScript config router."},
				{Path: "/api/config/sso-signup-disabled", Category: "config", Description: "Read SSO-signup-disabled state through the TypeScript config router."},
				{Path: "/api/config/sso-signup-disabled/set", Category: "config", Description: "Update SSO-signup-disabled state through the TypeScript config router."},
				{Path: "/api/config/basic-auth-disabled", Category: "config", Description: "Read basic-auth-disabled state through the TypeScript config router."},
				{Path: "/api/config/basic-auth-disabled/set", Category: "config", Description: "Update basic-auth-disabled state through the TypeScript config router."},
				{Path: "/api/config/auth-providers", Category: "config", Description: "Read auth providers, with a local Go OIDC availability fallback when the TypeScript config router is unavailable."},
				{Path: "/api/config/always-visible-tools", Category: "config", Description: "Read always-visible tools, with a local JSONC fallback that mirrors TypeScript tool-selection precedence when the config router is unavailable."},
				{Path: "/api/config/always-visible-tools/set", Category: "config", Description: "Update always-visible tools through the TypeScript config router."},
				{Path: "/api/providers/status", Category: "providers", Description: "Provider credential presence and auth-method hints."},
				{Path: "/api/providers/catalog", Category: "providers", Description: "Provider catalog metadata including default models and preferred tasks."},
				{Path: "/api/providers/summary", Category: "providers", Description: "Compact provider counts and auth/preferred-task buckets."},
				{Path: "/api/providers/routing-summary", Category: "providers", Description: "Read-only preview of intended task routing order."},
				{Path: "/api/sessions", Category: "sessions", Description: "Discovered session artifacts exposed as read-only sessions."},
				{Path: "/api/sessions/summary", Category: "sessions", Description: "Compact summary of discovered sessions by tool, format, task, and model hint."},
				{Path: "/api/sessions/context", Category: "sessions", Description: "Go-owned session context summary combining startup readiness, memory bootstrap, and tool advertisements."},
				{Path: "/api/sessions/supervisor/catalog", Category: "sessions", Description: "Bridge to the TypeScript session harness catalog."},
				{Path: "/api/sessions/supervisor/list", Category: "sessions", Description: "Bridge to the TypeScript supervised session list."},
				{Path: "/api/sessions/supervisor/get", Category: "sessions", Description: "Bridge to a specific TypeScript supervised session snapshot."},
				{Path: "/api/sessions/supervisor/create", Category: "sessions", Description: "Create a supervised session through the TypeScript control plane."},
				{Path: "/api/sessions/supervisor/start", Category: "sessions", Description: "Start a supervised session through the TypeScript control plane."},
				{Path: "/api/sessions/supervisor/stop", Category: "sessions", Description: "Stop a supervised session through the TypeScript control plane."},
				{Path: "/api/sessions/supervisor/restart", Category: "sessions", Description: "Restart a supervised session through the TypeScript control plane."},
				{Path: "/api/sessions/supervisor/logs", Category: "sessions", Description: "Bridge to buffered logs for a specific TypeScript supervised session."},
				{Path: "/api/sessions/supervisor/execute-shell", Category: "sessions", Description: "Execute a contextual shell command in a TypeScript supervised session."},
				{Path: "/api/sessions/supervisor/attach-info", Category: "sessions", Description: "Bridge to supervised-session readiness and process metadata."},
				{Path: "/api/sessions/supervisor/health", Category: "sessions", Description: "Bridge to health details for a TypeScript supervised session."},
				{Path: "/api/sessions/supervisor/state", Category: "sessions", Description: "Bridge to the shared TypeScript session-manager state."},
				{Path: "/api/sessions/supervisor/update-state", Category: "sessions", Description: "Update the shared TypeScript session-manager state."},
				{Path: "/api/sessions/supervisor/clear", Category: "sessions", Description: "Clear the shared TypeScript session-manager state."},
				{Path: "/api/sessions/supervisor/heartbeat", Category: "sessions", Description: "Touch the shared TypeScript session-manager heartbeat."},
				{Path: "/api/sessions/supervisor/restore", Category: "sessions", Description: "Restore supervised sessions through the TypeScript control plane."},
				{Path: "/api/sessions/imported/list", Category: "sessions", Description: "Bridge to imported sessions already processed by the TypeScript control plane."},
				{Path: "/api/sessions/imported/get", Category: "sessions", Description: "Bridge to a specific imported session record from the TypeScript control plane."},
				{Path: "/api/sessions/imported/scan", Category: "sessions", Description: "Trigger TypeScript imported-session scanning, import, and memory extraction."},
				{Path: "/api/sessions/imported/instruction-docs", Category: "sessions", Description: "Bridge to imported-session instruction documents generated by the TypeScript control plane."},
				{Path: "/api/sessions/imported/maintenance-stats", Category: "sessions", Description: "Bridge to imported-session archive and retention maintenance counters."},
				{Path: "/api/billing/status", Category: "providers", Description: "Read billing status, with a local Go preview when the TypeScript billing router is unavailable."},
				{Path: "/api/billing/provider-quotas", Category: "providers", Description: "Read provider quota state, with a local Go environment-backed preview when the TypeScript billing router is unavailable."},
				{Path: "/api/billing/cost-history", Category: "providers", Description: "Read billing cost history, with a local Go zero-state preview when the TypeScript billing router is unavailable."},
				{Path: "/api/billing/model-pricing", Category: "providers", Description: "Read model pricing, with a local Go catalog preview when the TypeScript billing router is unavailable."},
				{Path: "/api/billing/fallback-chain", Category: "providers", Description: "Read provider fallback chain, with a local Go routing preview when the TypeScript billing router is unavailable."},
				{Path: "/api/billing/task-routing-rules", Category: "providers", Description: "Read task routing rules, with a local Go routing preview when the TypeScript billing router is unavailable."},
				{Path: "/api/billing/routing-strategy", Category: "providers", Description: "Update global routing strategy through the TypeScript billing router."},
				{Path: "/api/billing/task-routing-rule", Category: "providers", Description: "Update task-specific routing through the TypeScript billing router."},
				{Path: "/api/billing/depleted-models", Category: "providers", Description: "Read depleted model state, with a local Go empty-state preview when the TypeScript billing router is unavailable."},
				{Path: "/api/billing/fallback-history", Category: "providers", Description: "Read provider fallback history, with a local Go empty-state preview when the TypeScript billing router is unavailable."},
				{Path: "/api/billing/fallback-history/clear", Category: "providers", Description: "Clear provider fallback history, with a local Go no-op when the TypeScript billing router is unavailable."},
				{Path: "/api/mcp/status", Category: "mcp", Description: "Bridge to TypeScript MCP runtime status and pool state."},
				{Path: "/api/mcp/servers/runtime", Category: "mcp", Description: "Bridge to TypeScript runtime MCP server visibility."},
				{Path: "/api/mcp/servers/configured", Category: "mcp", Description: "Bridge to configured MCP server records managed by the TypeScript control plane."},
				{Path: "/api/mcp/servers/get", Category: "mcp", Description: "Bridge to a specific configured MCP server record."},
				{Path: "/api/mcp/servers/create", Category: "mcp", Description: "Create a configured MCP server through the TypeScript control plane."},
				{Path: "/api/mcp/servers/update", Category: "mcp", Description: "Update a configured MCP server through the TypeScript control plane."},
				{Path: "/api/mcp/servers/delete", Category: "mcp", Description: "Delete a configured MCP server through the TypeScript control plane."},
				{Path: "/api/mcp/servers/bulk-import", Category: "mcp", Description: "Bulk import configured MCP servers through the TypeScript control plane."},
				{Path: "/api/mcp/servers/reload-metadata", Category: "mcp", Description: "Refresh cached MCP metadata through the TypeScript control plane."},
				{Path: "/api/mcp/servers/clear-metadata-cache", Category: "mcp", Description: "Clear cached MCP metadata through the TypeScript control plane."},
				{Path: "/api/mcp/servers/registry-snapshot", Category: "mcp", Description: "Bridge to the TypeScript MCP registry snapshot."},
				{Path: "/api/mcp/servers/sync-targets", Category: "mcp", Description: "Bridge to MCP client-config sync targets discovered by the TypeScript control plane."},
				{Path: "/api/mcp/servers/export-client-config", Category: "mcp", Description: "Preview a generated MCP client config through the TypeScript control plane."},
				{Path: "/api/mcp/servers/sync-client-config", Category: "mcp", Description: "Write an MCP client config through the TypeScript control plane."},
				{Path: "/api/mcp/tools", Category: "mcp", Description: "Bridge to aggregated MCP tools from the TypeScript control plane."},
				{Path: "/api/mcp/tools/search", Category: "mcp", Description: "Bridge to TypeScript MCP tool search with optional profile hinting."},
				{Path: "/api/mcp/tools/call", Category: "mcp", Description: "Execute an MCP tool through the TypeScript control plane."},
				{Path: "/api/mcp/tools/auto-call", Category: "mcp", Description: "Run one-shot semantic tool discovery and execution through the TypeScript auto_call_tool meta-tool."},
				{Path: "/api/mcp/tool-ads", Category: "mcp", Description: "Bridge goal/objective-aware tool advertisements through the TypeScript list_all_tools helper."},
				{Path: "/api/mcp/tools/schema", Category: "mcp", Description: "Hydrate and return a specific MCP tool schema through the TypeScript control plane."},
				{Path: "/api/mcp/preferences", Category: "mcp", Description: "Get or update MCP tool-selection preferences via the TypeScript control plane."},
				{Path: "/api/mcp/traffic", Category: "mcp", Description: "Read recent MCP server traffic events, with a local empty-state fallback when the TypeScript MCP router is unavailable."},
				{Path: "/api/mcp/tool-selection-telemetry", Category: "mcp", Description: "Read auto-tool selection telemetry, with a local empty-state fallback when the TypeScript MCP router is unavailable."},
				{Path: "/api/mcp/tool-selection-telemetry/clear", Category: "mcp", Description: "Clear auto-tool selection telemetry, with a local no-op fallback when the TypeScript MCP router is unavailable."},
				{Path: "/api/mcp/server-test", Category: "mcp", Description: "Run an MCP server probe, with structured local validation and failure simulation when the TypeScript MCP router is unavailable."},
				{Path: "/api/mcp/lifecycle-modes", Category: "mcp", Description: "Update MCP pool lifecycle modes, with local Go state fallback when the TypeScript MCP router is unavailable."},
				{Path: "/api/mcp/runtime-servers/add", Category: "mcp", Description: "Add a downstream runtime MCP server through the TypeScript control plane."},
				{Path: "/api/mcp/runtime-servers/remove", Category: "mcp", Description: "Remove a downstream runtime MCP server through the TypeScript control plane."},
				{Path: "/api/mcp/config/jsonc", Category: "mcp", Description: "Read or update the TypeScript MCP JSONC config through the sidecar."},
				{Path: "/api/mcp/working-set", Category: "mcp", Description: "Read the MCP working-set snapshot, with a local empty-state fallback when the TypeScript MCP router is unavailable."},
				{Path: "/api/mcp/working-set/evictions", Category: "mcp", Description: "Read MCP working-set eviction history, with a local empty-state fallback when the TypeScript MCP router is unavailable."},
				{Path: "/api/mcp/working-set/evictions/clear", Category: "mcp", Description: "Clear MCP working-set eviction history, with a local no-op fallback when the TypeScript MCP router is unavailable."},
				{Path: "/api/mcp/working-set/load", Category: "mcp", Description: "Load an MCP tool into the TypeScript working set."},
				{Path: "/api/mcp/working-set/unload", Category: "mcp", Description: "Unload an MCP tool from the TypeScript working set."},
				{Path: "/api/memory/search", Category: "memory", Description: "Bridge to TypeScript contextual memory search."},
				{Path: "/api/memory/contexts", Category: "memory", Description: "Bridge to TypeScript saved context listing."},
				{Path: "/api/memory/context/save", Category: "memory", Description: "Save a memory context through the TypeScript control plane."},
				{Path: "/api/memory/context/get", Category: "memory", Description: "Bridge to a specific saved memory context."},
				{Path: "/api/memory/context/delete", Category: "memory", Description: "Delete a saved memory context through the TypeScript control plane."},
				{Path: "/api/memory/agent-stats", Category: "memory", Description: "Bridge to TypeScript agent-memory statistics."},
				{Path: "/api/memory/agent-search", Category: "memory", Description: "Bridge to TypeScript agent-memory search."},
				{Path: "/api/memory/facts/add", Category: "memory", Description: "Add a memory fact through the TypeScript control plane."},
				{Path: "/api/memory/observations/record", Category: "memory", Description: "Record a structured observation through the TypeScript control plane."},
				{Path: "/api/memory/observations/recent", Category: "memory", Description: "Bridge to recent structured observations from the TypeScript control plane."},
				{Path: "/api/memory/observations/search", Category: "memory", Description: "Search structured observations through the TypeScript control plane."},
				{Path: "/api/memory/user-prompts/capture", Category: "memory", Description: "Capture a structured user prompt through the TypeScript control plane."},
				{Path: "/api/memory/user-prompts/recent", Category: "memory", Description: "Bridge to recent user prompts from the TypeScript control plane."},
				{Path: "/api/memory/user-prompts/search", Category: "memory", Description: "Search captured user prompts through the TypeScript control plane."},
				{Path: "/api/memory/pivot/search", Category: "memory", Description: "Bridge to pivot-based memory search through the TypeScript control plane."},
				{Path: "/api/memory/timeline/window", Category: "memory", Description: "Bridge to memory timeline window queries from the TypeScript control plane."},
				{Path: "/api/memory/cross-session-links", Category: "memory", Description: "Bridge to cross-session memory-link queries from the TypeScript control plane."},
				{Path: "/api/memory/session-bootstrap", Category: "memory", Description: "Bridge to TypeScript session bootstrap memory context."},
				{Path: "/api/memory/tool-context", Category: "memory", Description: "Bridge to TypeScript tool-context memory lookup."},
				{Path: "/api/memory/session-summaries/capture", Category: "memory", Description: "Capture a session-summary memory through the TypeScript control plane."},
				{Path: "/api/memory/session-summaries/recent", Category: "memory", Description: "Bridge to recent session-summary memories from the TypeScript control plane."},
				{Path: "/api/memory/session-summaries/search", Category: "memory", Description: "Bridge to session-summary memory search from the TypeScript control plane."},
				{Path: "/api/memory/sectioned-status", Category: "memory", Description: "Bridge to the TypeScript sectioned-memory status snapshot."},
				{Path: "/api/memory/interchange-formats", Category: "memory", Description: "Bridge to the TypeScript memory interchange-format list."},
				{Path: "/api/memory/export", Category: "memory", Description: "Bridge to TypeScript memory export, with a local export fallback from memory.json or the contexts registry when upstream is unavailable."},
				{Path: "/api/memory/import", Category: "memory", Description: "Bridge to TypeScript memory import."},
				{Path: "/api/memory/convert", Category: "memory", Description: "Bridge to TypeScript memory format conversion."},
				{Path: "/api/agent-memory/search", Category: "memory", Description: "Bridge to TypeScript agent-memory search across namespaces and tiers, with an explicit empty-result fallback when agent memory is unavailable."},
				{Path: "/api/agent-memory/add", Category: "memory", Description: "Add an agent-memory entry through the TypeScript control plane."},
				{Path: "/api/agent-memory/recent", Category: "memory", Description: "Bridge to recent TypeScript agent-memory entries, with an explicit empty-result fallback when agent memory is unavailable."},
				{Path: "/api/agent-memory/by-type", Category: "memory", Description: "Bridge to TypeScript agent-memory entries for a specific tier, with an explicit empty-result fallback when agent memory is unavailable."},
				{Path: "/api/agent-memory/by-namespace", Category: "memory", Description: "Bridge to TypeScript agent-memory entries for a specific namespace, with an explicit empty-result fallback when agent memory is unavailable."},
				{Path: "/api/agent-memory/delete", Category: "memory", Description: "Delete a TypeScript agent-memory entry by id."},
				{Path: "/api/agent-memory/clear-session", Category: "memory", Description: "Clear session-tier agent memory through the TypeScript control plane."},
				{Path: "/api/agent-memory/export", Category: "memory", Description: "Bridge to TypeScript agent-memory export, with an explicit empty export fallback when agent memory is unavailable."},
				{Path: "/api/agent-memory/handoff", Category: "memory", Description: "Create an agent-memory handoff artifact through the TypeScript control plane."},
				{Path: "/api/agent-memory/pickup", Category: "memory", Description: "Restore an agent-memory handoff artifact through the TypeScript control plane."},
				{Path: "/api/agent-memory/stats", Category: "memory", Description: "Bridge to TypeScript agent-memory counts by tier, with an explicit zero-state fallback when agent memory is unavailable."},
				{Path: "/api/graph", Category: "code", Description: "Bridge to the TypeScript repository graph snapshot."},
				{Path: "/api/graph/rebuild", Category: "code", Description: "Rebuild the TypeScript repository graph and return the latest snapshot."},
				{Path: "/api/graph/consumers", Category: "code", Description: "Bridge to repository graph consumers for a given file path, with an explicit empty-state fallback when the repo graph is unavailable."},
				{Path: "/api/graph/dependencies", Category: "code", Description: "Bridge to repository graph dependencies for a given file path, with an explicit empty-state fallback when the repo graph is unavailable."},
				{Path: "/api/graph/symbols", Category: "code", Description: "Bridge to the TypeScript symbol graph, with an explicit empty graph fallback when symbol graph data is unavailable."},
				{Path: "/api/context/list", Category: "code", Description: "Bridge to the current TypeScript context file list, with a local empty-state fallback when the TypeScript context manager is unavailable."},
				{Path: "/api/context/add", Category: "code", Description: "Add a file to the TypeScript context manager."},
				{Path: "/api/context/remove", Category: "code", Description: "Remove a file from the TypeScript context manager."},
				{Path: "/api/context/clear", Category: "code", Description: "Clear the TypeScript context manager state."},
				{Path: "/api/context/prompt", Category: "code", Description: "Bridge to the TypeScript context prompt output, with a local empty-state fallback when the TypeScript context manager is unavailable."},
				{Path: "/api/git/modules", Category: "code", Description: "Bridge to parsed git submodule metadata, with a local .gitmodules fallback when the TypeScript control plane is unavailable."},
				{Path: "/api/git/log", Category: "code", Description: "Read git log output through the TypeScript control plane, with a local Go git fallback when the router is unavailable."},
				{Path: "/api/git/status", Category: "code", Description: "Read git status through the TypeScript control plane, with a local Go git fallback when the router is unavailable."},
				{Path: "/api/git/revert", Category: "code", Description: "Request a git revert through the TypeScript control plane."},
				{Path: "/api/tests/status", Category: "code", Description: "Bridge to TypeScript auto-test service status, with a local zero-state fallback when the auto-test service is unavailable."},
				{Path: "/api/tests/start", Category: "code", Description: "Start the TypeScript auto-test service."},
				{Path: "/api/tests/stop", Category: "code", Description: "Stop the TypeScript auto-test service."},
				{Path: "/api/tests/run", Category: "code", Description: "Run the relevant TypeScript test file for a given source path."},
				{Path: "/api/tests/results", Category: "code", Description: "Bridge to recent TypeScript auto-test results, with a local empty-state fallback when the auto-test service is unavailable."},
				{Path: "/api/autodev/start-loop", Category: "code", Description: "Start an autoDev loop through the TypeScript autoDev router."},
				{Path: "/api/autodev/cancel-loop", Category: "code", Description: "Cancel an autoDev loop through the TypeScript autoDev router."},
				{Path: "/api/autodev/loops", Category: "code", Description: "List autoDev loops through the TypeScript autoDev router."},
				{Path: "/api/autodev/loop", Category: "code", Description: "Read one autoDev loop through the TypeScript autoDev router."},
				{Path: "/api/autodev/clear-completed", Category: "code", Description: "Clear completed autoDev loops through the TypeScript autoDev router."},
				{Path: "/api/darwin/evolve", Category: "code", Description: "Propose a Darwin mutation through the TypeScript darwin router."},
				{Path: "/api/darwin/experiment", Category: "code", Description: "Start a Darwin experiment through the TypeScript darwin router."},
				{Path: "/api/darwin/status", Category: "code", Description: "Read Darwin experiment status through the TypeScript darwin router."},
				{Path: "/api/squad", Category: "agents", Description: "List squad members through the TypeScript squad router."},
				{Path: "/api/squad/spawn", Category: "agents", Description: "Spawn a squad member through the TypeScript squad router."},
				{Path: "/api/squad/kill", Category: "agents", Description: "Terminate a squad member through the TypeScript squad router."},
				{Path: "/api/squad/chat", Category: "agents", Description: "Send a message to a squad member through the TypeScript squad router."},
				{Path: "/api/squad/indexer/toggle", Category: "agents", Description: "Toggle the squad indexer through the TypeScript squad router."},
				{Path: "/api/squad/indexer/status", Category: "agents", Description: "Read squad indexer status through the TypeScript squad router."},
				{Path: "/api/supervisor/decompose", Category: "agents", Description: "Decompose a goal through the TypeScript supervisor router."},
				{Path: "/api/supervisor/supervise", Category: "agents", Description: "Run a supervised task through the TypeScript supervisor router."},
				{Path: "/api/supervisor/status", Category: "agents", Description: "Read supervisor status through the TypeScript supervisor router."},
				{Path: "/api/supervisor/tasks", Category: "agents", Description: "List supervisor tasks through the TypeScript supervisor router."},
				{Path: "/api/supervisor/cancel", Category: "agents", Description: "Cancel a supervisor task through the TypeScript supervisor router."},
				{Path: "/api/metrics/stats", Category: "ops", Description: "Read aggregated metrics stats for a time window, with a local zero-state Go fallback when the TypeScript metrics router is unavailable."},
				{Path: "/api/metrics/track", Category: "ops", Description: "Track a custom metric event through the TypeScript control plane."},
				{Path: "/api/metrics/system-snapshot", Category: "ops", Description: "Read a real-time system resource snapshot, with a native Go fallback when the TypeScript metrics router is unavailable."},
				{Path: "/api/metrics/timeline", Category: "ops", Description: "Read downsampled metrics timeline data, with a local zero-state Go fallback when the TypeScript metrics router is unavailable."},
				{Path: "/api/metrics/provider-breakdown", Category: "ops", Description: "Read provider request, latency, and cost breakdowns, with a local zero-usage Go fallback when the TypeScript metrics router is unavailable."},
				{Path: "/api/metrics/monitoring", Category: "ops", Description: "Toggle TypeScript metrics monitoring state."},
				{Path: "/api/metrics/routing-history", Category: "ops", Description: "Read recent LLM routing and failover decisions, with a local empty-state Go fallback when the TypeScript metrics router is unavailable."},
				{Path: "/api/logs", Category: "ops", Description: "List observability logs, with a local metamcp.db fallback when the TypeScript log store is unavailable."},
				{Path: "/api/logs/summary", Category: "ops", Description: "Read the observability summary rollup, with a local metamcp.db fallback when the TypeScript log store is unavailable."},
				{Path: "/api/logs/clear", Category: "ops", Description: "Clear observability logs, with a local metamcp.db fallback when the TypeScript log store is unavailable."},
				{Path: "/api/server-health/check", Category: "ops", Description: "Bridge to the TypeScript MCP server health state for a specific server UUID, with a local cached mcp.jsonc metadata fallback when unavailable."},
				{Path: "/api/server-health/reset", Category: "ops", Description: "Reset the TypeScript MCP server health error state for a specific server UUID."},
				{Path: "/api/settings", Category: "control", Description: "Bridge to the full TypeScript configuration object, with a local Go .borg/config.json fallback when unavailable."},
				{Path: "/api/settings/update", Category: "control", Description: "Update the TypeScript configuration object with a partial config payload."},
				{Path: "/api/settings/providers", Category: "control", Description: "Read provider visibility, with a local Go provider catalog fallback when the TypeScript settings router is unavailable."},
				{Path: "/api/settings/test-connection", Category: "control", Description: "Test a provider connection through the TypeScript control plane."},
				{Path: "/api/settings/environment", Category: "control", Description: "Read environment diagnostics through the TypeScript settings router, with a local Go runtime fallback when the router is unavailable."},
				{Path: "/api/settings/mcp-servers", Category: "control", Description: "Read configured MCP servers through the TypeScript settings router, with a local Go config fallback when the router is unavailable."},
				{Path: "/api/settings/provider-key", Category: "control", Description: "Persist a provider key through the TypeScript settings layer."},
				{Path: "/api/tools", Category: "control", Description: "List tools, with a local source-backed Go inventory fallback when the TypeScript tools router is unavailable."},
				{Path: "/api/tools/by-server", Category: "control", Description: "List tools filtered by MCP server, with a local source-backed Go inventory fallback when the TypeScript tools router is unavailable."},
				{Path: "/api/tools/search", Category: "control", Description: "Search tools, with a local source-backed Go inventory fallback when the TypeScript tools router is unavailable."},
				{Path: "/api/tools/context", Category: "control", Description: "Go-owned tool guidance snapshot combining startup readiness, tool context memory, and related tool advertisements."},
				{Path: "/api/tools/detect-cli-harnesses", Category: "control", Description: "Read CLI harness detection through the TypeScript tools router, with a local Go runtime fallback when the router is unavailable."},
				{Path: "/api/tools/detect-execution-environment", Category: "control", Description: "Read execution-environment diagnostics through the TypeScript tools router, with a local Go runtime fallback when the router is unavailable."},
				{Path: "/api/tools/detect-install-surfaces", Category: "control", Description: "Read install-surface artifact detection through the TypeScript tools router, with a local Go filesystem fallback when the router is unavailable."},
				{Path: "/api/tools/get", Category: "control", Description: "Read a specific tool definition, with a local metamcp.db tool inventory fallback when the TypeScript tools router is unavailable."},
				{Path: "/api/tools/create", Category: "control", Description: "Create a tool through the TypeScript control plane."},
				{Path: "/api/tools/upsert-batch", Category: "control", Description: "Upsert a batch of tools through the TypeScript control plane."},
				{Path: "/api/tools/delete", Category: "control", Description: "Delete a tool through the TypeScript control plane."},
				{Path: "/api/tools/always-on", Category: "control", Description: "Toggle the TypeScript always-on state for a tool."},
				{Path: "/api/tool-sets", Category: "control", Description: "Bridge to the TypeScript tool-set list."},
				{Path: "/api/tool-sets/get", Category: "control", Description: "Bridge to a specific TypeScript tool set."},
				{Path: "/api/tool-sets/create", Category: "control", Description: "Create a tool set through the TypeScript control plane."},
				{Path: "/api/tool-sets/update", Category: "control", Description: "Update a tool set through the TypeScript control plane."},
				{Path: "/api/tool-sets/delete", Category: "control", Description: "Delete a tool set through the TypeScript control plane."},
				{Path: "/api/project/context", Category: "control", Description: "Bridge to the TypeScript project context document, with a local .borg/project_context.md fallback when the TypeScript control plane is unavailable."},
				{Path: "/api/project/context/update", Category: "control", Description: "Update the TypeScript project context document."},
				{Path: "/api/project/handoffs", Category: "control", Description: "Bridge to TypeScript project handoff metadata, with a local .borg/handoffs listing fallback when the TypeScript control plane is unavailable."},
				{Path: "/api/shell/log", Category: "control", Description: "Log a shell command through the TypeScript shell service."},
				{Path: "/api/shell/history/query", Category: "control", Description: "Bridge to TypeScript shell history search, with a local .borg/shell_history.json fallback when unavailable."},
				{Path: "/api/shell/history/system", Category: "control", Description: "Bridge to recent TypeScript system shell history, with a local shell history file fallback when unavailable."},
				{Path: "/api/agent/tool", Category: "agents", Description: "Run a tool through the TypeScript agent router."},
				{Path: "/api/agent/chat", Category: "agents", Description: "Bridge to the TypeScript agent chat surface."},
				{Path: "/api/commands/execute", Category: "agents", Description: "Execute a TypeScript command-registry entry."},
				{Path: "/api/commands", Category: "agents", Description: "Bridge to the TypeScript command registry list, with a local empty-state fallback when the registry is unavailable."},
				{Path: "/api/skills", Category: "agents", Description: "Bridge to the TypeScript skill registry list."},
				{Path: "/api/skills/summary", Category: "agents", Description: "List skills with progressive-disclosure metadata only (id, name, folder)."},
				{Path: "/api/skills/read", Category: "agents", Description: "Read a skill through the TypeScript skill registry."},
				{Path: "/api/skills/create", Category: "agents", Description: "Create a skill through the TypeScript skill registry."},
				{Path: "/api/skills/save", Category: "agents", Description: "Save skill content through the TypeScript skill registry."},
				{Path: "/api/skills/assimilate", Category: "agents", Description: "Assimilate docs into a skill through the TypeScript skill-assimilation service."},
				{Path: "/api/workflows", Category: "workflow", Description: "Bridge to TypeScript workflow definitions."},
				{Path: "/api/workflows/graph", Category: "workflow", Description: "Bridge to a TypeScript workflow graph."},
				{Path: "/api/workflows/start", Category: "workflow", Description: "Start a TypeScript workflow execution."},
				{Path: "/api/workflows/executions", Category: "workflow", Description: "List TypeScript workflow executions."},
				{Path: "/api/workflows/execution", Category: "workflow", Description: "Bridge to a TypeScript workflow execution record."},
				{Path: "/api/workflows/history", Category: "workflow", Description: "Bridge to TypeScript workflow execution history."},
				{Path: "/api/workflows/resume", Category: "workflow", Description: "Resume a TypeScript workflow execution."},
				{Path: "/api/workflows/pause", Category: "workflow", Description: "Pause a TypeScript workflow execution."},
				{Path: "/api/workflows/approve", Category: "workflow", Description: "Approve a TypeScript workflow execution."},
				{Path: "/api/workflows/reject", Category: "workflow", Description: "Reject a TypeScript workflow execution."},
				{Path: "/api/workflows/canvases", Category: "workflow", Description: "List saved TypeScript workflow canvases."},
				{Path: "/api/workflows/canvas", Category: "workflow", Description: "Load a saved TypeScript workflow canvas."},
				{Path: "/api/workflows/canvas/save", Category: "workflow", Description: "Save a TypeScript workflow canvas."},
				{Path: "/api/symbols", Category: "code", Description: "List pinned symbols through the TypeScript symbols router, with an explicit empty-state fallback when symbol pins are unavailable."},
				{Path: "/api/symbols/find", Category: "code", Description: "Search symbols through the TypeScript symbols router, with an explicit empty-state fallback when symbol search is unavailable."},
				{Path: "/api/symbols/pin", Category: "code", Description: "Pin a symbol through the TypeScript symbols router."},
				{Path: "/api/symbols/unpin", Category: "code", Description: "Unpin a symbol through the TypeScript symbols router."},
				{Path: "/api/symbols/priority", Category: "code", Description: "Update symbol priority through the TypeScript symbols router."},
				{Path: "/api/symbols/notes", Category: "code", Description: "Add symbol notes through the TypeScript symbols router."},
				{Path: "/api/symbols/clear", Category: "code", Description: "Clear pinned symbols through the TypeScript symbols router."},
				{Path: "/api/symbols/file", Category: "code", Description: "List pinned symbols for a file through the TypeScript symbols router, with an explicit empty-state fallback when symbol pins are unavailable."},
				{Path: "/api/lsp/find-symbol", Category: "code", Description: "Bridge to the TypeScript LSP find-symbol surface."},
				{Path: "/api/lsp/find-references", Category: "code", Description: "Bridge to the TypeScript LSP reference search surface."},
				{Path: "/api/lsp/symbols", Category: "code", Description: "Bridge to the TypeScript LSP file-symbol surface."},
				{Path: "/api/lsp/search", Category: "code", Description: "Bridge to the TypeScript LSP symbol-search surface."},
				{Path: "/api/lsp/index", Category: "code", Description: "Trigger TypeScript LSP indexing through the bridge."},
				{Path: "/api/api-keys", Category: "governance", Description: "List API keys through the TypeScript API keys router, with a local empty-state fallback when the key store is unavailable."},
				{Path: "/api/api-keys/get", Category: "governance", Description: "Read an API key through the TypeScript API keys router."},
				{Path: "/api/api-keys/create", Category: "governance", Description: "Create an API key through the TypeScript API keys router."},
				{Path: "/api/api-keys/update", Category: "governance", Description: "Update an API key through the TypeScript API keys router."},
				{Path: "/api/api-keys/delete", Category: "governance", Description: "Delete an API key through the TypeScript API keys router."},
				{Path: "/api/api-keys/validate", Category: "governance", Description: "Validate an API key through the TypeScript API keys router."},
				{Path: "/api/audit", Category: "governance", Description: "List audit logs through the TypeScript audit router, with a local empty-state fallback when the audit service is unavailable."},
				{Path: "/api/audit/query", Category: "governance", Description: "Query audit logs through the TypeScript audit router, with a local empty-state fallback when the audit service is unavailable."},
				{Path: "/api/scripts", Category: "operator", Description: "List saved scripts through the TypeScript saved scripts router, with a local borg config fallback when unavailable."},
				{Path: "/api/scripts/get", Category: "operator", Description: "Read a saved script through the TypeScript saved scripts router, with a local borg config fallback when unavailable."},
				{Path: "/api/scripts/create", Category: "operator", Description: "Create a saved script through the TypeScript saved scripts router."},
				{Path: "/api/scripts/update", Category: "operator", Description: "Update a saved script through the TypeScript saved scripts router."},
				{Path: "/api/scripts/delete", Category: "operator", Description: "Delete a saved script through the TypeScript saved scripts router."},
				{Path: "/api/scripts/execute", Category: "operator", Description: "Execute a saved script through the TypeScript saved scripts router."},
				{Path: "/api/links-backlog", Category: "operator", Description: "List BobbyBookmarks backlog links through the TypeScript links backlog router."},
				{Path: "/api/links-backlog/stats", Category: "operator", Description: "Read BobbyBookmarks backlog stats through the TypeScript links backlog router."},
				{Path: "/api/links-backlog/get", Category: "operator", Description: "Read a BobbyBookmarks backlog item through the TypeScript links backlog router."},
				{Path: "/api/links-backlog/sync", Category: "operator", Description: "Sync BobbyBookmarks backlog data through the TypeScript links backlog router."},
				{Path: "/api/infrastructure", Category: "operator", Description: "Read infrastructure daemon status through the TypeScript infrastructure router, with a local binary/config fallback when the router is unavailable."},
				{Path: "/api/infrastructure/doctor", Category: "operator", Description: "Run the infrastructure doctor command through the TypeScript infrastructure router."},
				{Path: "/api/infrastructure/apply", Category: "operator", Description: "Apply infrastructure configuration through the TypeScript infrastructure router."},
				{Path: "/api/expert/research", Category: "agents", Description: "Dispatch a research task through the TypeScript expert router."},
				{Path: "/api/expert/code", Category: "agents", Description: "Dispatch a coding task through the TypeScript expert router."},
				{Path: "/api/expert/status", Category: "agents", Description: "Read TypeScript expert agent status, with a local offline-state fallback when the expert agents are unavailable."},
				{Path: "/api/autonomy/get-level", Category: "governance", Description: "Read the current autonomy level through the TypeScript autonomy router."},
				{Path: "/api/autonomy/set-level", Category: "governance", Description: "Set autonomy level through the TypeScript autonomy router."},
				{Path: "/api/autonomy/activate-full", Category: "governance", Description: "Activate full autonomy through the TypeScript autonomy router."},
				{Path: "/api/director/memorize", Category: "governance", Description: "Send memory content to the TypeScript director router."},
				{Path: "/api/director/chat", Category: "governance", Description: "Chat with the TypeScript director runtime."},
				{Path: "/api/director/status", Category: "governance", Description: "Read TypeScript director runtime status."},
				{Path: "/api/director/config/update", Category: "governance", Description: "Update TypeScript director config."},
				{Path: "/api/director-config", Category: "governance", Description: "Read TypeScript directorConfig settings."},
				{Path: "/api/director-config/test", Category: "governance", Description: "Run TypeScript directorConfig readiness checks."},
				{Path: "/api/director-config/update", Category: "governance", Description: "Update TypeScript directorConfig settings."},
				{Path: "/api/director/auto-drive/stop", Category: "governance", Description: "Stop auto-drive through the TypeScript director router."},
				{Path: "/api/director/auto-drive/start", Category: "governance", Description: "Start auto-drive through the TypeScript director router."},
				{Path: "/api/council/members", Category: "governance", Description: "Read council members through the TypeScript council router."},
				{Path: "/api/council/members/update", Category: "governance", Description: "Update council members through the TypeScript council router."},
				{Path: "/api/council/status", Category: "governance", Description: "Read council base status through the TypeScript council router."},
				{Path: "/api/council/config/update", Category: "governance", Description: "Update council base config through the TypeScript council router."},
				{Path: "/api/council/supervisors/add", Category: "governance", Description: "Add council supervisors through the TypeScript council router."},
				{Path: "/api/council/supervisors/clear", Category: "governance", Description: "Clear council supervisors through the TypeScript council router."},
				{Path: "/api/council/debate", Category: "governance", Description: "Run a council debate through the TypeScript council router."},
				{Path: "/api/council/toggle", Category: "governance", Description: "Toggle the council through the TypeScript council router."},
				{Path: "/api/council/mock/add", Category: "governance", Description: "Add a mock council supervisor through the TypeScript council router."},
				{Path: "/api/council/sessions", Category: "governance", Description: "List council sessions through the TypeScript council router."},
				{Path: "/api/council/sessions/active", Category: "governance", Description: "List active council sessions through the TypeScript council router."},
				{Path: "/api/council/sessions/stats", Category: "governance", Description: "Read council session stats through the TypeScript council router."},
				{Path: "/api/council/sessions/get", Category: "governance", Description: "Read a specific council session through the TypeScript council router."},
				{Path: "/api/council/sessions/start", Category: "governance", Description: "Start a council session through the TypeScript council router."},
				{Path: "/api/council/sessions/bulk-start", Category: "governance", Description: "Start multiple council sessions through the TypeScript council router."},
				{Path: "/api/council/sessions/bulk-stop", Category: "governance", Description: "Stop all council sessions through the TypeScript council router."},
				{Path: "/api/council/sessions/bulk-resume", Category: "governance", Description: "Resume all council sessions through the TypeScript council router."},
				{Path: "/api/council/sessions/stop", Category: "governance", Description: "Stop a council session through the TypeScript council router."},
				{Path: "/api/council/sessions/resume", Category: "governance", Description: "Resume a council session through the TypeScript council router."},
				{Path: "/api/council/sessions/delete", Category: "governance", Description: "Delete a council session through the TypeScript council router."},
				{Path: "/api/council/sessions/guidance", Category: "governance", Description: "Send guidance to a council session through the TypeScript council router."},
				{Path: "/api/council/sessions/logs", Category: "governance", Description: "Read council session logs through the TypeScript council router."},
				{Path: "/api/council/sessions/templates", Category: "governance", Description: "Read council session templates through the TypeScript council router."},
				{Path: "/api/council/sessions/from-template", Category: "governance", Description: "Start a council session from a template through the TypeScript council router."},
				{Path: "/api/council/sessions/persisted", Category: "governance", Description: "List persisted council sessions through the TypeScript council router."},
				{Path: "/api/council/sessions/by-tag", Category: "governance", Description: "List council sessions by tag through the TypeScript council router."},
				{Path: "/api/council/sessions/by-template", Category: "governance", Description: "List council sessions by template through the TypeScript council router."},
				{Path: "/api/council/sessions/by-cli", Category: "governance", Description: "List council sessions by CLI type through the TypeScript council router."},
				{Path: "/api/council/sessions/tags/update", Category: "governance", Description: "Replace council session tags through the TypeScript council router."},
				{Path: "/api/council/sessions/tags/add", Category: "governance", Description: "Add a council session tag through the TypeScript council router."},
				{Path: "/api/council/sessions/tags/remove", Category: "governance", Description: "Remove a council session tag through the TypeScript council router."},
				{Path: "/api/council/quota/status", Category: "governance", Description: "Read council quota status through the TypeScript council router."},
				{Path: "/api/council/quota/config", Category: "governance", Description: "Read or update council quota config through the TypeScript council router."},
				{Path: "/api/council/quota/enabled", Category: "governance", Description: "Enable or disable council quota enforcement through the TypeScript council router."},
				{Path: "/api/council/quota/check", Category: "governance", Description: "Check council quota availability for a provider through the TypeScript council router."},
				{Path: "/api/council/quota/stats", Category: "governance", Description: "Read council quota stats through the TypeScript council router."},
				{Path: "/api/council/quota/limits", Category: "governance", Description: "Read or update council quota limits through the TypeScript council router."},
				{Path: "/api/council/quota/reset", Category: "governance", Description: "Reset council quota usage through the TypeScript council router."},
				{Path: "/api/council/quota/unthrottle", Category: "governance", Description: "Unthrottle a council quota provider through the TypeScript council router."},
				{Path: "/api/council/quota/record-request", Category: "governance", Description: "Record a council quota request through the TypeScript council router."},
				{Path: "/api/council/quota/rate-limit-error", Category: "governance", Description: "Record a council quota rate-limit error through the TypeScript council router."},
				{Path: "/api/council/history/status", Category: "governance", Description: "Read council debate-history status through the TypeScript history router."},
				{Path: "/api/council/history/config", Category: "governance", Description: "Read or update council debate-history config through the TypeScript history router."},
				{Path: "/api/council/history/toggle", Category: "governance", Description: "Toggle council debate-history through the TypeScript history router."},
				{Path: "/api/council/history/stats", Category: "governance", Description: "Read council debate-history stats through the TypeScript history router."},
				{Path: "/api/council/history/list", Category: "governance", Description: "List council debate-history records through the TypeScript history router."},
				{Path: "/api/council/history/get", Category: "governance", Description: "Read a council debate-history record through the TypeScript history router."},
				{Path: "/api/council/history/delete", Category: "governance", Description: "Delete a council debate-history record through the TypeScript history router."},
				{Path: "/api/council/history/supervisor", Category: "governance", Description: "Read council supervisor vote history through the TypeScript history router."},
				{Path: "/api/council/history/clear", Category: "governance", Description: "Clear council debate-history through the TypeScript history router."},
				{Path: "/api/council/history/initialize", Category: "governance", Description: "Initialize council debate-history through the TypeScript history router."},
				{Path: "/api/council/smart-pilot/status", Category: "governance", Description: "Read council smart-pilot status through the TypeScript smartPilot router."},
				{Path: "/api/council/smart-pilot/config", Category: "governance", Description: "Read or update council smart-pilot config through the TypeScript smartPilot router."},
				{Path: "/api/council/smart-pilot/trigger", Category: "governance", Description: "Trigger a smart-pilot task through the TypeScript smartPilot router."},
				{Path: "/api/council/smart-pilot/reset-count", Category: "governance", Description: "Reset a smart-pilot approval count through the TypeScript smartPilot router."},
				{Path: "/api/council/smart-pilot/reset-all", Category: "governance", Description: "Reset all smart-pilot approval counts through the TypeScript smartPilot router."},
				{Path: "/api/council/hooks", Category: "governance", Description: "List registered council auto-continue hooks through the TypeScript hooks router."},
				{Path: "/api/council/hooks/register", Category: "governance", Description: "Register a council auto-continue hook through the TypeScript hooks router."},
				{Path: "/api/council/hooks/unregister", Category: "governance", Description: "Unregister a council auto-continue hook through the TypeScript hooks router."},
				{Path: "/api/council/hooks/clear", Category: "governance", Description: "Clear all registered council auto-continue hooks through the TypeScript hooks router."},
				{Path: "/api/council/ide/status", Category: "governance", Description: "Read council IDE bridge status through the TypeScript IDE router."},
				{Path: "/api/council/ide/submit-task", Category: "governance", Description: "Submit an IDE-generated task through the TypeScript IDE router."},
				{Path: "/api/council/evolution/start", Category: "governance", Description: "Start council self-evolution through the TypeScript evolution router."},
				{Path: "/api/council/evolution/stop", Category: "governance", Description: "Stop council self-evolution through the TypeScript evolution router."},
				{Path: "/api/council/evolution/optimize", Category: "governance", Description: "Optimize council self-evolution weights through the TypeScript evolution router."},
				{Path: "/api/council/evolution/evolve", Category: "governance", Description: "Run a council evolution task through the TypeScript evolution router."},
				{Path: "/api/council/evolution/test", Category: "governance", Description: "Run a council evolution self-test through the TypeScript evolution router."},
				{Path: "/api/council/fine-tune/datasets", Category: "governance", Description: "Create or list fine-tuning datasets through the TypeScript fineTune router."},
				{Path: "/api/council/fine-tune/datasets/get", Category: "governance", Description: "Read a fine-tuning dataset through the TypeScript fineTune router."},
				{Path: "/api/council/fine-tune/jobs", Category: "governance", Description: "Create or list fine-tuning jobs through the TypeScript fineTune router."},
				{Path: "/api/council/fine-tune/jobs/start", Category: "governance", Description: "Start a fine-tuning job through the TypeScript fineTune router."},
				{Path: "/api/council/fine-tune/models", Category: "governance", Description: "Register or list fine-tuned models through the TypeScript fineTune router."},
				{Path: "/api/council/fine-tune/models/deploy", Category: "governance", Description: "Deploy a fine-tuned model through the TypeScript fineTune router."},
				{Path: "/api/council/fine-tune/chat", Category: "governance", Description: "Chat through a deployed fine-tuned model via the TypeScript fineTune router."},
				{Path: "/api/council/fine-tune/stats", Category: "governance", Description: "Read fine-tuning statistics through the TypeScript fineTune router."},
				{Path: "/api/council/rotation", Category: "governance", Description: "List shared-context council rotation rooms through the TypeScript rotation router."},
				{Path: "/api/council/rotation/get", Category: "governance", Description: "Read a shared-context council rotation room through the TypeScript rotation router."},
				{Path: "/api/council/rotation/create", Category: "governance", Description: "Create a shared-context council rotation room through the TypeScript rotation router."},
				{Path: "/api/council/rotation/add-participant", Category: "governance", Description: "Add a participant to a council rotation room through the TypeScript rotation router."},
				{Path: "/api/council/rotation/post-message", Category: "governance", Description: "Post a council rotation chatroom message through the TypeScript rotation router."},
				{Path: "/api/council/rotation/set-agreement", Category: "governance", Description: "Record a plan-mode agreement vote through the TypeScript rotation router."},
				{Path: "/api/council/rotation/advance-turn", Category: "governance", Description: "Advance a council rotation room turn through the TypeScript rotation router."},
				{Path: "/api/council/rotation/configure-supervisor", Category: "governance", Description: "Configure a supervisor for a council rotation room through the TypeScript rotation router."},
				{Path: "/api/council/rotation/run-supervisor-check", Category: "governance", Description: "Run a supervisor evaluation for a council rotation room through the TypeScript rotation router."},
				{Path: "/api/council/rotation/update-shared-context", Category: "governance", Description: "Update shared context for a council rotation room through the TypeScript rotation router."},
				{Path: "/api/council/rotation/pause", Category: "governance", Description: "Pause a council rotation room through the TypeScript rotation router."},
				{Path: "/api/council/rotation/resume", Category: "governance", Description: "Resume a council rotation room through the TypeScript rotation router."},
				{Path: "/api/council/rotation/start-execution", Category: "governance", Description: "Start execution mode for a council rotation room through the TypeScript rotation router."},
				{Path: "/api/council/rotation/complete", Category: "governance", Description: "Complete a council rotation room through the TypeScript rotation router."},
				{Path: "/api/council/visual/system-diagram", Category: "governance", Description: "Read the council system diagram through the TypeScript visual router."},
				{Path: "/api/council/visual/plan-diagram", Category: "governance", Description: "Render a council plan diagram through the TypeScript visual router."},
				{Path: "/api/council/visual/parse-plan", Category: "governance", Description: "Parse a council Mermaid plan through the TypeScript visual router."},
				{Path: "/api/deerflow/status", Category: "operator", Description: "Read DeerFlow bridge availability through the TypeScript DeerFlow router."},
				{Path: "/api/deerflow/models", Category: "operator", Description: "List DeerFlow models through the TypeScript DeerFlow router."},
				{Path: "/api/deerflow/skills", Category: "operator", Description: "List DeerFlow skills through the TypeScript DeerFlow router."},
				{Path: "/api/deerflow/memory", Category: "operator", Description: "Read DeerFlow memory status through the TypeScript DeerFlow router."},
				{Path: "/api/healer/diagnose", Category: "operator", Description: "Analyze an error through the TypeScript healer router."},
				{Path: "/api/healer/heal", Category: "operator", Description: "Attempt a heal action through the TypeScript healer router."},
				{Path: "/api/healer/history", Category: "operator", Description: "Read healer history through the TypeScript healer router, with an explicit empty-history fallback when the healer runtime is unavailable."},
				{Path: "/api/clouddev/providers", Category: "operator", Description: "List cloud development providers through the TypeScript cloudDev router."},
				{Path: "/api/clouddev/sessions/create", Category: "operator", Description: "Create a cloud development session through the TypeScript cloudDev router."},
				{Path: "/api/clouddev/sessions", Category: "operator", Description: "List cloud development sessions through the TypeScript cloudDev router."},
				{Path: "/api/clouddev/sessions/get", Category: "operator", Description: "Read a cloud development session through the TypeScript cloudDev router."},
				{Path: "/api/clouddev/sessions/status", Category: "operator", Description: "Update a cloud development session status through the TypeScript cloudDev router."},
				{Path: "/api/clouddev/sessions/delete", Category: "operator", Description: "Delete a cloud development session through the TypeScript cloudDev router."},
				{Path: "/api/clouddev/messages/send", Category: "operator", Description: "Send a cloud development session message through the TypeScript cloudDev router."},
				{Path: "/api/clouddev/messages/broadcast", Category: "operator", Description: "Broadcast a cloud development session message through the TypeScript cloudDev router."},
				{Path: "/api/clouddev/messages/preview-recipients", Category: "operator", Description: "Preview cloud development broadcast recipients through the TypeScript cloudDev router."},
				{Path: "/api/clouddev/plan/accept", Category: "operator", Description: "Accept a cloud development plan through the TypeScript cloudDev router."},
				{Path: "/api/clouddev/plan/auto-accept", Category: "operator", Description: "Toggle cloud development auto-accept through the TypeScript cloudDev router."},
				{Path: "/api/clouddev/messages/get", Category: "operator", Description: "Read cloud development session messages through the TypeScript cloudDev router."},
				{Path: "/api/clouddev/logs", Category: "operator", Description: "Read cloud development session logs through the TypeScript cloudDev router."},
				{Path: "/api/clouddev/stats", Category: "operator", Description: "Read aggregate cloud development stats through the TypeScript cloudDev router."},
				{Path: "/api/policies", Category: "governance", Description: "List policies through the TypeScript policies router, with a local empty-state fallback when the policy store is unavailable."},
				{Path: "/api/policies/get", Category: "governance", Description: "Read a policy through the TypeScript policies router."},
				{Path: "/api/policies/create", Category: "governance", Description: "Create a policy through the TypeScript policies router."},
				{Path: "/api/policies/update", Category: "governance", Description: "Update a policy through the TypeScript policies router."},
				{Path: "/api/policies/delete", Category: "governance", Description: "Delete a policy through the TypeScript policies router."},
				{Path: "/api/secrets", Category: "governance", Description: "List secrets through the TypeScript secrets router."},
				{Path: "/api/secrets/set", Category: "governance", Description: "Set a secret through the TypeScript secrets router."},
				{Path: "/api/secrets/delete", Category: "governance", Description: "Delete a secret through the TypeScript secrets router."},
				{Path: "/api/marketplace", Category: "registry", Description: "List marketplace entries through the TypeScript marketplace router."},
				{Path: "/api/marketplace/install", Category: "registry", Description: "Install a marketplace entry through the TypeScript marketplace router."},
				{Path: "/api/marketplace/publish", Category: "registry", Description: "Publish a marketplace entry through the TypeScript marketplace router."},
				{Path: "/api/catalog", Category: "registry", Description: "List published catalog entries through the TypeScript catalog router."},
				{Path: "/api/catalog/get", Category: "registry", Description: "Read a published catalog entry through the TypeScript catalog router."},
				{Path: "/api/catalog/runs", Category: "registry", Description: "List validation runs through the TypeScript catalog router."},
				{Path: "/api/catalog/ingest", Category: "registry", Description: "Trigger published catalog ingestion through the TypeScript catalog router."},
				{Path: "/api/catalog/validate", Category: "registry", Description: "Trigger published catalog validation through the TypeScript catalog router."},
				{Path: "/api/catalog/install", Category: "registry", Description: "Install an MCP server from a TypeScript catalog recipe."},
				{Path: "/api/catalog/validate-batch", Category: "registry", Description: "Trigger batch catalog validation through the TypeScript catalog router."},
				{Path: "/api/catalog/stats", Category: "registry", Description: "Read catalog summary stats through the TypeScript catalog router."},
				{Path: "/api/catalog/linked-servers", Category: "registry", Description: "List managed servers linked to a published catalog entry."},
				{Path: "/api/oauth/clients/create", Category: "auth", Description: "Create an OAuth client through the TypeScript OAuth router."},
				{Path: "/api/oauth/clients/get", Category: "auth", Description: "Read an OAuth client through the TypeScript OAuth router."},
				{Path: "/api/oauth/sessions/upsert", Category: "auth", Description: "Upsert an OAuth session through the TypeScript OAuth router."},
				{Path: "/api/oauth/sessions/by-server", Category: "auth", Description: "Read an OAuth session for an MCP server through the TypeScript OAuth router."},
				{Path: "/api/oauth/exchange", Category: "auth", Description: "Exchange an OAuth authorization code through the TypeScript OAuth router."},
				{Path: "/api/research/conduct", Category: "research", Description: "Run a research task through the TypeScript research router."},
				{Path: "/api/research/ingest", Category: "research", Description: "Ingest a research URL through the TypeScript research router."},
				{Path: "/api/research/recursive", Category: "research", Description: "Run recursive research through the TypeScript research router."},
				{Path: "/api/research/queries", Category: "research", Description: "Generate research queries through the TypeScript research router, with a local topic-as-query fallback when the deep research service is unavailable."},
				{Path: "/api/research/queue", Category: "research", Description: "Read research ingestion queue state through the TypeScript research router, with a local file-backed fallback when the router is unavailable."},
				{Path: "/api/research/retry-failed", Category: "research", Description: "Retry a failed research URL through the TypeScript research router."},
				{Path: "/api/research/retry-all-failed", Category: "research", Description: "Retry all failed research URLs through the TypeScript research router."},
				{Path: "/api/research/enqueue", Category: "research", Description: "Enqueue a research URL through the TypeScript research router."},
				{Path: "/api/pulse/events", Category: "observability", Description: "Read pulse event history through the TypeScript pulse router, with a local empty-state fallback when the event bus is unavailable."},
				{Path: "/api/pulse/status", Category: "observability", Description: "Read pulse system status through the TypeScript pulse router, with a local offline-state fallback when the router is unavailable."},
				{Path: "/api/pulse/providers", Category: "observability", Description: "Check local provider status, with a local Go provider availability fallback when the TypeScript pulse router is unavailable."},
				{Path: "/api/session-export/export", Category: "sessions", Description: "Export sessions through the TypeScript session export router."},
				{Path: "/api/session-export/import", Category: "sessions", Description: "Import sessions through the TypeScript session export router."},
				{Path: "/api/session-export/detect-format", Category: "sessions", Description: "Detect session export format, with a local Go JSON-shape fallback when the TypeScript session export router is unavailable."},
				{Path: "/api/session-export/formats", Category: "sessions", Description: "List known session export formats, with a local Go fallback when the TypeScript session export router is unavailable."},
				{Path: "/api/session-export/history", Category: "sessions", Description: "Read session export history, with a local empty-state fallback when the TypeScript session export router is unavailable."},
				{Path: "/api/browser/status", Category: "browser", Description: "Read browser status through the TypeScript browser router, with a local explicit-unavailable fallback when the router is unavailable."},
				{Path: "/api/browser/close-page", Category: "browser", Description: "Close a browser page through the TypeScript browser router."},
				{Path: "/api/browser/close-all", Category: "browser", Description: "Close all browser pages through the TypeScript browser router."},
				{Path: "/api/browser/search-history", Category: "browser", Description: "Search browser history through the TypeScript browser router."},
				{Path: "/api/browser/scrape", Category: "browser", Description: "Scrape the current browser page through the TypeScript browser router."},
				{Path: "/api/browser/screenshot", Category: "browser", Description: "Capture a browser screenshot through the TypeScript browser router."},
				{Path: "/api/browser/debug", Category: "browser", Description: "Issue browser debug actions through the TypeScript browser router."},
				{Path: "/api/browser/proxy-fetch", Category: "browser", Description: "Run a browser proxy fetch through the TypeScript browser router."},
				{Path: "/api/browser-extension/save-memory", Category: "ui", Description: "Save a browser-extension memory through the TypeScript browser extension router."},
				{Path: "/api/browser-extension/parse-dom", Category: "ui", Description: "Parse browser DOM content through the TypeScript browser extension router."},
				{Path: "/api/browser-extension/memories", Category: "ui", Description: "List browser-extension memories through the TypeScript browser extension router, with a local empty-state fallback when the browser memory store is unavailable."},
				{Path: "/api/browser-extension/delete-memory", Category: "ui", Description: "Delete a browser-extension memory through the TypeScript browser extension router."},
				{Path: "/api/browser-extension/stats", Category: "ui", Description: "Read browser-extension memory stats through the TypeScript browser extension router, with a local zero-state fallback when the browser memory store is unavailable."},
				{Path: "/api/open-webui/status", Category: "ui", Description: "Read Open WebUI status through the TypeScript OpenWebUI router, with local deterministic status defaults when the integration is unavailable."},
				{Path: "/api/open-webui/embed-url", Category: "ui", Description: "Read Open WebUI embed URL through the TypeScript OpenWebUI router, with a local environment-backed fallback when the integration is unavailable."},
				{Path: "/api/code-mode/status", Category: "ui", Description: "Read Code Mode status through the TypeScript code mode router, with a local zero-state fallback when Code Mode is unavailable."},
				{Path: "/api/code-mode/enable", Category: "ui", Description: "Enable Code Mode through the TypeScript code mode router."},
				{Path: "/api/code-mode/disable", Category: "ui", Description: "Disable Code Mode through the TypeScript code mode router."},
				{Path: "/api/code-mode/execute", Category: "ui", Description: "Execute Code Mode code through the TypeScript code mode router."},
				{Path: "/api/submodules", Category: "ui", Description: "List submodules through the TypeScript submodule router, with a local Go .gitmodules fallback when the router is unavailable."},
				{Path: "/api/submodules/update-all", Category: "ui", Description: "Update all submodules through the TypeScript submodule router."},
				{Path: "/api/submodules/install-dependencies", Category: "ui", Description: "Install submodule dependencies through the TypeScript submodule router."},
				{Path: "/api/submodules/build", Category: "ui", Description: "Build a submodule through the TypeScript submodule router."},
				{Path: "/api/submodules/enable", Category: "ui", Description: "Enable a submodule through the TypeScript submodule router."},
				{Path: "/api/submodules/capabilities", Category: "ui", Description: "Read submodule capabilities through the TypeScript submodule router, with a local Go filesystem fallback when the router is unavailable."},
				{Path: "/api/suggestions", Category: "ui", Description: "List suggestions through the TypeScript suggestions router, with a local empty-state fallback when suggestions are unavailable."},
				{Path: "/api/suggestions/resolve", Category: "ui", Description: "Resolve a suggestion through the TypeScript suggestions router."},
				{Path: "/api/suggestions/clear", Category: "ui", Description: "Clear suggestions through the TypeScript suggestions router."},
				{Path: "/api/plan/mode", Category: "ui", Description: "Read or update plan mode through the TypeScript plan router, with a local default-mode fallback when the runtime is unavailable."},
				{Path: "/api/plan/diffs", Category: "ui", Description: "List pending plan diffs through the TypeScript plan router, with a local sandbox-file fallback when the runtime is unavailable."},
				{Path: "/api/plan/approve-diff", Category: "ui", Description: "Approve a plan diff through the TypeScript plan router."},
				{Path: "/api/plan/reject-diff", Category: "ui", Description: "Reject a plan diff through the TypeScript plan router."},
				{Path: "/api/plan/apply-all", Category: "ui", Description: "Apply approved plan diffs through the TypeScript plan router."},
				{Path: "/api/plan/summary", Category: "ui", Description: "Read plan sandbox summary through the TypeScript plan router, with a local sandbox-file fallback when the runtime is unavailable."},
				{Path: "/api/plan/checkpoints", Category: "ui", Description: "List plan checkpoints through the TypeScript plan router, with a local sandbox-file fallback when the runtime is unavailable."},
				{Path: "/api/plan/create-checkpoint", Category: "ui", Description: "Create a plan checkpoint through the TypeScript plan router."},
				{Path: "/api/plan/rollback", Category: "ui", Description: "Rollback a plan checkpoint through the TypeScript plan router."},
				{Path: "/api/plan/clear", Category: "ui", Description: "Clear plan sandbox state through the TypeScript plan router."},
				{Path: "/api/knowledge/graph", Category: "knowledge", Description: "Read the knowledge graph through the TypeScript knowledge router, with an explicit empty-graph fallback when graph data is unavailable."},
				{Path: "/api/knowledge/stats", Category: "knowledge", Description: "Read knowledge stats through the TypeScript knowledge router, with a local Go memory-context fallback when the router is unavailable."},
				{Path: "/api/knowledge/ingest", Category: "knowledge", Description: "Ingest a knowledge URL through the TypeScript knowledge router."},
				{Path: "/api/knowledge/resources", Category: "knowledge", Description: "Read knowledge resources through the TypeScript knowledge router, with a local Go resources.json fallback when the router is unavailable."},
				{Path: "/api/rag/file", Category: "knowledge", Description: "Ingest a file into RAG through the TypeScript RAG router."},
				{Path: "/api/rag/text", Category: "knowledge", Description: "Ingest text into RAG through the TypeScript RAG router."},
				{Path: "/api/directory", Category: "knowledge", Description: "List unified directory items through the TypeScript unified directory router."},
				{Path: "/api/directory/stats", Category: "knowledge", Description: "Read unified directory stats through the TypeScript unified directory router."},
				{Path: "/api/tool-chains/aliases", Category: "tools", Description: "List tool aliases through the TypeScript tool chaining router, with a local empty-state fallback when aliases are unavailable."},
				{Path: "/api/tool-chains/aliases/create", Category: "tools", Description: "Create a tool alias through the TypeScript tool chaining router."},
				{Path: "/api/tool-chains/aliases/remove", Category: "tools", Description: "Remove a tool alias through the TypeScript tool chaining router."},
				{Path: "/api/tool-chains/aliases/resolve", Category: "tools", Description: "Resolve a tool alias through the TypeScript tool chaining router, with a local unresolved fallback when aliases are unavailable."},
				{Path: "/api/tool-chains", Category: "tools", Description: "List tool chains through the TypeScript tool chaining router, with a local empty-state fallback when chains are unavailable."},
				{Path: "/api/tool-chains/get", Category: "tools", Description: "Read a tool chain through the TypeScript tool chaining router."},
				{Path: "/api/tool-chains/create", Category: "tools", Description: "Create a tool chain through the TypeScript tool chaining router."},
				{Path: "/api/tool-chains/execute", Category: "tools", Description: "Execute a tool chain through the TypeScript tool chaining router."},
				{Path: "/api/tool-chains/delete", Category: "tools", Description: "Delete a tool chain through the TypeScript tool chaining router."},
				{Path: "/api/tool-chains/lazy", Category: "tools", Description: "Read lazy tool states through the TypeScript tool chaining router, with a local empty-state fallback when no lazy tools are registered."},
				{Path: "/api/tool-chains/lazy/register", Category: "tools", Description: "Register a lazy tool through the TypeScript tool chaining router."},
				{Path: "/api/tool-chains/lazy/mark-loaded", Category: "tools", Description: "Mark a lazy tool as loaded through the TypeScript tool chaining router."},
				{Path: "/api/browser-controls/scrape", Category: "browser", Description: "Scrape a page through the TypeScript browser controls router."},
				{Path: "/api/browser-controls/history/push", Category: "browser", Description: "Push browser history through the TypeScript browser controls router."},
				{Path: "/api/browser-controls/history/query", Category: "browser", Description: "Query browser history through the TypeScript browser controls router."},
				{Path: "/api/browser-controls/logs/push", Category: "browser", Description: "Push browser console logs through the TypeScript browser controls router."},
				{Path: "/api/browser-controls/logs/query", Category: "browser", Description: "Query browser console logs through the TypeScript browser controls router."},
				{Path: "/api/browser-controls/stats", Category: "browser", Description: "Read browser controls stats through the TypeScript browser controls router."},
				{Path: "/api/swarm/start", Category: "orchestration", Description: "Start a TypeScript swarm orchestration mission."},
				{Path: "/api/swarm/resume", Category: "orchestration", Description: "Resume a TypeScript swarm mission."},
				{Path: "/api/swarm/approve-task", Category: "orchestration", Description: "Approve or reject a swarm task through the TypeScript swarm router."},
				{Path: "/api/swarm/decompose-task", Category: "orchestration", Description: "Decompose a swarm task through the TypeScript swarm router."},
				{Path: "/api/swarm/update-task-priority", Category: "orchestration", Description: "Update swarm task priority through the TypeScript swarm router."},
				{Path: "/api/swarm/debate", Category: "orchestration", Description: "Run a multi-model swarm debate through the TypeScript swarm router."},
				{Path: "/api/swarm/consensus", Category: "orchestration", Description: "Seek multi-model consensus through the TypeScript swarm router."},
				{Path: "/api/swarm/missions", Category: "orchestration", Description: "List swarm mission history through the TypeScript swarm router."},
				{Path: "/api/swarm/risk/summary", Category: "orchestration", Description: "Read swarm mission risk summary through the TypeScript swarm router."},
				{Path: "/api/swarm/risk/rows", Category: "orchestration", Description: "Read swarm mission risk rows through the TypeScript swarm router."},
				{Path: "/api/swarm/risk/facets", Category: "orchestration", Description: "Read swarm mission risk facets through the TypeScript swarm router."},
				{Path: "/api/swarm/mesh-capabilities", Category: "orchestration", Description: "Read swarm mesh capabilities through the TypeScript swarm router."},
				{Path: "/api/swarm/direct-message", Category: "orchestration", Description: "Send a direct mesh message through the TypeScript swarm router."},
				{Path: "/api/cli/tools", Category: "cli", Description: "Detected local CLI tools and versions."},
				{Path: "/api/cli/harnesses", Category: "cli", Description: "Harness registry metadata and install visibility."},
				{Path: "/api/cli/summary", Category: "cli", Description: "Compact CLI and harness readiness summary."},
				{Path: "/api/memory/borg-memory/status", Category: "memory", Description: "Read-only sectioned-memory status snapshot."},
				{Path: "/api/import/roots", Category: "imports", Description: "Explicit import discovery roots and whether they exist."},
				{Path: "/api/import/sources", Category: "imports", Description: "Discovered import artifacts from explicit roots."},
				{Path: "/api/import/validate", Category: "imports", Description: "Validation summary for a single import artifact path."},
				{Path: "/api/import/candidates", Category: "imports", Description: "Validated import candidates with metadata."},
				{Path: "/api/import/manifest", Category: "imports", Description: "Structured manifest of validated import candidates."},
				{Path: "/api/import/summary", Category: "imports", Description: "Aggregate summary of validated import candidates."},
				{Path: "/api/runtime/locks", Category: "runtime", Description: "Visibility into main borg and sidecar lock files."},
				{Path: "/api/runtime/status", Category: "runtime", Description: "Top-level runtime summary across CLI, imports, providers, memory, and sessions."},
				{Path: "/api/runtime/imported-instructions", Category: "runtime", Description: "Read-only bridge to imported instructions generated by the main fork."},
				{Path: "/api/startup/status", Category: "runtime", Description: "Truthful Go-sidecar startup readiness snapshot, including upstream control-plane dependency state."},
				{Path: "/api/mesh/status", Category: "mesh", Description: "Native Go mesh node id plus current known peer count."},
				{Path: "/api/mesh/peers", Category: "mesh", Description: "Known mesh peers discovered from the Go mesh visibility layer."},
				{Path: "/api/mesh/capabilities", Category: "mesh", Description: "Combined capability map for the Go node plus upstream-discovered peers."},
				{Path: "/api/mesh/query-capabilities", Category: "mesh", Description: "Detailed capability lookup for a specific mesh node, with upstream fallback when available."},
				{Path: "/api/mesh/find-peer", Category: "mesh", Description: "Find the first known peer whose advertised capabilities match a required capability set."},
				{Path: "/api/mesh/broadcast", Category: "mesh", Description: "Broadcast a mesh message through the TypeScript mesh router."},
			},
		},
	})
}

func (s *Server) handleConfigStatus(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    config.Snapshot(s.cfg),
	})
}

func (s *Server) handleProviderStatus(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    providers.Snapshot(),
	})
}

func (s *Server) handleProviderCatalog(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    providers.Catalog(providers.Snapshot()),
	})
}

func (s *Server) handleProviderSummary(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    providers.BuildSummary(providers.Snapshot()),
	})
}

func (s *Server) handleRoutingSummary(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    providers.BuildRoutingSummary(providers.Snapshot()),
	})
}

func (s *Server) handleSessions(w http.ResponseWriter, _ *http.Request) {
	sessions, err := s.discoveredSessions()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   "failed to discover sessions: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    sessions,
	})
}

func (s *Server) handleSessionSummary(w http.ResponseWriter, _ *http.Request) {
	sessions, err := s.discoveredSessions()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   "failed to summarize sessions: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    summarizeSessions(sessions),
	})
}

func (s *Server) handleSupervisorSessionCatalog(w http.ResponseWriter, r *http.Request) {
	s.handleSessionBridgeCall(w, r, http.MethodGet, "session.catalog", nil)
}

func (s *Server) handleImportedSessionList(w http.ResponseWriter, r *http.Request) {
	limit := strings.TrimSpace(r.URL.Query().Get("limit"))
	parsedLimit := 50
	var err error
	if limit != "" {
		parsedLimit, err = strconv.Atoi(limit)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{
				"success": false,
				"error":   "invalid limit query parameter",
			})
			return
		}
	}
	if parsedLimit <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "invalid limit query parameter",
		})
		return
	}

	var importedSessions []ImportedSessionRecord
	upstreamBase, bridgeErr := s.callUpstreamJSON(r.Context(), "session.importedList", map[string]any{"limit": parsedLimit}, &importedSessions)
	if bridgeErr == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    importedSessions,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "session.importedList",
			},
		})
		return
	}

	if archivedRecords, archiveErr := s.loadArchivedImportedSessionRecords(); archiveErr == nil && len(archivedRecords) > 0 {
		if parsedLimit < len(archivedRecords) {
			archivedRecords = archivedRecords[:parsedLimit]
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    archivedRecords,
			"bridge": map[string]any{
				"fallback":  "go-sessionimport",
				"procedure": "session.importedList",
				"reason":    "upstream unavailable; using archived imported session records",
			},
		})
		return
	}

	candidates, scanErr := s.scanValidatedImportSources()
	if scanErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"success": false,
			"error":   scanErr.Error(),
			"detail":  scanErr.Error(),
		})
		return
	}
	fallbackSessions := s.importedSessionFallbackRecords(candidates)
	if parsedLimit < len(fallbackSessions) {
		fallbackSessions = fallbackSessions[:parsedLimit]
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    fallbackSessions,
		"bridge": map[string]any{
			"fallback":  "go-sessionimport",
			"procedure": "session.importedList",
			"reason":    "upstream unavailable; using scan-only imported session records",
		},
	})
}

func (s *Server) handleImportedSessionGet(w http.ResponseWriter, r *http.Request) {
	importedID := strings.TrimSpace(r.URL.Query().Get("id"))
	if importedID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "missing id query parameter",
		})
		return
	}

	var importedSession *ImportedSessionRecord
	upstreamBase, bridgeErr := s.callUpstreamJSON(r.Context(), "session.importedGet", map[string]any{"id": importedID}, &importedSession)
	if bridgeErr == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    importedSession,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "session.importedGet",
			},
		})
		return
	}

	if archivedRecords, archiveErr := s.loadArchivedImportedSessionRecords(); archiveErr == nil && len(archivedRecords) > 0 {
		for _, record := range archivedRecords {
			if record.ID != importedID {
				continue
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"success": true,
				"data":    record,
				"bridge": map[string]any{
					"fallback":  "go-sessionimport",
					"procedure": "session.importedGet",
					"reason":    "upstream unavailable; using archived imported session record",
				},
			})
			return
		}
	}

	candidates, scanErr := s.scanValidatedImportSources()
	if scanErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"success": false,
			"error":   scanErr.Error(),
			"detail":  scanErr.Error(),
		})
		return
	}
	for _, record := range s.importedSessionFallbackRecords(candidates) {
		if record.ID != importedID {
			continue
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    record,
			"bridge": map[string]any{
				"fallback":  "go-sessionimport",
				"procedure": "session.importedGet",
				"reason":    "upstream unavailable; using scan-only imported session record",
			},
		})
		return
	}

	writeJSON(w, http.StatusServiceUnavailable, map[string]any{
		"success": false,
		"error":   "imported session unavailable",
		"detail":  "upstream unavailable; imported session not present in archived or scan-only fallback records",
		"bridge": map[string]any{
			"fallback":  "go-sessionimport",
			"procedure": "session.importedGet",
			"reason":    "upstream unavailable; imported session not present in archived or scan-only fallback records",
		},
	})
}

func (s *Server) handleImportedSessionScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{
			"success": false,
			"error":   "method not allowed",
		})
		return
	}

	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && err.Error() != "EOF" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "invalid JSON body",
		})
		return
	}

	var summary struct {
		DiscoveredCount    int      `json:"discoveredCount"`
		ImportedCount      int      `json:"importedCount"`
		SkippedCount       int      `json:"skippedCount"`
		StoredMemoryCount  int      `json:"storedMemoryCount"`
		InstructionDocPath *string  `json:"instructionDocPath"`
		Tools              []string `json:"tools"`
	}
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "session.importedScan", payload, &summary)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    summary,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "session.importedScan",
			},
		})
		return
	}

	candidates, scanErr := s.scanValidatedImportSources()
	if scanErr != nil {
		if archivedRecords, archiveErr := s.loadArchivedImportedSessionRecords(); archiveErr == nil && len(archivedRecords) > 0 {
			fallbackSummary := s.archivedImportedSessionScanSummary(archivedRecords)
			writeJSON(w, http.StatusOK, map[string]any{
				"success": true,
				"data":    fallbackSummary,
				"bridge": map[string]any{
					"fallback":  "go-sessionimport",
					"procedure": "session.importedScan",
					"reason":    "upstream unavailable; using archived imported sessions because validated scan fallback failed: " + scanErr.Error(),
				},
			})
			return
		}
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"success": false,
			"error":   scanErr.Error(),
			"detail":  scanErr.Error(),
		})
		return
	}

	if archivedRecords, archiveErr := s.loadArchivedImportedSessionRecords(); archiveErr == nil && len(archivedRecords) > 0 {
		fallbackSummary := s.mergedImportedSessionScanSummary(archivedRecords, candidates)
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    fallbackSummary,
			"bridge": map[string]any{
				"fallback":  "go-sessionimport",
				"procedure": "session.importedScan",
				"reason":    "upstream unavailable; merged persisted imported sessions with validated scan candidates",
			},
		})
		return
	}

	discoveredCount := len(candidates)
	importedCount := 0
	skippedCount := 0
	toolsSet := make(map[string]struct{})
	for _, candidate := range candidates {
		if candidate.Valid {
			skippedCount++
		}
		if candidate.SourceTool != "" {
			toolsSet[candidate.SourceTool] = struct{}{}
		}
	}
	tools := make([]string, 0, len(toolsSet))
	for tool := range toolsSet {
		tools = append(tools, tool)
	}
	sort.Strings(tools)

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"discoveredCount":    discoveredCount,
			"importedCount":      importedCount,
			"skippedCount":       skippedCount,
			"storedMemoryCount":  0,
			"instructionDocPath": s.importedInstructionDocPath(),
			"tools":              tools,
		},
		"bridge": map[string]any{
			"fallback":  "go-sessionimport",
			"procedure": "session.importedScan",
			"reason":    "upstream unavailable; using scan-only imported session discovery summary",
		},
	})
}

func (s *Server) handleImportedSessionInstructionDocs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{
			"success": false,
			"error":   "method not allowed",
		})
		return
	}

	var docs []map[string]any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "session.importedInstructionDocs", nil, &docs)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    docs,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "session.importedInstructionDocs",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": func() []map[string]any {
			document := interop.ReadImportedInstructions(s.cfg.ImportedInstructionsPath())
			if !document.Available {
				return []map[string]any{}
			}
			return []map[string]any{
				{
					"path":       document.Path,
					"name":       filepath.Base(document.Path),
					"modifiedAt": document.ModifiedAt,
					"size":       document.Size,
				},
			}
		}(),
		"bridge": map[string]any{
			"fallback":  "go-sessionimport",
			"procedure": "session.importedInstructionDocs",
			"reason":    "upstream unavailable; using workspace imported instruction documents",
		},
	})
}

func (s *Server) handleImportedSessionMaintenanceStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{
			"success": false,
			"error":   "method not allowed",
		})
		return
	}

	var stats ImportedSessionMaintenanceStats
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "session.importedMaintenanceStats", nil, &stats)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    stats,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "session.importedMaintenanceStats",
			},
		})
		return
	}

	if archivedRecords, archiveErr := s.loadArchivedImportedSessionRecords(); archiveErr == nil && len(archivedRecords) > 0 {
		fallbackStats := archivedImportedSessionMaintenanceStats(archivedRecords)
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    fallbackStats,
			"bridge": map[string]any{
				"fallback":  "go-sessionimport",
				"procedure": "session.importedMaintenanceStats",
				"reason":    "upstream unavailable; using archived imported session maintenance stats",
			},
		})
		return
	}

	candidates, scanErr := s.scanValidatedImportSources()
	if scanErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"success": false,
			"error":   scanErr.Error(),
			"detail":  scanErr.Error(),
		})
		return
	}

	fallbackStats := ImportedSessionMaintenanceStats{
		TotalSessions:                len(candidates),
		InlineTranscriptCount:        0,
		ArchivedTranscriptCount:      0,
		MissingRetentionSummaryCount: 0,
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    fallbackStats,
		"bridge": map[string]any{
			"fallback":  "go-sessionimport",
			"procedure": "session.importedMaintenanceStats",
			"reason":    "upstream unavailable; using scan-only imported session maintenance stats",
		},
	})
}

func (s *Server) handleMCPConfiguredServers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{
			"success": false,
			"error":   "method not allowed",
		})
		return
	}

	var servers []map[string]any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "mcpServers.list", nil, &servers)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    servers,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "mcpServers.list",
			},
		})
		return
	}

	fallbackServers, fallbackErr := s.localConfiguredMCPServersFromDB()
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
		"data":    fallbackServers,
		"bridge": map[string]any{
			"fallback":  "go-local-mcp-db",
			"procedure": "mcpServers.list",
			"reason":    "upstream unavailable; using local MCP server definitions from metamcp.db with JSONC metadata overlay",
		},
	})
}

func (s *Server) handleMCPConfiguredServerGet(w http.ResponseWriter, r *http.Request) {
	uuid := strings.TrimSpace(r.URL.Query().Get("uuid"))
	if uuid == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "missing uuid query parameter",
		})
		return
	}

	var server map[string]any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "mcpServers.get", map[string]any{"uuid": uuid}, &server)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    server,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "mcpServers.get",
			},
		})
		return
	}

	fallbackServer, fallbackErr := s.localConfiguredMCPServerFromDB(uuid)
	if fallbackErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"success": false,
			"error":   fallbackErr.Error(),
			"detail":  fallbackErr.Error(),
		})
		return
	}
	if fallbackServer != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    fallbackServer,
			"bridge": map[string]any{
				"fallback":  "go-local-mcp-db",
				"procedure": "mcpServers.get",
				"reason":    "upstream unavailable; using local MCP server definition from metamcp.db with JSONC metadata overlay",
			},
		})
		return
	}

	writeJSON(w, http.StatusServiceUnavailable, map[string]any{
		"success": false,
		"error":   "configured MCP server unavailable",
		"detail":  "upstream unavailable; configured MCP server not present in local metamcp.db fallback",
		"bridge": map[string]any{
			"fallback":  "go-local-mcp-db",
			"procedure": "mcpServers.get",
			"reason":    "upstream unavailable; configured MCP server not present in local metamcp.db fallback",
		},
	})
}

func (s *Server) handleMCPConfiguredServerCreate(w http.ResponseWriter, r *http.Request) {
	s.handleConfiguredServerMutation(w, r, "mcpServers.create", func(payload map[string]any) (any, error) {
		return s.localCreateConfiguredServer(payload)
	})
}

func (s *Server) handleMCPConfiguredServerUpdate(w http.ResponseWriter, r *http.Request) {
	s.handleConfiguredServerMutation(w, r, "mcpServers.update", func(payload map[string]any) (any, error) {
		return s.localUpdateConfiguredServer(payload)
	})
}

func (s *Server) handleMCPConfiguredServerDelete(w http.ResponseWriter, r *http.Request) {
	s.handleConfiguredServerMutation(w, r, "mcpServers.delete", func(payload map[string]any) (any, error) {
		return s.localDeleteConfiguredServer(payload)
	})
}

func (s *Server) handleMCPConfiguredServerBulkImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{
			"success": false,
			"error":   "method not allowed",
		})
		return
	}

	var payload []map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "invalid JSON body",
		})
		return
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "mcpServers.bulkImport", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "mcpServers.bulkImport",
			},
		})
		return
	}

	fallbackResult, fallbackErr := s.localBulkImportConfiguredServers(payload)
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
		"data":    fallbackResult,
		"bridge": map[string]any{
			"fallback":  "go-local-jsonc",
			"procedure": "mcpServers.bulkImport",
			"reason":    "upstream unavailable; using local JSONC bulk import",
		},
	})
}

func (s *Server) handleMCPConfiguredServerReloadMetadata(w http.ResponseWriter, r *http.Request) {
	s.handleConfiguredServerMutation(w, r, "mcpServers.reloadMetadata", func(payload map[string]any) (any, error) {
		return s.localReloadConfiguredServerMetadata(payload)
	})
}

func (s *Server) handleMCPConfiguredServerClearMetadataCache(w http.ResponseWriter, r *http.Request) {
	s.handleConfiguredServerMutation(w, r, "mcpServers.clearMetadataCache", func(payload map[string]any) (any, error) {
		return s.localClearConfiguredServerMetadata(payload)
	})
}

func (s *Server) handleMCPRegistrySnapshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{
			"success": false,
			"error":   "method not allowed",
		})
		return
	}

	var snapshot []map[string]any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "mcpServers.registrySnapshot", nil, &snapshot)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    snapshot,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "mcpServers.registrySnapshot",
			},
		})
		return
	}

	fallbackSnapshot, fallbackErr := s.localMCPRegistrySnapshot()
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
		"data":    fallbackSnapshot,
		"bridge": map[string]any{
			"fallback":  "go-master-index",
			"procedure": "mcpServers.registrySnapshot",
			"reason":    "upstream unavailable; using local master index registry snapshot",
		},
	})
}

func (s *Server) handleMCPSyncTargets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{
			"success": false,
			"error":   "method not allowed",
		})
		return
	}

	var targets []map[string]any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "mcpServers.syncTargets", nil, &targets)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    targets,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "mcpServers.syncTargets",
			},
		})
		return
	}

	fallbackTargets, fallbackErr := s.localMCPSyncTargets()
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
		"data":    fallbackTargets,
		"bridge": map[string]any{
			"fallback":  "go-local-jsonc",
			"procedure": "mcpServers.syncTargets",
			"reason":    "upstream unavailable; using local JSONC sync targets",
		},
	})
}

func (s *Server) handleMCPExportClientConfig(w http.ResponseWriter, r *http.Request) {
	client := strings.TrimSpace(r.URL.Query().Get("client"))
	if client == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "missing client query parameter",
		})
		return
	}
	payload := map[string]any{"client": client}
	if path := strings.TrimSpace(r.URL.Query().Get("path")); path != "" {
		payload["path"] = path
	}

	var preview map[string]any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "mcpServers.exportClientConfig", payload, &preview)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    preview,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "mcpServers.exportClientConfig",
			},
		})
		return
	}

	overridePath, _ := payload["path"].(string)
	fallbackPreview, fallbackErr := s.localMCPExportClientConfig(client, overridePath)
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
		"data":    fallbackPreview,
		"bridge": map[string]any{
			"fallback":  "go-local-jsonc",
			"procedure": "mcpServers.exportClientConfig",
			"reason":    "upstream unavailable; using local JSONC client config export preview",
		},
	})
}

func (s *Server) handleMCPSyncClientConfig(w http.ResponseWriter, r *http.Request) {
	s.handleConfiguredServerMutation(w, r, "mcpServers.syncClientConfig", func(payload map[string]any) (any, error) {
		client, _ := payload["client"].(string)
		overridePath, _ := payload["path"].(string)
		preview, err := s.localMCPExportClientConfig(client, overridePath)
		if err != nil {
			return nil, err
		}
		targetPath, _ := preview["targetPath"].(string)
		jsonText, _ := preview["json"].(string)
		if strings.TrimSpace(targetPath) == "" {
			return nil, errors.New("missing client target path")
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(targetPath, []byte(jsonText), 0o644); err != nil {
			return nil, err
		}
		preview["written"] = true
		return preview, nil
	})
}

func (s *Server) handleMCPCallTool(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{
			"success": false,
			"error":   "method not allowed",
		})
		return
	}

	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "invalid JSON body",
		})
		return
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "mcp.callTool", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "mcp.callTool",
			},
		})
		return
	}

	fallbackResult, fallbackErr := s.localCallMCPMetaTool(r, payload)
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
		"data":    fallbackResult,
		"bridge": map[string]any{
			"fallback":  "go-local-mcp",
			"procedure": "mcp.callTool",
			"reason":    "upstream unavailable; using local MCP meta-tool execution",
		},
	})
}

func (s *Server) handleMCPAutoCallTool(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{
			"success": false,
			"error":   "method not allowed",
		})
		return
	}

	var rawArgs map[string]any
	if err := json.NewDecoder(r.Body).Decode(&rawArgs); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "invalid JSON body",
		})
		return
	}

	args := normalizeAutoCallArgs(rawArgs)
	payload := map[string]any{
		"name": "auto_call_tool",
		"args": args,
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "mcp.callTool", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "mcp.callTool",
			},
		})
		return
	}

	fallbackResult, fallbackErr := s.localCallMCPMetaTool(r, payload)
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
		"data":    fallbackResult,
		"bridge": map[string]any{
			"fallback":  "go-local-mcp",
			"procedure": "mcp.callTool",
			"reason":    "upstream unavailable; using local auto-call meta-tool execution",
		},
	})
}

func normalizeAutoCallArgs(input map[string]any) map[string]any {
	args := map[string]any{}
	for key, value := range input {
		args[key] = value
	}

	readString := func(keys ...string) string {
		for _, key := range keys {
			value, ok := input[key]
			if !ok {
				continue
			}
			text, ok := value.(string)
			if !ok {
				continue
			}
			trimmed := strings.TrimSpace(text)
			if trimmed != "" {
				return trimmed
			}
		}
		return ""
	}

	objective := readString("objective", "query", "task", "goal", "prompt", "instruction")
	if objective != "" {
		args["objective"] = objective
	}

	if contextValue := readString("context", "details", "description", "notes"); contextValue != "" {
		args["context"] = contextValue
	} else {
		contextParts := make([]string, 0, 4)
		for _, key := range []string{"path", "file", "filePath", "selection", "cwd"} {
			if value := readString(key); value != "" {
				contextParts = append(contextParts, key+": "+value)
			}
		}
		if len(contextParts) > 0 {
			args["context"] = strings.Join(contextParts, "; ")
		}
	}

	return args
}

func (s *Server) handleMCPToolAdvertisements(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("query"))
	goal := strings.TrimSpace(r.URL.Query().Get("goal"))
	objective := strings.TrimSpace(r.URL.Query().Get("objective"))
	if query == "" {
		query = strings.TrimSpace(strings.Join([]string{objective, goal}, " "))
	}

	limit := 8
	if rawLimit := strings.TrimSpace(r.URL.Query().Get("limit")); rawLimit != "" {
		if parsed, err := strconv.Atoi(rawLimit); err == nil && parsed > 0 && parsed <= 32 {
			limit = parsed
		}
	}

	toolSuggestions, err := s.buildToolSuggestionSnapshotWithLimit(r, query, limit)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": "failed to build tool advertisement snapshot: " + err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"recommendedTools": toolSuggestions.RecommendedTools,
			"relatedTools":     toolSuggestions.RelatedTools,
		},
		"bridge": map[string]any{
			"recommendedTools": toolSuggestions.Bridge["recommendedTools"],
			"relatedTools":     toolSuggestions.Bridge["relatedTools"],
		},
	})
}

func (s *Server) handleMCPToolSchema(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{
			"success": false,
			"error":   "method not allowed",
		})
		return
	}

	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "invalid JSON body",
		})
		return
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "mcp.getToolSchema", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "mcp.getToolSchema",
			},
		})
		return
	}

	fallbackResult, fallbackErr := localFallbackToolSchema(payload)
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
		"data":    fallbackResult,
		"bridge": map[string]any{
			"fallback":  "go-local-mcp",
			"procedure": "mcp.getToolSchema",
			"reason":    "upstream unavailable; using local MCP tool schema fallback",
		},
	})
}

func (s *Server) handleMCPToolPreferences(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		var preferences map[string]any
		upstreamBase, err := s.callUpstreamJSON(r.Context(), "mcp.getToolPreferences", nil, &preferences)
		if err == nil {
			writeJSON(w, http.StatusOK, map[string]any{
				"success": true,
				"data":    preferences,
				"bridge": map[string]any{
					"upstreamBase": upstreamBase,
					"procedure":    "mcp.getToolPreferences",
				},
			})
			return
		}

		fallbackPreferences, fallbackErr := s.localToolPreferences()
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
			"data":    fallbackPreferences,
			"bridge": map[string]any{
				"fallback":  "go-local-jsonc",
				"procedure": "mcp.getToolPreferences",
				"reason":    "upstream unavailable; using local JSONC tool preferences",
			},
		})
	case http.MethodPost:
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{
				"success": false,
				"error":   "invalid JSON body",
			})
			return
		}

		var result map[string]any
		upstreamBase, err := s.callUpstreamJSON(r.Context(), "mcp.setToolPreferences", payload, &result)
		if err == nil {
			writeJSON(w, http.StatusOK, map[string]any{
				"success": true,
				"data":    result,
				"bridge": map[string]any{
					"upstreamBase": upstreamBase,
					"procedure":    "mcp.setToolPreferences",
				},
			})
			return
		}

		fallbackPreferences, fallbackErr := s.saveLocalToolPreferences(payload)
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
			"data":    fallbackPreferences,
			"bridge": map[string]any{
				"fallback":  "go-local-jsonc",
				"procedure": "mcp.setToolPreferences",
				"reason":    "upstream unavailable; saving tool preferences to local JSONC",
			},
		})
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{
			"success": false,
			"error":   "method not allowed",
		})
	}
}

func (s *Server) handleMCPTraffic(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "mcp.traffic", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "mcp.traffic",
			},
		})
		return
	}

	writeJSON(w, http.StatusServiceUnavailable, map[string]any{
		"success": false,
		"error":   "MCP traffic history is unavailable: upstream MCP router is unavailable and no local traffic history is persisted.",
		"data":    []map[string]any{},
		"bridge": map[string]any{
			"fallback":  "go-local-mcp",
			"procedure": "mcp.traffic",
			"reason":    "upstream unavailable; using local empty MCP traffic history",
		},
	})
}

func (s *Server) handleMCPToolSelectionTelemetry(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "mcp.getToolSelectionTelemetry", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "mcp.getToolSelectionTelemetry",
			},
		})
		return
	}

	writeJSON(w, http.StatusServiceUnavailable, map[string]any{
		"success": false,
		"error":   "Tool-selection telemetry is unavailable: upstream MCP router is unavailable and no local telemetry history is persisted.",
		"data":    []map[string]any{},
		"bridge": map[string]any{
			"fallback":  "go-local-mcp",
			"procedure": "mcp.getToolSelectionTelemetry",
			"reason":    "upstream unavailable; using local empty tool-selection telemetry",
		},
	})
}

func (s *Server) handleMCPClearToolSelectionTelemetry(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "mcp.clearToolSelectionTelemetry", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "mcp.clearToolSelectionTelemetry",
			},
		})
		return
	}

	writeJSON(w, http.StatusServiceUnavailable, map[string]any{
		"success": false,
		"error":   "Tool-selection telemetry clearing is unavailable: upstream MCP router is unavailable and no local telemetry history exists.",
		"data": map[string]any{
			"ok": true,
		},
		"bridge": map[string]any{
			"fallback":  "go-local-mcp",
			"procedure": "mcp.clearToolSelectionTelemetry",
			"reason":    "upstream unavailable; clearing local empty tool-selection telemetry",
		},
	})
}

func (s *Server) handleMCPServerTest(w http.ResponseWriter, r *http.Request) {
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "mcp.runServerTest", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "mcp.runServerTest",
			},
		})
		return
	}

	startedAt := time.Now()
	targetKind, _ := payload["targetKind"].(string)
	operation, _ := payload["operation"].(string)
	serverName, _ := payload["serverName"].(string)
	toolName, _ := payload["toolName"].(string)
	args, _ := payload["args"].(map[string]any)
	if args == nil {
		args = map[string]any{}
	}

	requestPayload := map[string]any{
		"jsonrpc": "2.0",
		"id":      "probe-go-local",
		"method":  operation,
		"params": func() map[string]any {
			if operation == "tools/call" {
				return map[string]any{"name": toolName, "arguments": args}
			}
			return map[string]any{}
		}(),
		"target": func() string {
			if targetKind == "router" {
				return "borg-router"
			}
			if strings.TrimSpace(serverName) != "" {
				return serverName
			}
			return "unknown-server"
		}(),
		"via": func() string {
			if targetKind == "router" {
				return "borg-router"
			}
			return "direct-downstream"
		}(),
	}

	buildFailure := func(summary string, payloadBody map[string]any) map[string]any {
		endedAt := time.Now()
		return map[string]any{
			"success": false,
			"target": map[string]any{
				"kind": targetKind,
				"displayName": func() string {
					if targetKind == "router" {
						return "borg router"
					}
					if strings.TrimSpace(serverName) != "" {
						return serverName
					}
					return "Unknown downstream server"
				}(),
				"serverName": nullableString(serverName),
				"via":        requestPayload["via"],
			},
			"operation": operation,
			"startedAt": startedAt.UnixMilli(),
			"endedAt":   endedAt.UnixMilli(),
			"latencyMs": endedAt.Sub(startedAt).Milliseconds(),
			"request":   requestPayload,
			"response": map[string]any{
				"summary": summary,
				"payload": payloadBody,
			},
			"trafficEvents": []map[string]any{},
		}
	}

	switch {
	case targetKind == "server" && strings.TrimSpace(serverName) == "":
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    buildFailure("Downstream probe requires a server name.", map[string]any{"error": "Downstream probe requires a server name."}),
			"bridge":  map[string]any{"fallback": "go-local-mcp", "procedure": "mcp.runServerTest", "reason": "upstream unavailable; validating probe request locally"},
		})
		return
	case operation == "tools/call" && strings.TrimSpace(toolName) == "":
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    buildFailure("Tool call probe requires a tool name.", map[string]any{"error": "Tool call probe requires a tool name."}),
			"bridge":  map[string]any{"fallback": "go-local-mcp", "procedure": "mcp.runServerTest", "reason": "upstream unavailable; validating probe request locally"},
		})
		return
	case targetKind == "router":
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    buildFailure("borg MCP router is not initialized.", map[string]any{"error": "borg MCP router is not initialized."}),
			"bridge":  map[string]any{"fallback": "go-local-mcp", "procedure": "mcp.runServerTest", "reason": "upstream unavailable; simulating router probe failure locally"},
		})
		return
	default:
		message := "Downstream server '" + serverName + "' is not currently connected."
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    buildFailure(message, map[string]any{"error": message}),
			"bridge":  map[string]any{"fallback": "go-local-mcp", "procedure": "mcp.runServerTest", "reason": "upstream unavailable; simulating downstream probe failure locally"},
		})
		return
	}
}

func (s *Server) handleMCPSetLifecycleModes(w http.ResponseWriter, r *http.Request) {
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	if value, ok := payload["lazySessionMode"].(bool); ok {
		s.lifecycleModes["lazySessionMode"] = value
	}
	if value, ok := payload["singleActiveServerMode"].(bool); ok {
		s.lifecycleModes["singleActiveServerMode"] = value
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "mcp.setLifecycleModes", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "mcp.setLifecycleModes",
			},
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"ok": true,
			"lifecycle": map[string]any{
				"lazySessionMode":        s.lifecycleModes["lazySessionMode"],
				"singleActiveServerMode": s.lifecycleModes["singleActiveServerMode"],
			},
		},
		"bridge": map[string]any{
			"fallback":  "go-local-mcp",
			"procedure": "mcp.setLifecycleModes",
			"reason":    "upstream unavailable; using local MCP lifecycle mode state",
		},
	})
}

func (s *Server) handleMCPAddServer(w http.ResponseWriter, r *http.Request) {
	s.handleConfiguredServerMutation(w, r, "mcp.addServer", func(payload map[string]any) (any, error) {
		result, err := s.localCreateConfiguredServer(payload)
		if err != nil {
			return nil, err
		}
		name, _ := payload["name"].(string)
		return map[string]any{
			"success": true,
			"name":    name,
			"server":  result,
		}, nil
	})
}

func (s *Server) handleMCPRemoveServer(w http.ResponseWriter, r *http.Request) {
	s.handleConfiguredServerMutation(w, r, "mcp.removeServer", func(payload map[string]any) (any, error) {
		if _, err := s.localDeleteConfiguredServer(payload); err != nil {
			return nil, err
		}
		return map[string]any{
			"success": true,
		}, nil
	})
}

func (s *Server) handleMCPJsoncConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		var editor map[string]any
		upstreamBase, err := s.callUpstreamJSON(r.Context(), "mcp.getJsoncEditor", nil, &editor)
		if err == nil {
			writeJSON(w, http.StatusOK, map[string]any{
				"success": true,
				"data":    editor,
				"bridge": map[string]any{
					"upstreamBase": upstreamBase,
					"procedure":    "mcp.getJsoncEditor",
				},
			})
			return
		}

		editor, fallbackErr := s.localMCPJsoncEditor()
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
			"data":    editor,
			"bridge": map[string]any{
				"fallback":  "go-local-jsonc",
				"procedure": "mcp.getJsoncEditor",
				"reason":    "upstream unavailable; using local MCP JSONC editor payload",
			},
		})
	case http.MethodPost:
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{
				"success": false,
				"error":   "invalid JSON body",
			})
			return
		}

		var result map[string]any
		upstreamBase, err := s.callUpstreamJSON(r.Context(), "mcp.saveJsoncEditor", payload, &result)
		if err == nil {
			writeJSON(w, http.StatusOK, map[string]any{
				"success": true,
				"data":    result,
				"bridge": map[string]any{
					"upstreamBase": upstreamBase,
					"procedure":    "mcp.saveJsoncEditor",
				},
			})
			return
		}

		content, _ := payload["content"].(string)
		if strings.TrimSpace(content) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{
				"success": false,
				"error":   "missing content",
			})
			return
		}
		if fallbackErr := s.saveLocalMCPJsonc(content); fallbackErr != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{
				"success": false,
				"error":   fallbackErr.Error(),
				"detail":  fallbackErr.Error(),
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data": map[string]any{
				"ok": true,
			},
			"bridge": map[string]any{
				"fallback":  "go-local-jsonc",
				"procedure": "mcp.saveJsoncEditor",
				"reason":    "upstream unavailable; saving MCP JSONC through local compatibility writer",
			},
		})
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{
			"success": false,
			"error":   "method not allowed",
		})
	}
}

func (s *Server) handleMCPWorkingSet(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "mcp.getWorkingSet", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "mcp.getWorkingSet",
			},
		})
		return
	}

	writeJSON(w, http.StatusServiceUnavailable, map[string]any{
		"success": false,
		"error":   "MCP working set is unavailable: upstream MCP router is unavailable and no local working set manager is initialized.",
		"data": map[string]any{
			"limits": map[string]any{
				"maxLoadedTools":          0,
				"maxHydratedSchemas":      0,
				"idleEvictionThresholdMs": 0,
			},
			"tools": []map[string]any{},
		},
		"bridge": map[string]any{
			"fallback":  "go-local-mcp",
			"procedure": "mcp.getWorkingSet",
			"reason":    "upstream unavailable; using local empty MCP working set",
		},
	})
}

func (s *Server) handleMCPWorkingSetEvictions(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "mcp.getWorkingSetEvictionHistory", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "mcp.getWorkingSetEvictionHistory",
			},
		})
		return
	}

	writeJSON(w, http.StatusServiceUnavailable, map[string]any{
		"success": false,
		"error":   "MCP eviction history is unavailable: upstream MCP router is unavailable and no local eviction history is persisted.",
		"data":    []map[string]any{},
		"bridge": map[string]any{
			"fallback":  "go-local-mcp",
			"procedure": "mcp.getWorkingSetEvictionHistory",
			"reason":    "upstream unavailable; using local empty MCP eviction history",
		},
	})
}

func (s *Server) handleMCPClearWorkingSetEvictions(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "mcp.clearWorkingSetEvictionHistory", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "mcp.clearWorkingSetEvictionHistory",
			},
		})
		return
	}

	writeJSON(w, http.StatusServiceUnavailable, map[string]any{
		"success": false,
		"error":   "MCP eviction history clearing is unavailable: upstream MCP router is unavailable and no local eviction history exists.",
		"data": map[string]any{
			"ok":      true,
			"message": "MCP server unavailable; eviction history already empty.",
		},
		"bridge": map[string]any{
			"fallback":  "go-local-mcp",
			"procedure": "mcp.clearWorkingSetEvictionHistory",
			"reason":    "upstream unavailable; clearing local empty MCP eviction history",
		},
	})
}

func (s *Server) handleMCPLoadTool(w http.ResponseWriter, r *http.Request) {
	s.handleMCPManualToolMutation(w, r, "mcp.loadTool")
}

func (s *Server) handleMCPUnloadTool(w http.ResponseWriter, r *http.Request) {
	s.handleMCPManualToolMutation(w, r, "mcp.unloadTool")
}

func (s *Server) handleMemoryContextSave(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "memory.saveContext")
}

func (s *Server) handleMemoryContextGet(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "missing id query parameter",
		})
		return
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "memory.getContext", map[string]any{"id": id}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "memory.getContext",
			},
		})
		return
	}

	contextRecord, found, localErr := s.localFindMemoryContext(id)
	if localErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error(), "detail": localErr.Error()})
		return
	}
	if found && strings.TrimSpace(stringValue(contextRecord["content"])) != "" {
		metadata, _ := contextRecord["metadata"].(map[string]any)
		responseMetadata := cloneMap(metadata)
		responseMetadata["title"] = stringValue(contextRecord["title"])
		responseMetadata["source"] = stringValue(contextRecord["source"])
		responseMetadata["createdAt"] = contextRecord["createdAt"]
		responseMetadata["chunks"] = contextRecord["chunks"]
		response := map[string]any{
			"id":       localMemoryContextID(contextRecord, 0),
			"content":  stringValue(contextRecord["content"]),
			"metadata": responseMetadata,
			"score":    1,
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    response,
			"bridge": map[string]any{
				"fallback":  "go-local-memory",
				"procedure": "memory.getContext",
				"reason":    "upstream unavailable; using local persisted memory context body",
			},
		})
		return
	}

	writeJSON(w, http.StatusServiceUnavailable, map[string]any{
		"success": false,
		"error":   "memory context unavailable",
		"detail":  "upstream unavailable; local memory context fallback has no persisted context body",
		"bridge": map[string]any{
			"fallback":  "go-local-memory",
			"procedure": "memory.getContext",
			"reason":    "upstream unavailable; local memory context fallback has no persisted context body",
		},
	})
}

func (s *Server) handleMemoryContextDelete(w http.ResponseWriter, r *http.Request) {
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "memory.deleteContext", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": result, "bridge": map[string]any{"upstreamBase": upstreamBase, "procedure": "memory.deleteContext"}})
		return
	}

	id := strings.TrimSpace(stringValue(payload["id"]))
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing id"})
		return
	}
	deleted, localErr := s.localDeleteMemoryContext(id)
	if localErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error(), "detail": localErr.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    map[string]any{"success": deleted},
		"bridge": map[string]any{
			"fallback":  "go-local-memory",
			"procedure": "memory.deleteContext",
			"reason":    "upstream unavailable; using local memory context registry deletion",
		},
	})
}

func (s *Server) handleMemoryAgentStats(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "memory.getAgentStats", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "memory.getAgentStats",
			},
		})
		return
	}

	stats, localErr := s.localAgentMemoryStats()
	if localErr != nil {
		stats = localAgentMemoryZeroStats()
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    stats,
		"bridge": map[string]any{
			"fallback":  "go-local-memory",
			"procedure": "memory.getAgentStats",
			"reason":    "upstream unavailable; using local persisted memory agent stats",
		},
	})
}

func (s *Server) handleMemoryAgentSearch(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("query"))
	if query == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "missing query parameter",
		})
		return
	}
	payload := map[string]any{"query": query}
	limit := 10
	if limitParam := strings.TrimSpace(r.URL.Query().Get("limit")); limitParam != "" {
		if parsed, err := strconv.Atoi(limitParam); err == nil {
			payload["limit"] = parsed
			if parsed > 0 {
				limit = parsed
			}
		}
	}
	memoryType := strings.TrimSpace(r.URL.Query().Get("type"))
	if memoryType != "" {
		payload["type"] = memoryType
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "memory.searchAgentMemory", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "memory.searchAgentMemory",
			},
		})
		return
	}

	records, localErr := s.localSearchAgentMemoryRecords(query, localAgentMemorySearchOptions{Limit: limit, Type: memoryType})
	if localErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error(), "detail": localErr.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    localAgentMemoryMaps(records),
		"bridge": map[string]any{
			"fallback":  "go-local-memory",
			"procedure": "memory.searchAgentMemory",
			"reason":    "upstream unavailable; using local persisted agent memory search",
		},
	})
}

func (s *Server) handleMemoryAddFact(w http.ResponseWriter, r *http.Request) {
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "memory.addFact", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": result, "bridge": map[string]any{"upstreamBase": upstreamBase, "procedure": "memory.addFact"}})
		return
	}

	memory, localErr := s.localAddFactMemory(payload)
	if localErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error(), "detail": localErr.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"success": true,
			"memory":  memory,
		},
		"bridge": map[string]any{
			"fallback":  "go-local-memory",
			"procedure": "memory.addFact",
			"reason":    "upstream unavailable; persisted memory fact locally",
		},
	})
}

func (s *Server) handleMemoryRecordObservation(w http.ResponseWriter, r *http.Request) {
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "memory.recordObservation", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": result, "bridge": map[string]any{"upstreamBase": upstreamBase, "procedure": "memory.recordObservation"}})
		return
	}

	memory, localErr := s.localRecordObservationMemory(payload)
	if localErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error(), "detail": localErr.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"success": true,
			"memory":  memory,
		},
		"bridge": map[string]any{
			"fallback":  "go-local-memory",
			"procedure": "memory.recordObservation",
			"reason":    "upstream unavailable; persisted structured observation locally",
		},
	})
}

func (s *Server) handleMemoryRecentObservations(w http.ResponseWriter, r *http.Request) {
	payload := map[string]any{"limit": 10}
	limit := 10
	if limitParam := strings.TrimSpace(r.URL.Query().Get("limit")); limitParam != "" {
		if parsed, err := strconv.Atoi(limitParam); err == nil {
			payload["limit"] = parsed
			if parsed > 0 {
				limit = parsed
			}
		}
	}
	namespace := strings.TrimSpace(r.URL.Query().Get("namespace"))
	if namespace != "" {
		payload["namespace"] = namespace
	}
	observationType := strings.TrimSpace(r.URL.Query().Get("type"))
	if observationType != "" {
		payload["type"] = observationType
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "memory.getRecentObservations", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "memory.getRecentObservations",
			},
		})
		return
	}

	records, localErr := s.localRecentObservations(limit, namespace, observationType)
	if localErr != nil {
		records = []map[string]any{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    records,
		"bridge": map[string]any{
			"fallback":  "go-local-memory",
			"procedure": "memory.getRecentObservations",
			"reason":    "upstream unavailable; using local persisted recent observations",
		},
	})
}

func (s *Server) handleMemorySearchObservations(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("query"))
	if query == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "missing query parameter",
		})
		return
	}
	payload := map[string]any{"query": query, "limit": 10}
	limit := 10
	if limitParam := strings.TrimSpace(r.URL.Query().Get("limit")); limitParam != "" {
		if parsed, err := strconv.Atoi(limitParam); err == nil {
			payload["limit"] = parsed
			if parsed > 0 {
				limit = parsed
			}
		}
	}
	namespace := strings.TrimSpace(r.URL.Query().Get("namespace"))
	if namespace != "" {
		payload["namespace"] = namespace
	}
	observationType := strings.TrimSpace(r.URL.Query().Get("type"))
	if observationType != "" {
		payload["type"] = observationType
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "memory.searchObservations", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "memory.searchObservations",
			},
		})
		return
	}

	records, localErr := s.localSearchObservationMemories(query, limit, namespace, observationType)
	if localErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error(), "detail": localErr.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    records,
		"bridge": map[string]any{
			"fallback":  "go-local-memory",
			"procedure": "memory.searchObservations",
			"reason":    "upstream unavailable; using local persisted observation search",
		},
	})
}

func (s *Server) handleMemoryCaptureUserPrompt(w http.ResponseWriter, r *http.Request) {
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "memory.captureUserPrompt", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": result, "bridge": map[string]any{"upstreamBase": upstreamBase, "procedure": "memory.captureUserPrompt"}})
		return
	}

	memory, localErr := s.localCaptureUserPromptMemory(payload)
	if localErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error(), "detail": localErr.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"success": true,
			"memory":  memory,
		},
		"bridge": map[string]any{
			"fallback":  "go-local-memory",
			"procedure": "memory.captureUserPrompt",
			"reason":    "upstream unavailable; persisted user prompt locally",
		},
	})
}

func (s *Server) handleMemoryRecentUserPrompts(w http.ResponseWriter, r *http.Request) {
	payload := map[string]any{"limit": 10}
	limit := 10
	if limitParam := strings.TrimSpace(r.URL.Query().Get("limit")); limitParam != "" {
		if parsed, err := strconv.Atoi(limitParam); err == nil {
			payload["limit"] = parsed
			if parsed > 0 {
				limit = parsed
			}
		}
	}
	role := strings.TrimSpace(r.URL.Query().Get("role"))
	if role != "" {
		payload["role"] = role
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "memory.getRecentUserPrompts", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "memory.getRecentUserPrompts",
			},
		})
		return
	}

	records, localErr := s.localRecentUserPrompts(limit, role)
	if localErr != nil {
		records = []map[string]any{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    records,
		"bridge": map[string]any{
			"fallback":  "go-local-memory",
			"procedure": "memory.getRecentUserPrompts",
			"reason":    "upstream unavailable; using local persisted recent user prompts",
		},
	})
}

func (s *Server) handleMemorySearchUserPrompts(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("query"))
	if query == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "missing query parameter",
		})
		return
	}
	payload := map[string]any{"query": query, "limit": 10}
	limit := 10
	if limitParam := strings.TrimSpace(r.URL.Query().Get("limit")); limitParam != "" {
		if parsed, err := strconv.Atoi(limitParam); err == nil {
			payload["limit"] = parsed
			if parsed > 0 {
				limit = parsed
			}
		}
	}
	role := strings.TrimSpace(r.URL.Query().Get("role"))
	if role != "" {
		payload["role"] = role
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "memory.searchUserPrompts", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "memory.searchUserPrompts",
			},
		})
		return
	}

	records, localErr := s.localSearchUserPromptMemories(query, limit, role)
	if localErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error(), "detail": localErr.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    records,
		"bridge": map[string]any{
			"fallback":  "go-local-memory",
			"procedure": "memory.searchUserPrompts",
			"reason":    "upstream unavailable; using local persisted user prompt search",
		},
	})
}

func (s *Server) handleMemorySearchPivot(w http.ResponseWriter, r *http.Request) {
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "memory.searchMemoryPivot", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "memory.searchMemoryPivot",
			},
		})
		return
	}

	data, reason, localErr := s.localSearchMemoryPivotPayload(payload)
	if localErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error(), "detail": localErr.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    data,
		"bridge": map[string]any{
			"fallback":  "go-local-memory",
			"procedure": "memory.searchMemoryPivot",
			"reason":    reason,
		},
	})
}

func (s *Server) handleMemoryTimelineWindow(w http.ResponseWriter, r *http.Request) {
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "memory.getMemoryTimelineWindow", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "memory.getMemoryTimelineWindow",
			},
		})
		return
	}

	data, reason, localErr := s.localTimelineWindowPayload(payload)
	if localErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error(), "detail": localErr.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    data,
		"bridge": map[string]any{
			"fallback":  "go-local-memory",
			"procedure": "memory.getMemoryTimelineWindow",
			"reason":    reason,
		},
	})
}

func (s *Server) handleMemoryCrossSessionLinks(w http.ResponseWriter, r *http.Request) {
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "memory.getCrossSessionMemoryLinks", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "memory.getCrossSessionMemoryLinks",
			},
		})
		return
	}

	data, reason, localErr := s.localCrossSessionMemoryLinksPayload(payload)
	if localErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error(), "detail": localErr.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    data,
		"bridge": map[string]any{
			"fallback":  "go-local-memory",
			"procedure": "memory.getCrossSessionMemoryLinks",
			"reason":    reason,
		},
	})
}

func (s *Server) handleMemorySessionBootstrap(w http.ResponseWriter, r *http.Request) {
	payload := map[string]any{}
	if activeGoal := strings.TrimSpace(r.URL.Query().Get("activeGoal")); activeGoal != "" {
		payload["activeGoal"] = activeGoal
	}
	if lastObjective := strings.TrimSpace(r.URL.Query().Get("lastObjective")); lastObjective != "" {
		payload["lastObjective"] = lastObjective
	}

	var result map[string]any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "memory.getSessionBootstrap", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "memory.getSessionBootstrap",
			},
		})
		return
	}

	activeGoal, _ := payload["activeGoal"].(string)
	lastObjective, _ := payload["lastObjective"].(string)
	data, localErr := s.localSessionBootstrapPayload(activeGoal, lastObjective)
	if localErr != nil {
		data = map[string]any{
			"activeGoal":             activeGoal,
			"lastObjective":          lastObjective,
			"goal":                   activeGoal,
			"objective":              lastObjective,
			"summaryCount":           0,
			"observationCount":       0,
			"toolAdvertisementCount": 0,
			"prompt":                 strings.Join(localBootstrapLines(activeGoal, lastObjective, nil, nil), "\n"),
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    data,
		"bridge": map[string]any{
			"fallback":  "go-local-memory",
			"procedure": "memory.getSessionBootstrap",
			"reason":    "upstream unavailable; using local persisted session bootstrap",
		},
	})
}

func (s *Server) handleMemoryToolContext(w http.ResponseWriter, r *http.Request) {
	toolName := strings.TrimSpace(r.URL.Query().Get("toolName"))
	if toolName == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "missing toolName query parameter",
		})
		return
	}
	payload := map[string]any{"toolName": toolName}
	if activeGoal := strings.TrimSpace(r.URL.Query().Get("activeGoal")); activeGoal != "" {
		payload["activeGoal"] = activeGoal
	}
	if lastObjective := strings.TrimSpace(r.URL.Query().Get("lastObjective")); lastObjective != "" {
		payload["lastObjective"] = lastObjective
	}

	var result map[string]any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "memory.getToolContext", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "memory.getToolContext",
			},
		})
		return
	}

	activeGoal, _ := payload["activeGoal"].(string)
	lastObjective, _ := payload["lastObjective"].(string)
	data, localErr := s.localToolContextPayload(toolName, activeGoal, lastObjective)
	if localErr != nil {
		data = map[string]any{
			"toolName":         toolName,
			"query":            toolName,
			"matchedPaths":     []string{},
			"observationCount": 0,
			"summaryCount":     0,
			"prompt":           strings.Join(localToolContextLines(toolName, toolName, activeGoal, lastObjective, nil, nil, nil), "\n"),
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    data,
		"bridge": map[string]any{
			"fallback":  "go-local-memory",
			"procedure": "memory.getToolContext",
			"reason":    "upstream unavailable; using local persisted tool context",
		},
	})
}

func (s *Server) handleMemoryCaptureSessionSummary(w http.ResponseWriter, r *http.Request) {
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "memory.captureSessionSummary", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": result, "bridge": map[string]any{"upstreamBase": upstreamBase, "procedure": "memory.captureSessionSummary"}})
		return
	}

	memory, localErr := s.localCaptureSessionSummaryMemory(payload)
	if localErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error(), "detail": localErr.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"success": true,
			"memory":  memory,
		},
		"bridge": map[string]any{
			"fallback":  "go-local-memory",
			"procedure": "memory.captureSessionSummary",
			"reason":    "upstream unavailable; persisted session summary locally",
		},
	})
}

func (s *Server) handleMemoryRecentSessionSummaries(w http.ResponseWriter, r *http.Request) {
	payload := map[string]any{"limit": 10}
	limit := 10
	if limitParam := strings.TrimSpace(r.URL.Query().Get("limit")); limitParam != "" {
		if parsed, err := strconv.Atoi(limitParam); err == nil {
			payload["limit"] = parsed
			if parsed > 0 {
				limit = parsed
			}
		}
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "memory.getRecentSessionSummaries", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "memory.getRecentSessionSummaries",
			},
		})
		return
	}

	records, localErr := s.localRecentSessionSummaries(limit)
	if localErr != nil {
		records = []map[string]any{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    records,
		"bridge": map[string]any{
			"fallback":  "go-local-memory",
			"procedure": "memory.getRecentSessionSummaries",
			"reason":    "upstream unavailable; using local persisted recent session summaries",
		},
	})
}

func (s *Server) handleMemorySearchSessionSummaries(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("query"))
	if query == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "missing query parameter",
		})
		return
	}
	payload := map[string]any{"query": query, "limit": 10}
	limit := 10
	if limitParam := strings.TrimSpace(r.URL.Query().Get("limit")); limitParam != "" {
		if parsed, err := strconv.Atoi(limitParam); err == nil {
			payload["limit"] = parsed
			if parsed > 0 {
				limit = parsed
			}
		}
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "memory.searchSessionSummaries", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "memory.searchSessionSummaries",
			},
		})
		return
	}

	records, localErr := s.localSearchSessionSummaryMemories(query, limit)
	if localErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error(), "detail": localErr.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    records,
		"bridge": map[string]any{
			"fallback":  "go-local-memory",
			"procedure": "memory.searchSessionSummaries",
			"reason":    "upstream unavailable; using local persisted session summary search",
		},
	})
}

func (s *Server) handleMemoryInterchangeFormats(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "memory.listInterchangeFormats", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "memory.listInterchangeFormats",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    []string{"json", "markdown"},
		"bridge": map[string]any{
			"fallback":  "go-local-memory",
			"procedure": "memory.listInterchangeFormats",
			"reason":    "upstream unavailable; using local memory interchange formats",
		},
	})
}

func (s *Server) handleMemoryExport(w http.ResponseWriter, r *http.Request) {
	payload := map[string]any{"userId": "default", "format": "json"}
	if userID := strings.TrimSpace(r.URL.Query().Get("userId")); userID != "" {
		payload["userId"] = userID
	}
	if format := strings.TrimSpace(r.URL.Query().Get("format")); format != "" {
		payload["format"] = format
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "memory.exportMemories", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "memory.exportMemories",
			},
		})
		return
	}

	userID, _ := payload["userId"].(string)
	format, _ := payload["format"].(string)
	localExport, fallbackErr := s.localMemoryExport(userID, format)
	if fallbackErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   fallbackErr.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"data":       localExport,
			"format":     format,
			"exportedAt": time.Now().UTC().Format(time.RFC3339),
		},
		"bridge": map[string]any{
			"fallback":  "go-local-memory",
			"procedure": "memory.exportMemories",
			"reason":    "upstream unavailable; using local memory export sources",
		},
	})
}

func (s *Server) handleMemoryImport(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "memory.importMemories")
}

func (s *Server) handleMemoryConvert(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "memory.convertMemories")
}

func (s *Server) handleAgentMemorySearch(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("query"))
	if query == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "missing query parameter",
		})
		return
	}
	payload := map[string]any{"query": query}
	namespace := strings.TrimSpace(r.URL.Query().Get("namespace"))
	if namespace != "" {
		payload["namespace"] = namespace
	}
	memoryType := strings.TrimSpace(r.URL.Query().Get("type"))
	if memoryType != "" {
		payload["type"] = memoryType
	}
	limit := 10
	if limitParam := strings.TrimSpace(r.URL.Query().Get("limit")); limitParam != "" {
		if parsed, err := strconv.Atoi(limitParam); err == nil {
			payload["limit"] = parsed
			if parsed > 0 {
				limit = parsed
			}
		}
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "agentMemory.search", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "agentMemory.search",
			},
		})
		return
	}

	records, localErr := s.localSearchAgentMemoryRecords(query, localAgentMemorySearchOptions{Limit: limit, Type: memoryType, Namespace: namespace})
	if localErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error(), "detail": localErr.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    localAgentMemoryMaps(records),
		"bridge": map[string]any{
			"fallback":  "go-local-agent-memory",
			"procedure": "agentMemory.search",
			"reason":    "upstream unavailable; using local persisted agent memory search",
		},
	})
}

func (s *Server) handleAgentMemoryAdd(w http.ResponseWriter, r *http.Request) {
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "agentMemory.add", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": result, "bridge": map[string]any{"upstreamBase": upstreamBase, "procedure": "agentMemory.add"}})
		return
	}

	metadata := localAgentMemoryMetadata(payload["metadata"])
	metadata["tags"] = localUniqueStrings(localAgentMemoryTags(metadata), stringArray(payload["tags"])...)
	record, localErr := s.localAddAgentMemoryEntry(
		stringValue(payload["content"]),
		localNormalizeAgentMemoryType(stringValue(payload["type"]), "working"),
		localNormalizeAgentMemoryNamespace(stringValue(payload["namespace"]), "project"),
		metadata,
	)
	if localErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error(), "detail": localErr.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    localAgentMemoryMap(record),
		"bridge": map[string]any{
			"fallback":  "go-local-agent-memory",
			"procedure": "agentMemory.add",
			"reason":    "upstream unavailable; persisted agent memory locally",
		},
	})
}

func (s *Server) handleAgentMemoryRecent(w http.ResponseWriter, r *http.Request) {
	payload := map[string]any{}
	memoryType := strings.TrimSpace(r.URL.Query().Get("type"))
	if memoryType != "" {
		payload["type"] = memoryType
	}
	limit := 10
	if limitParam := strings.TrimSpace(r.URL.Query().Get("limit")); limitParam != "" {
		if parsed, err := strconv.Atoi(limitParam); err == nil {
			payload["limit"] = parsed
			if parsed > 0 {
				limit = parsed
			}
		}
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "agentMemory.getRecent", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "agentMemory.getRecent",
			},
		})
		return
	}

	records, localErr := s.localRecentAgentMemoryRecords(limit, localAgentMemorySearchOptions{Type: memoryType})
	if localErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error(), "detail": localErr.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    localAgentMemoryMaps(records),
		"bridge": map[string]any{
			"fallback":  "go-local-agent-memory",
			"procedure": "agentMemory.getRecent",
			"reason":    "upstream unavailable; using local persisted recent agent memories",
		},
	})
}

func (s *Server) handleAgentMemoryByType(w http.ResponseWriter, r *http.Request) {
	memoryType := strings.TrimSpace(r.URL.Query().Get("type"))
	if memoryType == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "missing type query parameter",
		})
		return
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "agentMemory.getByType", map[string]any{"type": memoryType}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "agentMemory.getByType",
			},
		})
		return
	}

	records, localErr := s.localRecentAgentMemoryRecords(0, localAgentMemorySearchOptions{Type: memoryType})
	if localErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error(), "detail": localErr.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    localAgentMemoryMaps(records),
		"bridge": map[string]any{
			"fallback":  "go-local-agent-memory",
			"procedure": "agentMemory.getByType",
			"reason":    "upstream unavailable; using local persisted agent memories filtered by type",
		},
	})
}

func (s *Server) handleAgentMemoryByNamespace(w http.ResponseWriter, r *http.Request) {
	namespace := strings.TrimSpace(r.URL.Query().Get("namespace"))
	if namespace == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "missing namespace query parameter",
		})
		return
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "agentMemory.getByNamespace", map[string]any{"namespace": namespace}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "agentMemory.getByNamespace",
			},
		})
		return
	}

	records, localErr := s.localRecentAgentMemoryRecords(0, localAgentMemorySearchOptions{Namespace: namespace})
	if localErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error(), "detail": localErr.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    localAgentMemoryMaps(records),
		"bridge": map[string]any{
			"fallback":  "go-local-agent-memory",
			"procedure": "agentMemory.getByNamespace",
			"reason":    "upstream unavailable; using local persisted agent memories filtered by namespace",
		},
	})
}

func (s *Server) writeAgentMemoryListFallback(w http.ResponseWriter, procedure string) {
	writeJSON(w, http.StatusServiceUnavailable, map[string]any{
		"success": false,
		"error":   "Agent memory is unavailable: upstream agent memory service is unavailable and the local agent memory runtime is not initialized.",
		"data":    []any{},
		"bridge": map[string]any{
			"fallback":  "go-local-agent-memory",
			"procedure": procedure,
			"reason":    "upstream unavailable; local agent memory runtime is not initialized",
		},
	})
}

func (s *Server) handleAgentMemoryDelete(w http.ResponseWriter, r *http.Request) {
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "agentMemory.delete", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": result, "bridge": map[string]any{"upstreamBase": upstreamBase, "procedure": "agentMemory.delete"}})
		return
	}

	deleted, localErr := s.localDeleteAgentMemory(stringValue(payload["id"]))
	if localErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error(), "detail": localErr.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    deleted,
		"bridge": map[string]any{
			"fallback":  "go-local-agent-memory",
			"procedure": "agentMemory.delete",
			"reason":    "upstream unavailable; deleting persisted agent memory locally",
		},
	})
}

func (s *Server) handleAgentMemoryClearSession(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "agentMemory.clearSession", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": result, "bridge": map[string]any{"upstreamBase": upstreamBase, "procedure": "agentMemory.clearSession"}})
		return
	}

	cleared, localErr := s.localClearSessionAgentMemories()
	if localErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error(), "detail": localErr.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"success": true,
			"cleared": cleared,
		},
		"bridge": map[string]any{
			"fallback":  "go-local-agent-memory",
			"procedure": "agentMemory.clearSession",
			"reason":    "upstream unavailable; cleared persisted session-tier agent memories locally",
		},
	})
}

func (s *Server) handleAgentMemoryExport(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "agentMemory.export", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "agentMemory.export",
			},
		})
		return
	}

	exported, localErr := s.localAgentMemoryExport()
	if localErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error(), "detail": localErr.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    exported,
		"bridge": map[string]any{
			"fallback":  "go-local-agent-memory",
			"procedure": "agentMemory.export",
			"reason":    "upstream unavailable; exporting persisted agent memories locally",
		},
	})
}

func (s *Server) handleAgentMemoryHandoff(w http.ResponseWriter, r *http.Request) {
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "agentMemory.handoff", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": result, "bridge": map[string]any{"upstreamBase": upstreamBase, "procedure": "agentMemory.handoff"}})
		return
	}

	artifact, localErr := s.localAgentMemoryHandoffArtifact(payload)
	if localErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error(), "detail": localErr.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    artifact,
		"bridge": map[string]any{
			"fallback":  "go-local-agent-memory",
			"procedure": "agentMemory.handoff",
			"reason":    "upstream unavailable; generated local agent-memory handoff artifact",
		},
	})
}

func (s *Server) handleAgentMemoryPickup(w http.ResponseWriter, r *http.Request) {
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "agentMemory.pickup", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": result, "bridge": map[string]any{"upstreamBase": upstreamBase, "procedure": "agentMemory.pickup"}})
		return
	}

	pickupResult, localErr := s.localAgentMemoryPickupArtifact(stringValue(payload["artifact"]))
	if localErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error(), "detail": localErr.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    pickupResult,
		"bridge": map[string]any{
			"fallback":  "go-local-agent-memory",
			"procedure": "agentMemory.pickup",
			"reason":    "upstream unavailable; restored local agent-memory handoff artifact",
		},
	})
}

func (s *Server) handleAgentMemoryStats(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "agentMemory.stats", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "agentMemory.stats",
			},
		})
		return
	}

	stats, localErr := s.localAgentMemoryStatsCompact()
	if localErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": localErr.Error(), "detail": localErr.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    stats,
		"bridge": map[string]any{
			"fallback":  "go-local-agent-memory",
			"procedure": "agentMemory.stats",
			"reason":    "upstream unavailable; using local persisted agent memory stats",
		},
	})
}

func (s *Server) handleGraphGet(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodGet, "graph.get", nil)
}

func (s *Server) handleGraphRebuild(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodPost, "graph.rebuild", nil)
}

func (s *Server) handleGraphConsumers(w http.ResponseWriter, r *http.Request) {
	filePath := strings.TrimSpace(r.URL.Query().Get("filePath"))
	if filePath == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing filePath query parameter"})
		return
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "graph.getConsumers", map[string]any{"filePath": filePath}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "graph.getConsumers",
			},
		})
		return
	}

	s.writeGraphListFallback(w, "graph.getConsumers")
}

func (s *Server) handleGraphDependencies(w http.ResponseWriter, r *http.Request) {
	filePath := strings.TrimSpace(r.URL.Query().Get("filePath"))
	if filePath == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing filePath query parameter"})
		return
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "graph.getDependencies", map[string]any{"filePath": filePath}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "graph.getDependencies",
			},
		})
		return
	}

	s.writeGraphListFallback(w, "graph.getDependencies")
}

func (s *Server) writeGraphListFallback(w http.ResponseWriter, procedure string) {
	writeJSON(w, http.StatusServiceUnavailable, map[string]any{
		"success": false,
		"error":   "Repository graph is unavailable: upstream graph service is unavailable and the local graph index is not initialized.",
		"data":    []string{},
		"bridge": map[string]any{
			"fallback":  "go-local-graph",
			"procedure": procedure,
			"reason":    "upstream unavailable; repo graph is not initialized",
		},
	})
}

func (s *Server) handleGraphSymbols(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "graph.getSymbolsGraph", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "graph.getSymbolsGraph",
			},
		})
		return
	}

	writeJSON(w, http.StatusServiceUnavailable, map[string]any{
		"success": false,
		"error":   "Symbol graph is unavailable: upstream graph service is unavailable and local symbol graph data is not initialized.",
		"data": map[string]any{
			"nodes": []map[string]any{},
			"links": []map[string]any{},
		},
		"bridge": map[string]any{
			"fallback":  "go-local-graph",
			"procedure": "graph.getSymbolsGraph",
			"reason":    "upstream unavailable; symbol graph data is not initialized",
		},
	})
}

func (s *Server) handleContextList(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "borgContext.list", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "borgContext.list",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    []string{},
		"bridge": map[string]any{
			"fallback":  "go-local-context",
			"procedure": "borgContext.list",
			"reason":    "upstream unavailable; using local empty context list",
		},
	})
}

func (s *Server) handleContextAdd(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "borgContext.add")
}

func (s *Server) handleContextRemove(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "borgContext.remove")
}

func (s *Server) handleContextClear(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodPost, "borgContext.clear", nil)
}

func (s *Server) handleContextPrompt(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "borgContext.getPrompt", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "borgContext.getPrompt",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    "",
		"bridge": map[string]any{
			"fallback":  "go-local-context",
			"procedure": "borgContext.getPrompt",
			"reason":    "upstream unavailable; using local empty context prompt",
		},
	})
}

func (s *Server) handleGitModules(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "git.getModules", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "git.getModules",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    localGitModules(s.cfg.WorkspaceRoot),
		"bridge": map[string]any{
			"fallback":  "go-local-git",
			"procedure": "git.getModules",
			"reason":    "upstream unavailable; using local .gitmodules parsing",
		},
	})
}

func (s *Server) handleGitLog(w http.ResponseWriter, r *http.Request) {
	payload := map[string]any{}
	limit := 20
	if rawLimit := strings.TrimSpace(r.URL.Query().Get("limit")); rawLimit != "" {
		if parsed, err := strconv.Atoi(rawLimit); err == nil {
			payload["limit"] = parsed
			if parsed > 0 {
				limit = parsed
			}
		}
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "git.getLog", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "git.getLog",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    localGitLog(s.cfg.WorkspaceRoot, limit),
		"bridge": map[string]any{
			"fallback":  "go-local-git",
			"procedure": "git.getLog",
			"reason":    "upstream unavailable; using local git log fallback",
		},
	})
}

func (s *Server) handleGitStatus(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "git.getStatus", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "git.getStatus",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    localGitStatus(s.cfg.WorkspaceRoot),
		"bridge": map[string]any{
			"fallback":  "go-local-git",
			"procedure": "git.getStatus",
			"reason":    "upstream unavailable; using local git status fallback",
		},
	})
}

func (s *Server) handleGitRevert(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "git.revert")
}

func (s *Server) handleTestsStatus(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "tests.status", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "tests.status",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"isRunning": false,
			"results":   map[string]any{},
		},
		"bridge": map[string]any{
			"fallback":  "go-local-tests",
			"procedure": "tests.status",
			"reason":    "upstream unavailable; using local zero-state test status",
		},
	})
}

func (s *Server) handleTestsStart(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodPost, "tests.start", nil)
}

func (s *Server) handleTestsStop(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeCall(w, r, http.MethodPost, "tests.stop", nil)
}

func (s *Server) handleTestsRun(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "tests.run")
}

func (s *Server) handleTestsResults(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "tests.results", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "tests.results",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    []map[string]any{},
		"bridge": map[string]any{
			"fallback":  "go-local-tests",
			"procedure": "tests.results",
			"reason":    "upstream unavailable; using local empty test results",
		},
	})
}

func (s *Server) handleMetricsStats(w http.ResponseWriter, r *http.Request) {
	payload := map[string]any{}
	windowMs := 3600000
	if windowParam := strings.TrimSpace(r.URL.Query().Get("windowMs")); windowParam != "" {
		if parsed, err := strconv.Atoi(windowParam); err == nil {
			payload["windowMs"] = parsed
			windowMs = parsed
		}
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "metrics.getStats", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "metrics.getStats",
			},
		})
		return
	}

	writeJSON(w, http.StatusServiceUnavailable, map[string]any{
		"success": false,
		"error":   "Metrics stats are unavailable: upstream metrics service is unavailable and the local event store is not implemented.",
		"data": map[string]any{
			"windowMs":    windowMs,
			"totalEvents": 0,
			"counts":      map[string]any{},
			"averages":    map[string]any{},
			"series":      []any{},
		},
		"bridge": map[string]any{
			"fallback":  "go-local-metrics-preview",
			"procedure": "metrics.getStats",
			"reason":    "upstream unavailable; local metrics event store is not implemented",
		},
	})
}

func (s *Server) handleMetricsTrack(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "metrics.track")
}

func (s *Server) handleMetricsSystemSnapshot(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "metrics.systemSnapshot", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "metrics.systemSnapshot",
			},
		})
		return
	}

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	hostname, hostnameErr := os.Hostname()
	if hostnameErr != nil {
		hostname = "unknown"
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"timestamp": time.Now().UnixMilli(),
			"process": map[string]any{
				"heapUsed":      mem.HeapAlloc,
				"heapTotal":     mem.HeapSys,
				"rss":           mem.Sys,
				"external":      0,
				"arrayBuffers":  0,
				"uptimeSeconds": int(time.Since(s.startedAt).Seconds()),
				"pid":           os.Getpid(),
			},
			"system": map[string]any{
				"totalMemory":        0,
				"freeMemory":         0,
				"usedMemory":         0,
				"memoryUsagePercent": 0,
				"cpuCount":           runtime.NumCPU(),
				"cpuModel":           "unknown",
				"loadAvg":            []float64{},
				"platform":           runtime.GOOS,
				"arch":               runtime.GOARCH,
				"hostname":           hostname,
			},
		},
		"bridge": map[string]any{
			"fallback":  "go-local-system-snapshot",
			"procedure": "metrics.systemSnapshot",
			"reason":    "upstream unavailable; using native Go runtime snapshot",
		},
	})
}

func (s *Server) handleMetricsTimeline(w http.ResponseWriter, r *http.Request) {
	payload := map[string]any{}
	windowMs := 3600000
	if windowParam := strings.TrimSpace(r.URL.Query().Get("windowMs")); windowParam != "" {
		if parsed, err := strconv.Atoi(windowParam); err == nil {
			payload["windowMs"] = parsed
			windowMs = parsed
		}
	}
	buckets := 60
	if bucketParam := strings.TrimSpace(r.URL.Query().Get("buckets")); bucketParam != "" {
		if parsed, err := strconv.Atoi(bucketParam); err == nil {
			payload["buckets"] = parsed
			buckets = parsed
		}
	}
	metricType := "all"
	if metricTypeParam := strings.TrimSpace(r.URL.Query().Get("metricType")); metricTypeParam != "" {
		payload["metricType"] = metricTypeParam
		metricType = metricTypeParam
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "metrics.getTimeline", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "metrics.getTimeline",
			},
		})
		return
	}

	writeJSON(w, http.StatusServiceUnavailable, map[string]any{
		"success": false,
		"error":   "Metrics timeline is unavailable: upstream metrics service is unavailable and the local event store is not implemented.",
		"data": map[string]any{
			"windowMs":   windowMs,
			"buckets":    buckets,
			"metricType": metricType,
			"series":     []any{},
			"counts":     map[string]any{},
			"averages":   map[string]any{},
		},
		"bridge": map[string]any{
			"fallback":  "go-local-metrics-preview",
			"procedure": "metrics.getTimeline",
			"reason":    "upstream unavailable; local metrics event store is not implemented",
		},
	})
}

func (s *Server) handleMetricsProviderBreakdown(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "metrics.getProviderBreakdown", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "metrics.getProviderBreakdown",
			},
		})
		return
	}

	statuses := providers.Snapshot()
	catalog := providers.Catalog(statuses)
	providersPreview := make([]map[string]any, 0, len(catalog))
	for _, provider := range catalog {
		providersPreview = append(providersPreview, map[string]any{
			"provider": provider.Name,
			"cost":     0,
			"requests": 0,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"totalCost":      0,
			"totalRequests":  0,
			"averageLatency": 0,
			"providers":      providersPreview,
		},
		"bridge": map[string]any{
			"fallback":  "go-local-metrics-preview",
			"procedure": "metrics.getProviderBreakdown",
			"reason":    "upstream unavailable; using local provider catalog with zeroed usage",
		},
	})
}

func (s *Server) handleMetricsMonitoring(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "metrics.toggleMonitoring")
}

func (s *Server) handleMetricsRoutingHistory(w http.ResponseWriter, r *http.Request) {
	payload := map[string]any{}
	if limit := strings.TrimSpace(r.URL.Query().Get("limit")); limit != "" {
		if parsed, err := strconv.Atoi(limit); err == nil {
			payload["limit"] = parsed
		}
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "metrics.getRoutingHistory", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "metrics.getRoutingHistory",
			},
		})
		return
	}

	writeJSON(w, http.StatusServiceUnavailable, map[string]any{
		"success": false,
		"error":   "Routing history is unavailable: upstream metrics service is unavailable and the local routing history buffer is not implemented.",
		"data":    []any{},
		"bridge": map[string]any{
			"fallback":  "go-local-metrics-preview",
			"procedure": "metrics.getRoutingHistory",
			"reason":    "upstream unavailable; local routing history buffer is not implemented",
		},
	})
}

func (s *Server) handleLogsList(w http.ResponseWriter, r *http.Request) {
	payload := map[string]any{}
	filter := localLogsFilter{limit: 100}
	if limit := strings.TrimSpace(r.URL.Query().Get("limit")); limit != "" {
		if parsed, err := strconv.Atoi(limit); err == nil {
			payload["limit"] = parsed
			if parsed > 0 {
				filter.limit = parsed
			}
		}
	}
	if sessionID := strings.TrimSpace(r.URL.Query().Get("sessionId")); sessionID != "" {
		payload["sessionId"] = sessionID
		filter.sessionID = sessionID
	}
	if serverName := strings.TrimSpace(r.URL.Query().Get("serverName")); serverName != "" {
		payload["serverName"] = serverName
		filter.serverName = serverName
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "logs.list", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "logs.list",
			},
		})
		return
	}

	logs, fallbackErr := s.localObservabilityLogs(filter)
	if fallbackErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": fallbackErr.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    logs,
		"bridge": map[string]any{
			"fallback":  "go-local-observability",
			"procedure": "logs.list",
			"reason":    "upstream unavailable; using local tool_call_logs records",
		},
	})
}

func (s *Server) handleLogsSummary(w http.ResponseWriter, r *http.Request) {
	payload := map[string]any{}
	filter := localLogsFilter{limit: 1000}
	if limit := strings.TrimSpace(r.URL.Query().Get("limit")); limit != "" {
		if parsed, err := strconv.Atoi(limit); err == nil {
			payload["limit"] = parsed
			if parsed > 0 {
				filter.limit = parsed
			}
		}
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "logs.summary", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "logs.summary",
			},
		})
		return
	}

	summary, fallbackErr := s.localObservabilitySummary(filter)
	if fallbackErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": fallbackErr.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    summary,
		"bridge": map[string]any{
			"fallback":  "go-local-observability",
			"procedure": "logs.summary",
			"reason":    "upstream unavailable; summarizing local tool_call_logs records",
		},
	})
}

func (s *Server) handleLogsClear(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "logs.clear", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "logs.clear",
			},
		})
		return
	}
	s.handleLogsClearFallback(w)
}

type localLogsFilter struct {
	limit      int
	sessionID  string
	serverName string
}

type localObservabilityLog struct {
	ID             string
	Timestamp      int64
	ServerName     string
	Level          string
	Message        string
	ToolName       string
	Error          any
	Arguments      any
	Result         any
	DurationMs     string
	SessionID      string
	ParentCallUUID any
}

func (s *Server) localObservabilityLogs(filter localLogsFilter) ([]map[string]any, error) {
	db, err := sql.Open("sqlite", s.localMetaMCPDBPath())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	if filter.limit <= 0 {
		filter.limit = 100
	}

	query := `
		SELECT l.uuid, l.created_at, COALESCE(m.name, ''), l.tool_name, l.error, l.args, l.result, l.duration_ms, COALESCE(l.session_id, ''), l.parent_call_uuid
		FROM tool_call_logs l
		LEFT JOIN mcp_servers m ON l.mcp_server_uuid = m.uuid
		WHERE (? = '' OR l.session_id = ?)
		  AND (? = '' OR m.name = ?)
		ORDER BY l.created_at DESC
		LIMIT ?
	`
	rows, err := db.Query(query, filter.sessionID, filter.sessionID, filter.serverName, filter.serverName, filter.limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := []map[string]any{}
	for rows.Next() {
		log, scanErr := scanObservabilityLog(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		logs = append(logs, log)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return logs, nil
}

func (s *Server) localObservabilitySummary(filter localLogsFilter) (map[string]any, error) {
	logs, err := s.localObservabilityLogs(filter)
	if err != nil {
		return nil, err
	}

	totalCalls := len(logs)
	errorCount := 0
	durationSum := 0
	toolMap := map[string]map[string]any{}
	recentActivity := make([]map[string]any, 0, min(30, len(logs)))

	for i, entry := range logs {
		level := strings.ToLower(jsonString(entry["level"]))
		isError := level == "error" || entry["error"] != nil
		if isError {
			errorCount++
		}
		duration := atoiDefault(jsonString(entry["durationMs"]), 0)
		durationSum += duration

		toolName := jsonString(entry["toolName"])
		if toolName == "" {
			toolName = "unknown"
		}
		current := toolMap[toolName]
		if current == nil {
			current = map[string]any{"name": toolName, "count": 0, "errors": 0}
		}
		current["count"] = anyInt64(current["count"]) + 1
		if isError {
			current["errors"] = anyInt64(current["errors"]) + 1
		}
		toolMap[toolName] = current

		if i < 30 {
			recentActivity = append(recentActivity, map[string]any{
				"toolName":   toolName,
				"durationMs": duration,
				"error":      isError,
				"timestamp":  entry["timestamp"],
			})
		}
	}

	avgDurationMs := 0
	errorRate := 0.0
	successRate := 100.0
	if totalCalls > 0 {
		avgDurationMs = int(math.Round(float64(durationSum) / float64(totalCalls)))
		errorRate = math.Round((float64(errorCount)/float64(totalCalls))*1000) / 10
		successRate = math.Round((100-errorRate)*10) / 10
	}

	topTools := make([]map[string]any, 0, len(toolMap))
	for _, entry := range toolMap {
		topTools = append(topTools, entry)
	}
	sort.Slice(topTools, func(i, j int) bool {
		return anyInt64(topTools[i]["count"]) > anyInt64(topTools[j]["count"])
	})
	if len(topTools) > 10 {
		topTools = topTools[:10]
	}

	return map[string]any{
		"totals": map[string]any{
			"totalCalls":    totalCalls,
			"errorCount":    errorCount,
			"errorRate":     errorRate,
			"avgDurationMs": avgDurationMs,
			"successRate":   successRate,
		},
		"topTools":       topTools,
		"recentActivity": recentActivity,
	}, nil
}

func (s *Server) clearLocalObservabilityLogs() error {
	db, err := sql.Open("sqlite", s.localMetaMCPDBPath())
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(`DELETE FROM tool_call_logs`)
	return err
}

func scanObservabilityLog(scanner interface{ Scan(dest ...any) error }) (map[string]any, error) {
	var (
		id             string
		createdAt      int64
		serverName     string
		toolName       string
		errorValue     sql.NullString
		argsRaw        sql.NullString
		resultRaw      sql.NullString
		durationMs     sql.NullInt64
		sessionID      string
		parentCallUUID sql.NullString
	)
	if err := scanner.Scan(&id, &createdAt, &serverName, &toolName, &errorValue, &argsRaw, &resultRaw, &durationMs, &sessionID, &parentCallUUID); err != nil {
		return nil, err
	}

	level := "info"
	if errorValue.Valid && strings.TrimSpace(errorValue.String) != "" {
		level = "error"
	}

	return map[string]any{
		"id":             id,
		"timestamp":      unixTimestampToRFC3339(createdAt),
		"serverName":     emptyStringToNilAny(serverName),
		"level":          level,
		"message":        "Tool call: " + toolName,
		"toolName":       toolName,
		"error":          nullStringToAny(errorValue),
		"arguments":      jsonNullStringObjectOrNil(argsRaw),
		"result":         jsonNullStringObjectOrNil(resultRaw),
		"durationMs":     nullInt64ToStringAny(durationMs),
		"sessionId":      emptyStringToNilAny(sessionID),
		"parentCallUuid": nullStringToAny(parentCallUUID),
	}, nil
}

func jsonNullStringObjectOrNil(raw sql.NullString) any {
	if !raw.Valid || strings.TrimSpace(raw.String) == "" {
		return nil
	}
	var parsed any
	if err := json.Unmarshal([]byte(raw.String), &parsed); err != nil {
		return nil
	}
	return parsed
}

func nullInt64ToStringAny(value sql.NullInt64) any {
	if !value.Valid {
		return nil
	}
	return strconv.FormatInt(value.Int64, 10)
}

func atoiDefault(raw string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return fallback
	}
	return value
}

func jsonString(value any) string {
	if s, ok := value.(string); ok {
		return s
	}
	return ""
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s *Server) handleLogsClearFallback(w http.ResponseWriter) {
	if err := s.clearLocalObservabilityLogs(); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"success": true,
			"message": "Logs cleared",
		},
		"bridge": map[string]any{
			"fallback":  "go-local-observability",
			"procedure": "logs.clear",
			"reason":    "upstream unavailable; cleared local tool_call_logs records",
		},
	})
}

func (s *Server) handleServerHealthCheck(w http.ResponseWriter, r *http.Request) {
	serverUUID := strings.TrimSpace(r.URL.Query().Get("serverUuid"))
	if serverUUID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing serverUuid query parameter"})
		return
	}
	payload := map[string]any{"serverUuid": serverUUID}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "serverHealth.check", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "serverHealth.check",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    s.localServerHealth(serverUUID),
		"bridge": map[string]any{
			"fallback":  "go-local-server-health",
			"procedure": "serverHealth.check",
			"reason":    "upstream unavailable; using cached local mcp.jsonc server metadata",
		},
	})
}

func (s *Server) handleServerHealthReset(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "serverHealth.reset")
}

func (s *Server) handleSettingsGet(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "settings.get", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "settings.get",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    localSettingsConfig(s.cfg.WorkspaceRoot),
		"bridge": map[string]any{
			"fallback":  "go-local-settings",
			"procedure": "settings.get",
			"reason":    "upstream unavailable; using local .borg config fallback",
		},
	})
}

func (s *Server) handleSettingsUpdate(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "settings.update")
}

func (s *Server) handleSettingsProviders(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "settings.getProviders", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "settings.getProviders",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    providers.Catalog(providers.Snapshot()),
		"bridge": map[string]any{
			"fallback":  "go-local-provider-routing",
			"procedure": "settings.getProviders",
			"reason":    "upstream unavailable; using local provider catalog visibility",
		},
	})
}

func (s *Server) handleSettingsTestConnection(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "settings.testConnection")
}

func (s *Server) handleSettingsEnvironment(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "settings.getEnvironment", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "settings.getEnvironment",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    localSettingsEnvironment(),
		"bridge": map[string]any{
			"fallback":  "go-local-settings",
			"procedure": "settings.getEnvironment",
			"reason":    "upstream unavailable; using local Go runtime environment diagnostics",
		},
	})
}

func (s *Server) handleSettingsMCPServers(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "settings.getMcpServers", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "settings.getMcpServers",
			},
		})
		return
	}

	servers, fallbackErr := s.localConfiguredMCPServers()
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
		"data":    servers,
		"bridge": map[string]any{
			"fallback":  "go-local-settings",
			"procedure": "settings.getMcpServers",
			"reason":    "upstream unavailable; using local MCP config servers",
		},
	})
}

func (s *Server) handleSettingsProviderKey(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "settings.updateProviderKey")
}

func (s *Server) handleToolsList(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "tools.list", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "tools.list",
			},
		})
		return
	}

	tools, fallbackErr := s.localDBTools()
	if fallbackErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   fallbackErr.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    tools,
		"bridge": map[string]any{
			"fallback":  "go-local-tool-db",
			"procedure": "tools.list",
			"reason":    "upstream unavailable; using local tools from metamcp.db",
		},
	})
}

func (s *Server) handleToolsByServer(w http.ResponseWriter, r *http.Request) {
	serverID := strings.TrimSpace(r.URL.Query().Get("mcpServerUuid"))
	if serverID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing mcpServerUuid query parameter"})
		return
	}
	payload := map[string]any{"mcpServerUuid": serverID}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "tools.listByServer", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "tools.listByServer",
			},
		})
		return
	}

	filtered, fallbackErr := s.localDBToolsByServer(serverID)
	if fallbackErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   fallbackErr.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    filtered,
		"bridge": map[string]any{
			"fallback":  "go-local-tool-db",
			"procedure": "tools.listByServer",
			"reason":    "upstream unavailable; filtering local tools from metamcp.db by server",
		},
	})
}

func (s *Server) handleToolsSearch(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("query"))
	if query == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing query parameter"})
		return
	}
	payload := map[string]any{"query": query}
	if limit := strings.TrimSpace(r.URL.Query().Get("limit")); limit != "" {
		if parsed, err := strconv.Atoi(limit); err == nil {
			payload["limit"] = parsed
		}
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "tools.search", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "tools.search",
			},
		})
		return
	}

	limit := 30
	if rawLimit, ok := payload["limit"].(int); ok && rawLimit > 0 {
		limit = rawLimit
	}
	results, fallbackErr := s.localDBToolSearch(query, limit)
	if fallbackErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   fallbackErr.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    results,
		"bridge": map[string]any{
			"fallback":  "go-local-tool-db",
			"procedure": "tools.search",
			"reason":    "upstream unavailable; searching local tools from metamcp.db",
		},
	})
}

func (s *Server) handleToolsDetectCLIHarnesses(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "tools.detectCliHarnesses", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "tools.detectCliHarnesses",
			},
		})
		return
	}

	data, fallbackErr := s.localDetectedCliHarnesses(r.Context())
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
		"data":    data,
		"bridge": map[string]any{
			"fallback":  "go-local-tools",
			"procedure": "tools.detectCliHarnesses",
			"reason":    "upstream unavailable; using local Go harness detection",
		},
	})
}

func (s *Server) handleToolsDetectExecutionEnvironment(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "tools.detectExecutionEnvironment", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "tools.detectExecutionEnvironment",
			},
		})
		return
	}

	data, fallbackErr := s.localExecutionEnvironment(r.Context())
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
		"data":    data,
		"bridge": map[string]any{
			"fallback":  "go-local-tools",
			"procedure": "tools.detectExecutionEnvironment",
			"reason":    "upstream unavailable; using local Go execution environment detection",
		},
	})
}

func (s *Server) handleToolsDetectInstallSurfaces(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "tools.detectInstallSurfaces", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "tools.detectInstallSurfaces",
			},
		})
		return
	}

	data := s.localInstallSurfaces()
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    data,
		"bridge": map[string]any{
			"fallback":  "go-local-tools",
			"procedure": "tools.detectInstallSurfaces",
			"reason":    "upstream unavailable; using local Go install-surface detection",
		},
	})
}

func (s *Server) handleToolsGet(w http.ResponseWriter, r *http.Request) {
	uuid := strings.TrimSpace(r.URL.Query().Get("uuid"))
	if uuid == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing uuid query parameter"})
		return
	}
	payload := map[string]any{"uuid": uuid}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "tools.get", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "tools.get",
			},
		})
		return
	}

	fallbackTool, fallbackErr := s.localDBTool(uuid)
	if fallbackErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   fallbackErr.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    fallbackTool,
		"bridge": map[string]any{
			"fallback":  "go-local-tool-db",
			"procedure": "tools.get",
			"reason":    "upstream unavailable; using local tool from metamcp.db",
		},
	})
}

func (s *Server) handleToolsCreate(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "tools.create")
}

func (s *Server) handleToolsUpsertBatch(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "tools.upsertBatch")
}

func (s *Server) handleToolsDelete(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "tools.delete")
}

func (s *Server) handleToolsAlwaysOn(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "tools.setAlwaysOn")
}

func (s *Server) handleToolSetsList(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "toolSets.list", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "toolSets.list",
			},
		})
		return
	}

	toolSets, fallbackErr := s.localToolSets()
	if fallbackErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   fallbackErr.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    toolSets,
		"bridge": map[string]any{
			"fallback":  "go-local-operator",
			"procedure": "toolSets.list",
			"reason":    "upstream unavailable; using local tool sets from metamcp.db",
		},
	})
}

func (s *Server) handleToolSetsGet(w http.ResponseWriter, r *http.Request) {
	uuid := strings.TrimSpace(r.URL.Query().Get("uuid"))
	if uuid == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing uuid query parameter"})
		return
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "toolSets.get", map[string]any{"uuid": uuid}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "toolSets.get",
			},
		})
		return
	}

	toolSets, fallbackErr := s.localToolSets()
	if fallbackErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   fallbackErr.Error(),
		})
		return
	}

	for _, toolSet := range toolSets {
		if stringValue(toolSet["uuid"]) == uuid {
			writeJSON(w, http.StatusOK, map[string]any{
				"success": true,
				"data":    toolSet,
				"bridge": map[string]any{
					"fallback":  "go-local-operator",
					"procedure": "toolSets.get",
					"reason":    "upstream unavailable; using local tool set from metamcp.db",
				},
			})
			return
		}
	}

	writeJSON(w, http.StatusServiceUnavailable, map[string]any{
		"success": false,
		"error":   "tool set unavailable",
		"detail":  "upstream unavailable; tool set was not found in local metamcp.db",
		"bridge": map[string]any{
			"fallback":  "go-local-operator",
			"procedure": "toolSets.get",
			"reason":    "upstream unavailable; tool set was not found in local metamcp.db",
		},
	})
}

func (s *Server) handleToolSetsCreate(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "toolSets.create")
}

func (s *Server) handleToolSetsUpdate(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "toolSets.update")
}

func (s *Server) handleToolSetsDelete(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "toolSets.delete")
}

func (s *Server) handleProjectContext(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "project.getContext", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "project.getContext",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    localProjectContext(s.cfg.WorkspaceRoot),
		"bridge": map[string]any{
			"fallback":  "go-local-project",
			"procedure": "project.getContext",
			"reason":    "upstream unavailable; using local project context document",
		},
	})
}

func (s *Server) handleProjectContextUpdate(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "project.updateContext")
}

func (s *Server) handleProjectHandoffs(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "project.getHandoffs", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "project.getHandoffs",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    localProjectHandoffs(s.cfg.WorkspaceRoot),
		"bridge": map[string]any{
			"fallback":  "go-local-project",
			"procedure": "project.getHandoffs",
			"reason":    "upstream unavailable; using local handoff directory listing",
		},
	})
}

func (s *Server) handleShellLog(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "shell.logCommand")
}

func (s *Server) handleShellQueryHistory(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("query"))
	if query == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing query parameter"})
		return
	}
	payload := map[string]any{"query": query}
	if limit := strings.TrimSpace(r.URL.Query().Get("limit")); limit != "" {
		if parsed, err := strconv.Atoi(limit); err == nil {
			payload["limit"] = parsed
		}
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "shell.queryHistory", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "shell.queryHistory",
			},
		})
		return
	}

	limit := 20
	if value, ok := payload["limit"].(int); ok && value > 0 {
		limit = value
	}
	results, fallbackErr := s.localShellQueryHistory(query, limit)
	if fallbackErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   fallbackErr.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    results,
		"bridge": map[string]any{
			"fallback":  "go-local-shell",
			"procedure": "shell.queryHistory",
			"reason":    "upstream unavailable; using local enriched shell history",
		},
	})
}

func (s *Server) handleShellSystemHistory(w http.ResponseWriter, r *http.Request) {
	payload := map[string]any{}
	if limit := strings.TrimSpace(r.URL.Query().Get("limit")); limit != "" {
		if parsed, err := strconv.Atoi(limit); err == nil {
			payload["limit"] = parsed
		}
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "shell.getSystemHistory", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "shell.getSystemHistory",
			},
		})
		return
	}

	limit := 50
	if value, ok := payload["limit"].(int); ok && value > 0 {
		limit = value
	}
	results, fallbackErr := s.localShellSystemHistory(limit)
	if fallbackErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   fallbackErr.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    results,
		"bridge": map[string]any{
			"fallback":  "go-local-shell",
			"procedure": "shell.getSystemHistory",
			"reason":    "upstream unavailable; using local shell history file",
		},
	})
}

func (s *Server) handleAgentRunTool(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	var payload struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	// 1. Try native Go tool handlers first (Total Autonomy)
	if s.toolsRegistry != nil && s.toolsRegistry.HasTool(payload.Name) {
		result, err := s.toolsRegistry.Execute(r.Context(), payload.Name, payload.Arguments)
		if err == nil {
			writeJSON(w, http.StatusOK, map[string]any{
				"success": true,
				"data":    result,
				"bridge": map[string]any{
					"source": "go-native-tool",
					"tool":   payload.Name,
				},
			})
			return
		}
		// If native tool fails, return the error
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   err.Error(),
			"source":  "go-native-tool",
		})
		return
	}

	// 2. Try upstream Node server (Bridge)
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "agent.runTool", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "agent.runTool",
			},
		})
		return
	}

	// 3. Fallback: Descriptive error
	writeJSON(w, http.StatusServiceUnavailable, map[string]any{
		"success": false,
		"error":   "Tool not found or upstream unavailable",
		"detail":  fmt.Sprintf("Tool '%s' not implemented in Go and Node server is unreachable.", payload.Name),
	})
}

func (s *Server) handleCommandsExecute(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "commands.execute")
}

func (s *Server) handleCommandsList(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "commands.list", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "commands.list",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    []map[string]any{},
		"bridge": map[string]any{
			"fallback":  "go-local-registry",
			"procedure": "commands.list",
			"reason":    "upstream unavailable; using local empty command registry",
		},
	})
}

func (s *Server) handleSkillsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	var skills []map[string]any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "skills.list", nil, &skills)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    skills,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "skills.list",
			},
		})
		return
	}

	fallbackSkills, fallbackErr := s.localSkillsMetadata()
	if fallbackErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": fallbackErr.Error(), "detail": fallbackErr.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    fallbackSkills,
		"bridge": map[string]any{
			"fallback":  "go-local-skills",
			"procedure": "skills.list",
			"reason":    "upstream unavailable; using local skills metadata",
		},
	})
}

func (s *Server) handleSkillsSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	var rawSkills []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Path string `json:"path"`
	}
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "skills.list", nil, &rawSkills)
	if err == nil {
		query := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("query")))
		summaries := make([]SkillSummary, 0, len(rawSkills))
		for _, skill := range rawSkills {
			folder := strings.TrimSpace(filepath.Base(filepath.Dir(skill.Path)))
			if folder == "." || folder == string(filepath.Separator) {
				folder = ""
			}
			summary := SkillSummary{
				ID:     skill.ID,
				Name:   skill.Name,
				Folder: folder,
			}
			if query != "" {
				haystack := strings.ToLower(strings.Join([]string{summary.ID, summary.Name, summary.Folder}, " "))
				if !strings.Contains(haystack, query) {
					continue
				}
			}
			summaries = append(summaries, summary)
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    summaries,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "skills.list",
			},
		})
		return
	}

	fallbackSummaries, fallbackErr := s.localSkillSummaries(strings.TrimSpace(r.URL.Query().Get("query")))
	if fallbackErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": fallbackErr.Error(), "detail": fallbackErr.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    fallbackSummaries,
		"bridge": map[string]any{
			"fallback":  "go-local-skills",
			"procedure": "skills.list",
			"reason":    "upstream unavailable; using local skill folder summaries",
		},
	})
}

func (s *Server) handleSkillsRead(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	if name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing name query parameter"})
		return
	}

	var result map[string]any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "skills.read", map[string]any{"name": name}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "skills.read",
			},
		})
		return
	}

	fallbackResult, fallbackErr := s.localReadSkill(name)
	if fallbackErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": fallbackErr.Error(), "detail": fallbackErr.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    fallbackResult,
		"bridge": map[string]any{
			"fallback":  "go-local-skills",
			"procedure": "skills.read",
			"reason":    "upstream unavailable; using local skill document",
		},
	})
}

func (s *Server) handleSkillsCreate(w http.ResponseWriter, r *http.Request) {
	s.handleSkillMutation(w, r, "skills.create", func(payload map[string]any) (any, error) {
		return s.localCreateSkill(payload)
	})
}

func (s *Server) handleSkillsSave(w http.ResponseWriter, r *http.Request) {
	s.handleSkillMutation(w, r, "skills.save", func(payload map[string]any) (any, error) {
		return s.localSaveSkill(payload)
	})
}

func (s *Server) handleSkillsSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "query parameter required"})
		return
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "skills.search", map[string]any{"query": query}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "skills.search",
			},
		})
		return
	}

	res, fallbackErr := s.skillStore.SearchSkills(query)
	if fallbackErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": fallbackErr.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    res,
		"bridge": map[string]any{
			"fallback": "go-local-skills",
		},
	})
}

func (s *Server) handleSkillsAssimilate(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "skills.assimilate")
}

func (s *Server) handleSkillMutation(w http.ResponseWriter, r *http.Request, procedure string, fallback func(map[string]any) (any, error)) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), procedure, payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    procedure,
			},
		})
		return
	}

	fallbackResult, fallbackErr := fallback(payload)
	if fallbackErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": fallbackErr.Error(), "detail": fallbackErr.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    fallbackResult,
		"bridge": map[string]any{
			"fallback":  "go-local-skills",
			"procedure": procedure,
			"reason":    "upstream unavailable; applying local skill mutation",
		},
	})
}

func (s *Server) handleWorkflowList(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "workflow.list", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "workflow.list",
			},
		})
		return
	}

	writeJSON(w, http.StatusServiceUnavailable, map[string]any{
		"success": false,
		"error":   "Workflow definitions are unavailable: upstream workflow engine is unavailable and the local workflow engine is not initialized.",
		"data":    []map[string]any{},
		"bridge": map[string]any{
			"fallback":  "go-local-workflow",
			"procedure": "workflow.list",
			"reason":    "upstream unavailable; workflow engine is not initialized",
		},
	})
}

func (s *Server) handleWorkflowGraph(w http.ResponseWriter, r *http.Request) {
	workflowID := strings.TrimSpace(r.URL.Query().Get("workflowId"))
	if workflowID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing workflowId query parameter"})
		return
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "workflow.getGraph", map[string]any{"workflowId": workflowID}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "workflow.getGraph",
			},
		})
		return
	}

	writeJSON(w, http.StatusServiceUnavailable, map[string]any{
		"success": false,
		"error":   "Workflow graph is unavailable: upstream workflow engine is unavailable and the local workflow engine is not initialized.",
		"data": map[string]any{
			"nodes": []map[string]any{},
			"edges": []map[string]any{},
		},
		"bridge": map[string]any{
			"fallback":  "go-local-workflow",
			"procedure": "workflow.getGraph",
			"reason":    "upstream unavailable; workflow engine is not initialized",
		},
	})
}

func (s *Server) handleWorkflowStart(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "workflow.start")
}

func (s *Server) handleWorkflowExecutions(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "workflow.listExecutions", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "workflow.listExecutions",
			},
		})
		return
	}

	writeJSON(w, http.StatusServiceUnavailable, map[string]any{
		"success": false,
		"error":   "Workflow executions are unavailable: upstream workflow engine is unavailable and the local workflow engine is not initialized.",
		"data":    []map[string]any{},
		"bridge": map[string]any{
			"fallback":  "go-local-workflow",
			"procedure": "workflow.listExecutions",
			"reason":    "upstream unavailable; workflow engine is not initialized",
		},
	})
}

func (s *Server) handleWorkflowExecution(w http.ResponseWriter, r *http.Request) {
	executionID := strings.TrimSpace(r.URL.Query().Get("executionId"))
	if executionID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing executionId query parameter"})
		return
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "workflow.getExecution", map[string]any{"executionId": executionID}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "workflow.getExecution",
			},
		})
		return
	}

	writeJSON(w, http.StatusServiceUnavailable, map[string]any{
		"success": false,
		"error":   "Workflow execution is unavailable: upstream workflow engine is unavailable and the local workflow engine is not initialized.",
		"data":    nil,
		"bridge": map[string]any{
			"fallback":  "go-local-workflow",
			"procedure": "workflow.getExecution",
			"reason":    "upstream unavailable; workflow engine is not initialized",
		},
	})
}

func (s *Server) handleWorkflowHistory(w http.ResponseWriter, r *http.Request) {
	executionID := strings.TrimSpace(r.URL.Query().Get("executionId"))
	if executionID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing executionId query parameter"})
		return
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "workflow.getHistory", map[string]any{"executionId": executionID}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "workflow.getHistory",
			},
		})
		return
	}

	writeJSON(w, http.StatusServiceUnavailable, map[string]any{
		"success": false,
		"error":   "Workflow history is unavailable: upstream workflow engine is unavailable and the local workflow engine is not initialized.",
		"data":    []map[string]any{},
		"bridge": map[string]any{
			"fallback":  "go-local-workflow",
			"procedure": "workflow.getHistory",
			"reason":    "upstream unavailable; workflow engine is not initialized",
		},
	})
}

func (s *Server) handleWorkflowResume(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "workflow.resume")
}

func (s *Server) handleWorkflowPause(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "workflow.pause")
}

func (s *Server) handleWorkflowApprove(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "workflow.approve")
}

func (s *Server) handleWorkflowReject(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "workflow.reject")
}

func (s *Server) handleWorkflowCanvases(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "workflow.listCanvases", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "workflow.listCanvases",
			},
		})
		return
	}

	workflows, fallbackErr := s.localWorkflowCanvases()
	if fallbackErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": fallbackErr.Error(), "detail": fallbackErr.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    workflows,
		"bridge": map[string]any{
			"fallback":  "go-local-workflows-db",
			"procedure": "workflow.listCanvases",
			"reason":    "upstream unavailable; using local metamcp workflow canvases",
		},
	})
}

func (s *Server) handleWorkflowCanvas(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing id query parameter"})
		return
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "workflow.loadCanvas", map[string]any{"id": id}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "workflow.loadCanvas",
			},
		})
		return
	}

	workflow, fallbackErr := s.localWorkflowCanvas(id)
	if fallbackErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": fallbackErr.Error(), "detail": fallbackErr.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    workflow,
		"bridge": map[string]any{
			"fallback":  "go-local-workflows-db",
			"procedure": "workflow.loadCanvas",
			"reason":    "upstream unavailable; using local metamcp workflow canvas",
		},
	})
}

func (s *Server) handleWorkflowCanvasSave(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "workflow.saveCanvas")
}

func (s *Server) handleSymbolsList(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "symbols.list", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "symbols.list",
			},
		})
		return
	}

	s.writeSymbolsEmptyFallback(w, "symbols.list", "upstream unavailable; symbol pins are not initialized")
}

func (s *Server) handleSymbolsFind(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("query"))
	if query == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing query query parameter"})
		return
	}
	payload := map[string]any{"query": query}
	if limit := strings.TrimSpace(r.URL.Query().Get("limit")); limit != "" {
		if parsed, err := strconv.Atoi(limit); err == nil {
			payload["limit"] = parsed
		}
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "symbols.find", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "symbols.find",
			},
		})
		return
	}

	s.writeSymbolsEmptyFallback(w, "symbols.find", "upstream unavailable; symbol search is not initialized")
}

func (s *Server) handleSymbolsPin(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "symbols.pin")
}

func (s *Server) handleSymbolsUnpin(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "symbols.unpin")
}

func (s *Server) handleSymbolsUpdatePriority(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "symbols.updatePriority")
}

func (s *Server) handleSymbolsAddNotes(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "symbols.addNotes")
}

func (s *Server) handleSymbolsClear(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "symbols.clear")
}

func (s *Server) handleSymbolsForFile(w http.ResponseWriter, r *http.Request) {
	filePath := strings.TrimSpace(r.URL.Query().Get("filePath"))
	if filePath == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing filePath query parameter"})
		return
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "symbols.forFile", map[string]any{"filePath": filePath}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "symbols.forFile",
			},
		})
		return
	}

	s.writeSymbolsEmptyFallback(w, "symbols.forFile", "upstream unavailable; symbol pins are not initialized")
}

func (s *Server) writeSymbolsEmptyFallback(w http.ResponseWriter, procedure string, reason string) {
	writeJSON(w, http.StatusServiceUnavailable, map[string]any{
		"success": false,
		"error":   "Symbol data is unavailable: upstream symbols service is unavailable and the local symbol index is not initialized.",
		"data":    []map[string]any{},
		"bridge": map[string]any{
			"fallback":  "go-local-symbols",
			"procedure": procedure,
			"reason":    reason,
		},
	})
}

func (s *Server) handleLSPFindSymbol(w http.ResponseWriter, r *http.Request) {
	filePath := strings.TrimSpace(r.URL.Query().Get("filePath"))
	symbolName := strings.TrimSpace(r.URL.Query().Get("symbolName"))
	if filePath == "" || symbolName == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing filePath or symbolName query parameter"})
		return
	}
	s.handleTRPCBridgeCall(w, r, http.MethodGet, "lsp.findSymbol", map[string]any{"filePath": filePath, "symbolName": symbolName})
}

func (s *Server) handleLSPFindReferences(w http.ResponseWriter, r *http.Request) {
	filePath := strings.TrimSpace(r.URL.Query().Get("filePath"))
	line, lineErr := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("line")))
	character, charErr := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("character")))
	if filePath == "" || lineErr != nil || charErr != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing or invalid filePath, line, or character query parameter"})
		return
	}
	s.handleTRPCBridgeCall(w, r, http.MethodGet, "lsp.findReferences", map[string]any{"filePath": filePath, "line": line, "character": character})
}

func (s *Server) handleLSPGetSymbols(w http.ResponseWriter, r *http.Request) {
	filePath := strings.TrimSpace(r.URL.Query().Get("filePath"))
	if filePath == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing filePath query parameter"})
		return
	}
	s.handleTRPCBridgeCall(w, r, http.MethodGet, "lsp.getSymbols", map[string]any{"filePath": filePath})
}

func (s *Server) handleLSPSearchSymbols(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("query"))
	if query == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing query query parameter"})
		return
	}
	s.handleTRPCBridgeCall(w, r, http.MethodGet, "lsp.searchSymbols", map[string]any{"query": query})
}

func (s *Server) handleLSPIndexProject(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "lsp.indexProject")
}

func (s *Server) handleAPIKeysList(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "apiKeys.list", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "apiKeys.list",
			},
		})
		return
	}

	apiKeys, fallbackErr := s.localAPIKeys()
	if fallbackErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   fallbackErr.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    apiKeys,
		"bridge": map[string]any{
			"fallback":  "go-local-operator",
			"procedure": "apiKeys.list",
			"reason":    "upstream unavailable; using local metamcp workspace API key metadata",
		},
	})
}

func (s *Server) handleAPIKeysGet(w http.ResponseWriter, r *http.Request) {
	uuid := strings.TrimSpace(r.URL.Query().Get("uuid"))
	if uuid == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing uuid query parameter"})
		return
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "apiKeys.get", map[string]any{"uuid": uuid}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "apiKeys.get",
			},
		})
		return
	}

	apiKey, fallbackErr := s.localAPIKey(uuid)
	if fallbackErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"success": false,
			"error":   fallbackErr.Error(),
			"detail":  fallbackErr.Error(),
		})
		return
	}
	if apiKey == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"success": false,
			"error":   "API key unavailable",
			"detail":  "upstream unavailable; API key was not found in local metamcp workspace metadata",
			"bridge": map[string]any{
				"fallback":  "go-local-policy-db",
				"procedure": "apiKeys.get",
				"reason":    "upstream unavailable; API key was not found in local metamcp workspace metadata",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    apiKey,
		"bridge": map[string]any{
			"fallback":  "go-local-policy-db",
			"procedure": "apiKeys.get",
			"reason":    "upstream unavailable; using local metamcp api key record",
		},
	})
}

func (s *Server) handleAPIKeysCreate(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "apiKeys.create")
}

func (s *Server) handleAPIKeysUpdate(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "apiKeys.update")
}

func (s *Server) handleAPIKeysDelete(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "apiKeys.delete")
}

func (s *Server) handleAPIKeysValidate(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "apiKeys.validate")
}

func (s *Server) handleAuditList(w http.ResponseWriter, r *http.Request) {
	payload := map[string]any{}
	limit := 50
	if rawLimit := strings.TrimSpace(r.URL.Query().Get("limit")); rawLimit != "" {
		if parsed, err := strconv.Atoi(rawLimit); err == nil {
			payload["limit"] = parsed
			if parsed > 0 {
				limit = parsed
			}
		}
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "audit.list", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "audit.list",
			},
		})
		return
	}

	logs, fallbackErr := s.localAuditLogs(localAuditFilter{limit: limit})
	if fallbackErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": fallbackErr.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    logs,
		"bridge": map[string]any{
			"fallback":  "go-local-audit",
			"procedure": "audit.list",
			"reason":    "upstream unavailable; using local file-backed audit log list",
		},
	})
}

func (s *Server) handleAuditQuery(w http.ResponseWriter, r *http.Request) {
	payload := map[string]any{}
	filter := localAuditFilter{limit: 100}
	if level := strings.TrimSpace(r.URL.Query().Get("level")); level != "" {
		payload["level"] = level
		filter.level = level
	}
	if agentID := strings.TrimSpace(r.URL.Query().Get("agentId")); agentID != "" {
		payload["agentId"] = agentID
		filter.agentID = agentID
	}
	if action := strings.TrimSpace(r.URL.Query().Get("action")); action != "" {
		payload["action"] = action
		filter.action = action
	}
	if limit := strings.TrimSpace(r.URL.Query().Get("limit")); limit != "" {
		if parsed, err := strconv.Atoi(limit); err == nil {
			payload["limit"] = parsed
			if parsed > 0 {
				filter.limit = parsed
			}
		}
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "audit.log", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "audit.log",
			},
		})
		return
	}

	logs, fallbackErr := s.localAuditLogs(filter)
	if fallbackErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": fallbackErr.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    logs,
		"bridge": map[string]any{
			"fallback":  "go-local-audit",
			"procedure": "audit.log",
			"reason":    "upstream unavailable; using local file-backed audit query results",
		},
	})
}

func (s *Server) handleSavedScriptsList(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "savedScripts.list", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "savedScripts.list",
			},
		})
		return
	}

	scripts, fallbackErr := s.localSavedScripts()
	if fallbackErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   fallbackErr.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    scripts,
		"bridge": map[string]any{
			"fallback":  "go-local-operator",
			"procedure": "savedScripts.list",
			"reason":    "upstream unavailable; using local saved scripts from borg config",
		},
	})
}

func (s *Server) handleSavedScriptsGet(w http.ResponseWriter, r *http.Request) {
	uuid := strings.TrimSpace(r.URL.Query().Get("uuid"))
	if uuid == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing uuid query parameter"})
		return
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "savedScripts.get", map[string]any{"uuid": uuid}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "savedScripts.get",
			},
		})
		return
	}

	scripts, fallbackErr := s.localSavedScripts()
	if fallbackErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   fallbackErr.Error(),
		})
		return
	}

	for _, script := range scripts {
		if stringValue(script["uuid"]) == uuid {
			writeJSON(w, http.StatusOK, map[string]any{
				"success": true,
				"data":    script,
				"bridge": map[string]any{
					"fallback":  "go-local-operator",
					"procedure": "savedScripts.get",
					"reason":    "upstream unavailable; using local saved script from borg config",
				},
			})
			return
		}
	}

	writeJSON(w, http.StatusServiceUnavailable, map[string]any{
		"success": false,
		"error":   "saved script unavailable",
		"detail":  "upstream unavailable; saved script was not found in local borg config",
		"bridge": map[string]any{
			"fallback":  "go-local-operator",
			"procedure": "savedScripts.get",
			"reason":    "upstream unavailable; saved script was not found in local borg config",
		},
	})
}

func (s *Server) handleSavedScriptsCreate(w http.ResponseWriter, r *http.Request) {
	var payload map[string]any
	if err := decodeJSONBody(r, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "savedScripts.create", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "savedScripts.create",
			},
		})
		return
	}

	script, fallbackErr := s.localCreateSavedScript(payload)
	if fallbackErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": fallbackErr.Error(), "detail": fallbackErr.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    script,
		"bridge": map[string]any{
			"fallback":  "go-local-operator",
			"procedure": "savedScripts.create",
			"reason":    "upstream unavailable; saved script to local Borg config",
		},
	})
}

func (s *Server) handleSavedScriptsUpdate(w http.ResponseWriter, r *http.Request) {
	var payload map[string]any
	if err := decodeJSONBody(r, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "savedScripts.update", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "savedScripts.update",
			},
		})
		return
	}

	updateResult, fallbackErr := s.localUpdateSavedScript(payload)
	if fallbackErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": fallbackErr.Error(), "detail": fallbackErr.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    updateResult,
		"bridge": map[string]any{
			"fallback":  "go-local-operator",
			"procedure": "savedScripts.update",
			"reason":    "upstream unavailable; updated saved script in local Borg config",
		},
	})
}

func (s *Server) handleSavedScriptsDelete(w http.ResponseWriter, r *http.Request) {
	var payload map[string]any
	if err := decodeJSONBody(r, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "savedScripts.delete", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "savedScripts.delete",
			},
		})
		return
	}

	deleteResult, fallbackErr := s.localDeleteSavedScript(strings.TrimSpace(stringValue(payload["uuid"])))
	if fallbackErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": fallbackErr.Error(), "detail": fallbackErr.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    deleteResult,
		"bridge": map[string]any{
			"fallback":  "go-local-operator",
			"procedure": "savedScripts.delete",
			"reason":    "upstream unavailable; deleted saved script from local Borg config",
		},
	})
}

func (s *Server) handleSavedScriptsExecute(w http.ResponseWriter, r *http.Request) {
	var payload map[string]any
	if err := decodeJSONBody(r, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "savedScripts.execute", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "savedScripts.execute",
			},
		})
		return
	}

	executionResult, fallbackErr := s.localExecuteSavedScript(strings.TrimSpace(stringValue(payload["uuid"])))
	if fallbackErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"success": false, "error": fallbackErr.Error(), "detail": fallbackErr.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    executionResult,
		"bridge": map[string]any{
			"fallback":  "go-local-operator",
			"procedure": "savedScripts.execute",
			"reason":    "upstream unavailable; executed saved script through local node runtime without TypeScript code-executor services",
		},
	})
}

func (s *Server) handleLinksBacklogList(w http.ResponseWriter, r *http.Request) {
	payload := map[string]any{}
	limitValue := 50
	offsetValue := 0
	searchValue := ""
	sourceValue := ""
	statusValue := ""
	clusterIDValue := ""
	showDuplicatesValue := false
	if limit := strings.TrimSpace(r.URL.Query().Get("limit")); limit != "" {
		if parsed, err := strconv.Atoi(limit); err == nil {
			limitValue = parsed
			payload["limit"] = parsed
		}
	}
	if offset := strings.TrimSpace(r.URL.Query().Get("offset")); offset != "" {
		if parsed, err := strconv.Atoi(offset); err == nil {
			offsetValue = parsed
			payload["offset"] = parsed
		}
	}
	if search := strings.TrimSpace(r.URL.Query().Get("search")); search != "" {
		searchValue = search
		payload["search"] = search
	}
	if source := strings.TrimSpace(r.URL.Query().Get("source")); source != "" {
		sourceValue = source
		payload["source"] = source
	}
	if status := strings.TrimSpace(r.URL.Query().Get("research_status")); status != "" {
		statusValue = status
		payload["research_status"] = status
	}
	if clusterID := strings.TrimSpace(r.URL.Query().Get("cluster_id")); clusterID != "" {
		clusterIDValue = clusterID
		payload["cluster_id"] = clusterID
	}
	if showDuplicates := strings.TrimSpace(r.URL.Query().Get("show_duplicates")); showDuplicates != "" {
		showDuplicatesValue = strings.EqualFold(showDuplicates, "true") || showDuplicates == "1"
		payload["show_duplicates"] = showDuplicatesValue
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "linksBacklog.list", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "linksBacklog.list",
			},
		})
		return
	}

	listPayload, fallbackErr := s.localLinksBacklogList(limitValue, offsetValue, searchValue, sourceValue, statusValue, clusterIDValue, showDuplicatesValue)
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
		"data":    listPayload,
		"bridge": map[string]any{
			"fallback":  "go-local-links-db",
			"procedure": "linksBacklog.list",
			"reason":    "upstream unavailable; using local metamcp links backlog list",
		},
	})
}

func (s *Server) handleLinksBacklogStats(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "linksBacklog.stats", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "linksBacklog.stats",
			},
		})
		return
	}

	stats, fallbackErr := s.localLinksBacklogStats()
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
		"data":    stats,
		"bridge": map[string]any{
			"fallback":  "go-local-links-db",
			"procedure": "linksBacklog.stats",
			"reason":    "upstream unavailable; using local metamcp links backlog aggregates",
		},
	})
}

func (s *Server) handleLinksBacklogGet(w http.ResponseWriter, r *http.Request) {
	uuid := strings.TrimSpace(r.URL.Query().Get("uuid"))
	if uuid == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing uuid query parameter"})
		return
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "linksBacklog.get", map[string]any{"uuid": uuid}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "linksBacklog.get",
			},
		})
		return
	}

	item, fallbackErr := s.localLinksBacklogItem(uuid)
	if fallbackErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"success": false,
			"error":   fallbackErr.Error(),
			"detail":  fallbackErr.Error(),
		})
		return
	}
	if item == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"success": false,
			"error":   "links backlog item unavailable",
			"detail":  "upstream unavailable; links backlog item was not found in local metamcp links backlog",
			"bridge": map[string]any{
				"fallback":  "go-local-links-db",
				"procedure": "linksBacklog.get",
				"reason":    "upstream unavailable; links backlog item was not found in local metamcp links backlog",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    item,
		"bridge": map[string]any{
			"fallback":  "go-local-links-db",
			"procedure": "linksBacklog.get",
			"reason":    "upstream unavailable; using local metamcp links backlog record",
		},
	})
}

func (s *Server) handleLinksBacklogSync(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "linksBacklog.syncFromBobbyBookmarks", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "linksBacklog.syncFromBobbyBookmarks",
			},
		})
		return
	}

	db, err := sql.Open("sqlite", s.localMetaMCPDBPath())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}
	defer db.Close()

	res, fallbackErr := hsync.SyncBobbyBookmarks(r.Context(), s.localMetaMCPDBPath(), "https://bobbybookmarks.com", 100, false, false)
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
		"data":    res,
		"bridge": map[string]any{
			"fallback":  "go-local-links-sync",
			"procedure": "linksBacklog.syncFromBobbyBookmarks",
			"reason":    "upstream unavailable; executing native Go links backlog sync",
		},
	})
}

func (s *Server) handleInfrastructureStatus(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "infrastructure.getInfrastructureStatus", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "infrastructure.getInfrastructureStatus",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    localInfrastructureStatus(s.cfg.WorkspaceRoot),
		"bridge": map[string]any{
			"fallback":  "go-local-infrastructure",
			"procedure": "infrastructure.getInfrastructureStatus",
			"reason":    "upstream unavailable; using local infrastructure binary/config visibility",
		},
	})
}

func (s *Server) handleInfrastructureDoctor(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "infrastructure.runDoctor", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "infrastructure.runDoctor",
			},
		})
		return
	}

	res, fallbackErr := hsync.RunInfrastructureDoctor(s.cfg.WorkspaceRoot)
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
		"data":    res,
		"bridge": map[string]any{
			"fallback":  "go-local-infrastructure",
			"procedure": "infrastructure.runDoctor",
			"reason":    "upstream unavailable; executing native Go infrastructure doctor",
		},
	})
}

func (s *Server) handleInfrastructureApply(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "infrastructure.applyConfigurations", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "infrastructure.applyConfigurations",
			},
		})
		return
	}

	res, fallbackErr := hsync.ApplyInfrastructureConfigurations(s.cfg.WorkspaceRoot)
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
		"data":    res,
		"bridge": map[string]any{
			"fallback":  "go-local-infrastructure",
			"procedure": "infrastructure.applyConfigurations",
			"reason":    "upstream unavailable; executing native Go infrastructure apply",
		},
	})
}

func (s *Server) handleExpertResearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	var payload struct {
		Topic string `json:"topic"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	// Try upstream first
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "expert.research", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "expert.research",
			},
		})
		return
	}

	// Fallback to local
	res, fallbackErr := s.expertManager.ExpertResearch(r.Context(), payload.Topic)
	if fallbackErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": fallbackErr.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    res,
		"bridge": map[string]any{
			"fallback": "go-local-expert",
		},
	})
}

func (s *Server) handleExpertCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	var payload struct {
		Instruction string `json:"instruction"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	// Try upstream first
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "expert.code", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "expert.code",
			},
		})
		return
	}

	// Fallback to local
	res, fallbackErr := s.expertManager.ExpertCode(r.Context(), payload.Instruction)
	if fallbackErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": fallbackErr.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    res,
		"bridge": map[string]any{
			"fallback": "go-local-expert",
		},
	})
}

func (s *Server) handleExpertStatus(w http.ResponseWriter, r *http.Request) {
	// Try upstream first
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "expert.getStatus", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "expert.getStatus",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"researcher": "online",
			"coder":      "online",
		},
		"bridge": map[string]any{
			"fallback": "go-local-expert",
		},
	})
}

func (s *Server) getCoderStatus() string {
	if s.coderAgent != nil {
		return "active"
	}
	return "offline"
}

func (s *Server) handlePoliciesList(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "policies.list", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "policies.list",
			},
		})
		return
	}

	policies, fallbackErr := s.localPolicies()
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
		"data":    policies,
		"bridge": map[string]any{
			"fallback":  "go-local-policy-db",
			"procedure": "policies.list",
			"reason":    "upstream unavailable; using local metamcp policy records",
		},
	})
}

func (s *Server) handlePoliciesGet(w http.ResponseWriter, r *http.Request) {
	uuid := strings.TrimSpace(r.URL.Query().Get("uuid"))
	if uuid == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing uuid query parameter"})
		return
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "policies.get", map[string]any{"uuid": uuid}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "policies.get",
			},
		})
		return
	}

	policy, fallbackErr := s.localPolicy(uuid)
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
		"data":    policy,
		"bridge": map[string]any{
			"fallback":  "go-local-policy-db",
			"procedure": "policies.get",
			"reason":    "upstream unavailable; using local metamcp policy record",
		},
	})
}

func (s *Server) handlePoliciesCreate(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "policies.create")
}

func (s *Server) handlePoliciesUpdate(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "policies.update")
}

func (s *Server) handlePoliciesDelete(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "policies.delete")
}

func (s *Server) handleSecretsList(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "secrets.list", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "secrets.list",
			},
		})
		return
	}

	secrets, fallbackErr := s.localSecrets()
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
		"data":    secrets,
		"bridge": map[string]any{
			"fallback":  "go-local-policy-db",
			"procedure": "secrets.list",
			"reason":    "upstream unavailable; using local metamcp workspace secrets metadata",
		},
	})
}

func (s *Server) handleSecretsSet(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "secrets.set")
}

func (s *Server) handleSecretsDelete(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "secrets.delete")
}

func (s *Server) handleMarketplaceList(w http.ResponseWriter, r *http.Request) {
	payload := map[string]any{}
	filterValue := ""
	if filter := strings.TrimSpace(r.URL.Query().Get("filter")); filter != "" {
		filterValue = filter
		payload["filter"] = filter
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "marketplace.list", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "marketplace.list",
			},
		})
		return
	}

	entries, fallbackErr := s.localMarketplaceList(filterValue)
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
		"data":    entries,
		"bridge": map[string]any{
			"fallback":  "go-local-marketplace",
			"procedure": "marketplace.list",
			"reason":    "upstream unavailable; using local marketplace registries and install-state checks",
		},
	})
}

func (s *Server) handleMarketplaceInstall(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "marketplace.install", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "marketplace.install",
			},
		})
		return
	}

	var payload struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid request body"})
		return
	}

	message, fallbackErr := mcp.InstallMarketplaceEntry(payload.ID)
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
		"data":    message,
		"bridge": map[string]any{
			"fallback":  "go-local-marketplace",
			"procedure": "marketplace.install",
			"reason":    "upstream unavailable; executing native Go marketplace install",
		},
	})
}

func (s *Server) handleMarketplacePublish(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "marketplace.publish")
}

func (s *Server) handleCatalogList(w http.ResponseWriter, r *http.Request) {
	payload := map[string]any{}
	limitValue := 50
	offsetValue := 0
	searchValue := ""
	statusValue := ""
	transportValue := ""
	installMethodValue := ""
	if limit := strings.TrimSpace(r.URL.Query().Get("limit")); limit != "" {
		if parsed, err := strconv.Atoi(limit); err == nil {
			limitValue = parsed
			payload["limit"] = parsed
		}
	}
	if offset := strings.TrimSpace(r.URL.Query().Get("offset")); offset != "" {
		if parsed, err := strconv.Atoi(offset); err == nil {
			offsetValue = parsed
			payload["offset"] = parsed
		}
	}
	if search := strings.TrimSpace(r.URL.Query().Get("search")); search != "" {
		searchValue = search
		payload["search"] = search
	}
	if status := strings.TrimSpace(r.URL.Query().Get("status")); status != "" {
		statusValue = status
		payload["status"] = status
	}
	if transport := strings.TrimSpace(r.URL.Query().Get("transport")); transport != "" {
		transportValue = transport
		payload["transport"] = transport
	}
	if installMethod := strings.TrimSpace(r.URL.Query().Get("install_method")); installMethod != "" {
		installMethodValue = installMethod
		payload["install_method"] = installMethod
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "catalog.list", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "catalog.list",
			},
		})
		return
	}

	listPayload, fallbackErr := s.localCatalogList(limitValue, offsetValue, searchValue, statusValue, transportValue, installMethodValue)
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
		"data":    listPayload,
		"bridge": map[string]any{
			"fallback":  "go-local-published-catalog-db",
			"procedure": "catalog.list",
			"reason":    "upstream unavailable; using local metamcp published catalog list",
		},
	})
}

func (s *Server) handleCatalogGet(w http.ResponseWriter, r *http.Request) {
	uuid := strings.TrimSpace(r.URL.Query().Get("uuid"))
	if uuid == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing uuid query parameter"})
		return
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "catalog.get", map[string]any{"uuid": uuid}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "catalog.get",
			},
		})
		return
	}

	payload, fallbackErr := s.localCatalogGet(uuid)
	if fallbackErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"success": false,
			"error":   fallbackErr.Error(),
			"detail":  fallbackErr.Error(),
		})
		return
	}
	if payload == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"success": false,
			"error":   "catalog entry unavailable",
			"detail":  "upstream unavailable; catalog entry was not found in local metamcp published catalog",
			"bridge": map[string]any{
				"fallback":  "go-local-published-catalog-db",
				"procedure": "catalog.get",
				"reason":    "upstream unavailable; catalog entry was not found in local metamcp published catalog",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    payload,
		"bridge": map[string]any{
			"fallback":  "go-local-published-catalog-db",
			"procedure": "catalog.get",
			"reason":    "upstream unavailable; using local metamcp published catalog records",
		},
	})
}

func (s *Server) handleCatalogRuns(w http.ResponseWriter, r *http.Request) {
	serverUUID := strings.TrimSpace(r.URL.Query().Get("server_uuid"))
	if serverUUID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing server_uuid query parameter"})
		return
	}
	payload := map[string]any{"server_uuid": serverUUID}
	limitValue := 10
	if limit := strings.TrimSpace(r.URL.Query().Get("limit")); limit != "" {
		if parsed, err := strconv.Atoi(limit); err == nil {
			limitValue = parsed
			payload["limit"] = parsed
		}
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "catalog.listRuns", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "catalog.listRuns",
			},
		})
		return
	}

	runs, fallbackErr := s.localCatalogRuns(serverUUID, limitValue)
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
		"data":    runs,
		"bridge": map[string]any{
			"fallback":  "go-local-published-catalog-db",
			"procedure": "catalog.listRuns",
			"reason":    "upstream unavailable; using local metamcp published catalog validation runs",
		},
	})
}

func (s *Server) handleCatalogIngest(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "catalog.triggerIngestion")
}

func (s *Server) handleCatalogValidate(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "catalog.triggerValidation")
}

func (s *Server) handleCatalogInstall(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "catalog.installFromRecipe")
}

func (s *Server) handleCatalogValidateBatch(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "catalog.triggerBatchValidation")
}

func (s *Server) handleCatalogStats(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "catalog.stats", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "catalog.stats",
			},
		})
		return
	}

	stats, fallbackErr := s.localCatalogStats()
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
		"data":    stats,
		"bridge": map[string]any{
			"fallback":  "go-local-published-catalog-db",
			"procedure": "catalog.stats",
			"reason":    "upstream unavailable; using local metamcp published catalog aggregates",
		},
	})
}

func (s *Server) handleCatalogLinkedServers(w http.ResponseWriter, r *http.Request) {
	publishedServerUUID := strings.TrimSpace(r.URL.Query().Get("published_server_uuid"))
	if publishedServerUUID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing published_server_uuid query parameter"})
		return
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "catalog.listLinkedServers", map[string]any{"published_server_uuid": publishedServerUUID}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "catalog.listLinkedServers",
			},
		})
		return
	}

	servers, fallbackErr := s.localCatalogLinkedServers(publishedServerUUID)
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
		"data":    servers,
		"bridge": map[string]any{
			"fallback":  "go-local-published-catalog-db",
			"procedure": "catalog.listLinkedServers",
			"reason":    "upstream unavailable; using local metamcp linked managed servers",
		},
	})
}

func (s *Server) handleOAuthClientCreate(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "oauth.clients.create")
}

func (s *Server) handleOAuthClientGet(w http.ResponseWriter, r *http.Request) {
	clientID := strings.TrimSpace(r.URL.Query().Get("clientId"))
	if clientID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing clientId query parameter"})
		return
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "oauth.clients.get", map[string]any{"clientId": clientID}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "oauth.clients.get",
			},
		})
		return
	}

	client, fallbackErr := s.localOAuthClient(clientID)
	if fallbackErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"success": false,
			"error":   fallbackErr.Error(),
			"detail":  fallbackErr.Error(),
		})
		return
	}
	if client == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"success": false,
			"error":   "OAuth client unavailable",
			"detail":  "upstream unavailable; OAuth client was not found in local metamcp oauth clients",
			"bridge": map[string]any{
				"fallback":  "go-local-oauth-clients-db",
				"procedure": "oauth.clients.get",
				"reason":    "upstream unavailable; OAuth client was not found in local metamcp oauth clients",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    client,
		"bridge": map[string]any{
			"fallback":  "go-local-oauth-clients-db",
			"procedure": "oauth.clients.get",
			"reason":    "upstream unavailable; using local metamcp oauth client record",
		},
	})
}

func (s *Server) handleOAuthSessionUpsert(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "oauth.sessions.upsert")
}

func (s *Server) handleOAuthSessionGetByServer(w http.ResponseWriter, r *http.Request) {
	serverUUID := strings.TrimSpace(r.URL.Query().Get("mcpServerUuid"))
	if serverUUID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing mcpServerUuid query parameter"})
		return
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "oauth.sessions.getByServer", map[string]any{"mcpServerUuid": serverUUID}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "oauth.sessions.getByServer",
			},
		})
		return
	}

	session, fallbackErr := s.localOAuthSessionByServer(serverUUID)
	if fallbackErr != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"success": false,
			"error":   fallbackErr.Error(),
			"detail":  fallbackErr.Error(),
		})
		return
	}
	if session == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"success": false,
			"error":   "OAuth session unavailable",
			"detail":  "upstream unavailable; OAuth session was not found in local metamcp oauth sessions",
			"bridge": map[string]any{
				"fallback":  "go-local-oauth-sessions-db",
				"procedure": "oauth.sessions.getByServer",
				"reason":    "upstream unavailable; OAuth session was not found in local metamcp oauth sessions",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    session,
		"bridge": map[string]any{
			"fallback":  "go-local-oauth-sessions-db",
			"procedure": "oauth.sessions.getByServer",
			"reason":    "upstream unavailable; using local metamcp oauth session record",
		},
	})
}

func (s *Server) handleOAuthExchange(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "oauth.exchange")
}

func (s *Server) handleResearchConduct(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "research.conduct")
}

func (s *Server) handleResearchIngest(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "research.ingest")
}

func (s *Server) handleResearchRecursive(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "research.recursiveResearch")
}

func (s *Server) handleResearchQueries(w http.ResponseWriter, r *http.Request) {
	topic := strings.TrimSpace(r.URL.Query().Get("topic"))
	if topic == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing topic query parameter"})
		return
	}
	payload := map[string]any{"topic": topic}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "research.generateQueries", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "research.generateQueries",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"queries": []string{topic},
		},
		"bridge": map[string]any{
			"fallback":  "go-local-research",
			"procedure": "research.generateQueries",
			"reason":    "upstream unavailable; using topic-as-query research fallback",
		},
	})
}

func (s *Server) handleResearchQueue(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "research.ingestionQueue", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "research.ingestionQueue",
			},
		})
		return
	}

	data, fallbackErr := s.localResearchQueue()
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
		"data":    data,
		"bridge": map[string]any{
			"fallback":  "go-local-research",
			"procedure": "research.ingestionQueue",
			"reason":    "upstream unavailable; using local research queue files",
		},
	})
}

func (s *Server) handleResearchRetryFailed(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "research.retryFailed")
}

func (s *Server) handleResearchRetryAllFailed(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "research.retryAllFailed")
}

func (s *Server) handleResearchEnqueuePending(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "research.enqueuePending")
}

func (s *Server) handlePulseEvents(w http.ResponseWriter, r *http.Request) {
	payload := map[string]any{}
	if limit := strings.TrimSpace(r.URL.Query().Get("limit")); limit != "" {
		if parsed, err := strconv.Atoi(limit); err == nil {
			payload["limit"] = parsed
		}
	}
	if afterTimestamp := strings.TrimSpace(r.URL.Query().Get("afterTimestamp")); afterTimestamp != "" {
		if parsed, err := strconv.ParseInt(afterTimestamp, 10, 64); err == nil {
			payload["afterTimestamp"] = parsed
		}
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "pulse.getLatestEvents", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "pulse.getLatestEvents",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    []map[string]any{},
		"bridge": map[string]any{
			"fallback":  "go-local-pulse",
			"procedure": "pulse.getLatestEvents",
			"reason":    "upstream unavailable; using local empty pulse event history",
		},
	})
}

func (s *Server) handlePulseStatus(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "pulse.getSystemStatus", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "pulse.getSystemStatus",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"status":            "offline",
			"uptime":            0,
			"agents":            []string{},
			"memoryInitialized": false,
		},
		"bridge": map[string]any{
			"fallback":  "go-local-pulse",
			"procedure": "pulse.getSystemStatus",
			"reason":    "upstream unavailable; using local offline pulse status",
		},
	})
}

func (s *Server) handlePulseProviders(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "pulse.checkLocalProviders", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "pulse.checkLocalProviders",
			},
		})
		return
	}

	statuses := providers.Snapshot()
	data := map[string]bool{
		"openai":     false,
		"anthropic":  false,
		"google":     false,
		"openrouter": false,
		"deepseek":   false,
		"xai":        false,
		"ollama":     false,
		"lmstudio":   false,
	}
	for _, status := range statuses {
		if _, ok := data[status.Provider]; ok {
			data[status.Provider] = status.Configured || status.Authenticated
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    data,
		"bridge": map[string]any{
			"fallback":  "go-local-provider-routing",
			"procedure": "pulse.checkLocalProviders",
			"reason":    "upstream unavailable; using local provider availability snapshot",
		},
	})
}

func (s *Server) handleSessionExport(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "sessionExport.export")
}

func (s *Server) handleSessionImport(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "sessionExport.import")
}

func (s *Server) handleSessionExportDetectFormat(w http.ResponseWriter, r *http.Request) {
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "sessionExport.detectFormat", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "sessionExport.detectFormat",
			},
		})
		return
	}

	raw, _ := payload["data"].(string)
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    localSessionExportFormatDetection(raw),
		"bridge": map[string]any{
			"fallback":  "go-local-session-export",
			"procedure": "sessionExport.detectFormat",
			"reason":    "upstream unavailable; using local session export format detection",
		},
	})
}

func (s *Server) handleSessionExportKnownFormats(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "sessionExport.knownFormats", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "sessionExport.knownFormats",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    sessionExportKnownFormats,
		"bridge": map[string]any{
			"fallback":  "go-local-session-export",
			"procedure": "sessionExport.knownFormats",
			"reason":    "upstream unavailable; using local known session export formats",
		},
	})
}

func (s *Server) handleSessionExportHistory(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "sessionExport.history", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "sessionExport.history",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    []map[string]any{},
		"bridge": map[string]any{
			"fallback":  "go-local-session-export",
			"procedure": "sessionExport.history",
			"reason":    "upstream unavailable; local session export history is empty",
		},
	})
}

func (s *Server) handleBrowserExtensionSaveMemory(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "browserExtension.saveMemory")
}

func (s *Server) handleBrowserExtensionParseDOM(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "browserExtension.parseDom")
}

func (s *Server) handleBrowserExtensionListMemories(w http.ResponseWriter, r *http.Request) {
	payload := map[string]any{}
	searchValue := strings.TrimSpace(r.URL.Query().Get("search"))
	tagValue := strings.TrimSpace(r.URL.Query().Get("tag"))
	limitValue := 50
	offsetValue := 0
	if searchValue != "" {
		payload["search"] = searchValue
	}
	if tagValue != "" {
		payload["tag"] = tagValue
	}
	if limit := strings.TrimSpace(r.URL.Query().Get("limit")); limit != "" {
		if parsed, err := strconv.Atoi(limit); err == nil {
			limitValue = parsed
			payload["limit"] = parsed
		}
	}
	if offset := strings.TrimSpace(r.URL.Query().Get("offset")); offset != "" {
		if parsed, err := strconv.Atoi(offset); err == nil {
			offsetValue = parsed
			payload["offset"] = parsed
		}
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "browserExtension.listMemories", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "browserExtension.listMemories",
			},
		})
		return
	}

	memories, fallbackErr := s.localBrowserExtensionMemories(searchValue, tagValue, limitValue, offsetValue)
	if fallbackErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   fallbackErr.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    memories,
		"bridge": map[string]any{
			"fallback":  "go-local-browser-memory",
			"procedure": "browserExtension.listMemories",
			"reason":    "upstream unavailable; using local browser memories from metamcp.db",
		},
	})
}

func (s *Server) handleBrowserExtensionDeleteMemory(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "browserExtension.deleteMemory")
}

func (s *Server) handleBrowserExtensionStats(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "browserExtension.stats", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "browserExtension.stats",
			},
		})
		return
	}

	stats, fallbackErr := s.localBrowserExtensionStats()
	if fallbackErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   fallbackErr.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    stats,
		"bridge": map[string]any{
			"fallback":  "go-local-browser-memory",
			"procedure": "browserExtension.stats",
			"reason":    "upstream unavailable; using local browser memory stats from metamcp.db",
		},
	})
}

func (s *Server) handleOpenWebUIStatus(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "openWebUI.getStatus", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "openWebUI.getStatus",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"status":          "active",
			"version":         "0.99.1",
			"connected_tools": 0,
			"message":         "Open-WebUI integration is initialized and ready.",
			"timestamp":       time.Now().UTC().Format(time.RFC3339),
		},
		"bridge": map[string]any{
			"fallback":  "go-local-status",
			"procedure": "openWebUI.getStatus",
			"reason":    "upstream unavailable; using local Open WebUI status defaults",
		},
	})
}

func (s *Server) handleOpenWebUIEmbedURL(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "openWebUI.getEmbedUrl", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "openWebUI.getEmbedUrl",
			},
		})
		return
	}

	url := strings.TrimSpace(os.Getenv("OPEN_WEBUI_URL"))
	if url == "" {
		url = "http://localhost:8080"
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"url": url,
		},
		"bridge": map[string]any{
			"fallback":  "go-local-status",
			"procedure": "openWebUI.getEmbedUrl",
			"reason":    "upstream unavailable; using local Open WebUI URL fallback",
		},
	})
}

func (s *Server) handleCodeModeStatus(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "codeMode.getStatus", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "codeMode.getStatus",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"enabled":   false,
			"toolCount": 0,
			"tools":     []map[string]any{},
			"reduction": map[string]any{
				"traditional":  0,
				"codeMode":     0,
				"reductionPct": 0,
			},
		},
		"bridge": map[string]any{
			"fallback":  "go-local-status",
			"procedure": "codeMode.getStatus",
			"reason":    "upstream unavailable; using local zero-state Code Mode status",
		},
	})
}

func (s *Server) handleCodeModeEnable(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "codeMode.enable")
}

func (s *Server) handleCodeModeDisable(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "codeMode.disable")
}

func (s *Server) handleCodeModeExecute(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "codeMode.execute")
}

func (s *Server) handleSubmoduleList(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "submodule.list", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "submodule.list",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    localSubmoduleList(s.cfg.WorkspaceRoot),
		"bridge": map[string]any{
			"fallback":  "go-local-submodules",
			"procedure": "submodule.list",
			"reason":    "upstream unavailable; using local .gitmodules submodule fallback",
		},
	})
}

func (s *Server) handleSubmoduleInstallDependencies(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "submodule.installDependencies")
}

func (s *Server) handleSubmoduleBuild(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "submodule.build")
}

func (s *Server) handleSubmoduleEnable(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "submodule.enable")
}

func (s *Server) handleSubmoduleCapabilities(w http.ResponseWriter, r *http.Request) {
	pathValue := strings.TrimSpace(r.URL.Query().Get("path"))
	if pathValue == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing path query parameter"})
		return
	}
	payload := map[string]any{"path": pathValue}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "submodule.detectCapabilities", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "submodule.detectCapabilities",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    localSubmoduleCapabilities(s.cfg.WorkspaceRoot, pathValue),
		"bridge": map[string]any{
			"fallback":  "go-local-submodules",
			"procedure": "submodule.detectCapabilities",
			"reason":    "upstream unavailable; using local submodule capability fallback",
		},
	})
}

func (s *Server) handleSuggestionsList(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "suggestions.list", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "suggestions.list",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    []map[string]any{},
		"bridge": map[string]any{
			"fallback":  "go-local-registry",
			"procedure": "suggestions.list",
			"reason":    "upstream unavailable; using local empty suggestions list",
		},
	})
}

func (s *Server) handleSuggestionsResolve(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "suggestions.resolve", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "suggestions.resolve",
			},
		})
		return
	}

	var payload struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid request body"})
		return
	}

	res, fallbackErr := hsync.ResolveSuggestion(payload.ID, payload.Status)
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
		"data":    res,
		"bridge": map[string]any{
			"fallback":  "go-local-suggestions",
			"procedure": "suggestions.resolve",
			"reason":    "upstream unavailable; executing native Go suggestion resolution",
		},
	})
}

func (s *Server) handleSuggestionsClear(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "suggestions.clearAll", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "suggestions.clearAll",
			},
		})
		return
	}

	res, fallbackErr := hsync.ClearAllSuggestions()
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
		"data":    res,
		"bridge": map[string]any{
			"fallback":  "go-local-suggestions",
			"procedure": "suggestions.clearAll",
			"reason":    "upstream unavailable; executing native Go suggestion clear",
		},
	})
}

func (s *Server) handlePlanMode(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		var result any
		upstreamBase, err := s.callUpstreamJSON(r.Context(), "plan.getMode", nil, &result)
		if err == nil {
			writeJSON(w, http.StatusOK, map[string]any{
				"success": true,
				"data":    result,
				"bridge": map[string]any{
					"upstreamBase": upstreamBase,
					"procedure":    "plan.getMode",
				},
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data": map[string]any{
				"mode": "PLAN",
			},
			"bridge": map[string]any{
				"fallback":  "go-local-plan",
				"procedure": "plan.getMode",
				"reason":    "upstream unavailable; plan mode is not persisted locally so defaulting to PLAN",
			},
		})
		return
	}
	s.handleTRPCBridgeBodyCall(w, r, "plan.setMode")
}

func (s *Server) handlePlanDiffs(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "plan.getDiffs", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "plan.getDiffs",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    s.localPlanDiffs(),
		"bridge": map[string]any{
			"fallback":  "go-local-plan",
			"procedure": "plan.getDiffs",
			"reason":    "upstream unavailable; using local sandbox diffs",
		},
	})
}

func (s *Server) handlePlanApproveDiff(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "plan.approveDiff")
}

func (s *Server) handlePlanRejectDiff(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "plan.rejectDiff")
}

func (s *Server) handlePlanApplyAll(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "plan.applyAll")
}

func (s *Server) handlePlanSummary(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "plan.getSummary", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "plan.getSummary",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    s.localPlanSummary(),
		"bridge": map[string]any{
			"fallback":  "go-local-plan",
			"procedure": "plan.getSummary",
			"reason":    "upstream unavailable; using local sandbox summary",
		},
	})
}

func (s *Server) handlePlanCheckpoints(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "plan.getCheckpoints", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "plan.getCheckpoints",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    s.localPlanCheckpoints(),
		"bridge": map[string]any{
			"fallback":  "go-local-plan",
			"procedure": "plan.getCheckpoints",
			"reason":    "upstream unavailable; using local sandbox checkpoints",
		},
	})
}

func (s *Server) handlePlanCreateCheckpoint(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "plan.createCheckpoint")
}

func (s *Server) handlePlanRollback(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "plan.rollback")
}

func (s *Server) handlePlanClear(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "plan.clear")
}

func (s *Server) handleKnowledgeGraph(w http.ResponseWriter, r *http.Request) {
	payload := map[string]any{}
	if query := strings.TrimSpace(r.URL.Query().Get("query")); query != "" {
		payload["query"] = query
	}
	if depth := strings.TrimSpace(r.URL.Query().Get("depth")); depth != "" {
		if parsed, err := strconv.Atoi(depth); err == nil {
			payload["depth"] = parsed
		}
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "knowledge.getGraph", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "knowledge.getGraph",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"nodes": []map[string]any{},
			"edges": []map[string]any{},
		},
		"bridge": map[string]any{
			"fallback":  "go-local-knowledge",
			"procedure": "knowledge.getGraph",
			"reason":    "upstream unavailable; knowledge graph data is unavailable",
		},
	})
}

func (s *Server) handleKnowledgeStats(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "knowledge.getStats", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "knowledge.getStats",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    s.localKnowledgeStats(),
		"bridge": map[string]any{
			"fallback":  "go-local-knowledge",
			"procedure": "knowledge.getStats",
			"reason":    "upstream unavailable; using local memory context count for knowledge stats",
		},
	})
}

func (s *Server) handleKnowledgeIngest(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "knowledge.ingest")
}

func (s *Server) handleKnowledgeResources(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "knowledge.getResources", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "knowledge.getResources",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    localKnowledgeResources(s.cfg.WorkspaceRoot),
		"bridge": map[string]any{
			"fallback":  "go-local-knowledge",
			"procedure": "knowledge.getResources",
			"reason":    "upstream unavailable; using local knowledge resources file",
		},
	})
}

func (s *Server) handleRAGIngestFile(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "rag.ingestFile")
}

func (s *Server) handleRAGIngestText(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "rag.ingestText")
}

func (s *Server) handleHighValueIngest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"success": false, "error": "method not allowed"})
		return
	}

	// Try upstream first
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "unifiedDirectory.highValueIngest", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "unifiedDirectory.highValueIngest",
			},
		})
		return
	}

	// Fallback: local Go high-value ingestor
	err = s.highValueIngestor.ProcessHighValueQueue(r.Context(), 5)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Local high-value link processing complete",
		"bridge": map[string]any{
			"fallback": "go-local-high-value-ingest",
		},
	})
}

func (s *Server) handleUnifiedDirectoryList(w http.ResponseWriter, r *http.Request) {
	payload := map[string]any{}
	limitValue := 50
	offsetValue := 0
	searchValue := ""
	sourceValue := "all"
	showDuplicatesValue := false
	duplicatesOnlyValue := false
	researchStatusValue := ""
	if limit := strings.TrimSpace(r.URL.Query().Get("limit")); limit != "" {
		if parsed, err := strconv.Atoi(limit); err == nil {
			limitValue = parsed
			payload["limit"] = parsed
		}
	}
	if offset := strings.TrimSpace(r.URL.Query().Get("offset")); offset != "" {
		if parsed, err := strconv.Atoi(offset); err == nil {
			offsetValue = parsed
			payload["offset"] = parsed
		}
	}
	if search := strings.TrimSpace(r.URL.Query().Get("search")); search != "" {
		searchValue = search
		payload["search"] = search
	}
	if source := strings.TrimSpace(r.URL.Query().Get("source")); source != "" {
		sourceValue = source
		payload["source"] = source
	}
	if showDuplicates := strings.TrimSpace(r.URL.Query().Get("show_duplicates")); showDuplicates != "" {
		showDuplicatesValue = strings.EqualFold(showDuplicates, "true") || showDuplicates == "1"
		payload["show_duplicates"] = showDuplicatesValue
	}
	if duplicatesOnly := strings.TrimSpace(r.URL.Query().Get("duplicates_only")); duplicatesOnly != "" {
		duplicatesOnlyValue = strings.EqualFold(duplicatesOnly, "true") || duplicatesOnly == "1"
		payload["duplicates_only"] = duplicatesOnlyValue
	}
	if researchStatus := strings.TrimSpace(r.URL.Query().Get("research_status")); researchStatus != "" {
		researchStatusValue = researchStatus
		payload["research_status"] = researchStatus
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "unifiedDirectory.list", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "unifiedDirectory.list",
			},
		})
		return
	}

	listPayload, fallbackErr := s.localUnifiedDirectoryList(limitValue, offsetValue, searchValue, sourceValue, showDuplicatesValue, duplicatesOnlyValue, researchStatusValue)
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
		"data":    listPayload,
		"bridge": map[string]any{
			"fallback":  "go-local-unified-directory",
			"procedure": "unifiedDirectory.list",
			"reason":    "upstream unavailable; using local published catalog and links backlog data",
		},
	})
}

func (s *Server) handleUnifiedDirectoryStats(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "unifiedDirectory.stats", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "unifiedDirectory.stats",
			},
		})
		return
	}

	stats, fallbackErr := s.localUnifiedDirectoryStats()
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
		"data":    stats,
		"bridge": map[string]any{
			"fallback":  "go-local-unified-directory",
			"procedure": "unifiedDirectory.stats",
			"reason":    "upstream unavailable; using local published catalog and links backlog stats",
		},
	})
}

func (s *Server) handleToolChainAliases(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "toolChaining.listAliases", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "toolChaining.listAliases",
			},
		})
		return
	}

	aliases, fallbackErr := s.localToolAliases()
	if fallbackErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   fallbackErr.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    aliases,
		"bridge": map[string]any{
			"fallback":  "go-local-toolchain-db",
			"procedure": "toolChaining.listAliases",
			"reason":    "upstream unavailable; using local tool aliases from metamcp.db",
		},
	})
}

func (s *Server) handleToolChainCreateAlias(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "toolChaining.createAlias")
}

func (s *Server) handleToolChainRemoveAlias(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "toolChaining.removeAlias")
}

func (s *Server) handleToolChainResolveAlias(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	if name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing name query parameter"})
		return
	}
	payload := map[string]any{"name": name}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "toolChaining.resolveAlias", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "toolChaining.resolveAlias",
			},
		})
		return
	}

	alias, fallbackErr := s.localToolAlias(name)
	if fallbackErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   fallbackErr.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    alias,
		"bridge": map[string]any{
			"fallback":  "go-local-toolchain-db",
			"procedure": "toolChaining.resolveAlias",
			"reason":    "upstream unavailable; using local tool alias from metamcp.db",
		},
	})
}

func (s *Server) handleToolChainsList(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "toolChaining.listChains", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "toolChaining.listChains",
			},
		})
		return
	}

	chains, fallbackErr := s.localToolChains()
	if fallbackErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   fallbackErr.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    chains,
		"bridge": map[string]any{
			"fallback":  "go-local-toolchain-db",
			"procedure": "toolChaining.listChains",
			"reason":    "upstream unavailable; using local tool chains from metamcp.db",
		},
	})
}

func (s *Server) handleToolChainsGet(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing id query parameter"})
		return
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "toolChaining.getChain", map[string]any{"id": id}, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "toolChaining.getChain",
			},
		})
		return
	}

	chain, fallbackErr := s.localToolChain(id)
	if fallbackErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   fallbackErr.Error(),
		})
		return
	}
	if chain == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"success": false,
			"error":   "tool chain unavailable",
			"detail":  "upstream unavailable; tool chain was not found in local metamcp tool chains",
			"bridge": map[string]any{
				"fallback":  "go-local-toolchain-db",
				"procedure": "toolChaining.getChain",
				"reason":    "upstream unavailable; tool chain was not found in local metamcp tool chains",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    chain,
		"bridge": map[string]any{
			"fallback":  "go-local-toolchain-db",
			"procedure": "toolChaining.getChain",
			"reason":    "upstream unavailable; using local tool chain from metamcp.db",
		},
	})
}

func (s *Server) handleToolChainsCreate(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "toolChaining.createChain")
}

func (s *Server) handleToolChainsExecute(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "toolChaining.executeChain")
}

func (s *Server) handleToolChainsDelete(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "toolChaining.deleteChain")
}

func (s *Server) handleToolChainsLazyStates(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "toolChaining.lazyStates", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "toolChaining.lazyStates",
			},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    []map[string]any{},
		"bridge": map[string]any{
			"fallback":  "go-local-registry",
			"procedure": "toolChaining.lazyStates",
			"reason":    "upstream unavailable; using local empty lazy-tool state",
		},
	})
}

func (s *Server) handleToolChainsRegisterLazy(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "toolChaining.registerLazy")
}

func (s *Server) handleToolChainsMarkLoaded(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "toolChaining.markLoaded")
}

func (s *Server) handleBrowserControlsScrape(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "browserControls.scrape")
}

func (s *Server) handleBrowserControlsPushHistory(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "browserControls.pushHistory")
}

func (s *Server) handleBrowserControlsQueryHistory(w http.ResponseWriter, r *http.Request) {
	payload := map[string]any{}
	queryValue := ""
	limitValue := 50
	sinceValue := int64(0)
	domainValue := ""
	if query := strings.TrimSpace(r.URL.Query().Get("query")); query != "" {
		queryValue = query
		payload["query"] = query
	}
	if limit := strings.TrimSpace(r.URL.Query().Get("limit")); limit != "" {
		if parsed, err := strconv.Atoi(limit); err == nil {
			limitValue = parsed
			payload["limit"] = parsed
		}
	}
	if since := strings.TrimSpace(r.URL.Query().Get("since")); since != "" {
		if parsed, err := strconv.ParseInt(since, 10, 64); err == nil {
			sinceValue = parsed
			payload["since"] = parsed
		}
	}
	if domain := strings.TrimSpace(r.URL.Query().Get("domain")); domain != "" {
		domainValue = domain
		payload["domain"] = domain
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "browserControls.queryHistory", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "browserControls.queryHistory",
			},
		})
		return
	}

	history, fallbackErr := s.localBrowserHistoryQuery(queryValue, limitValue, sinceValue, domainValue)
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
		"data":    history,
		"bridge": map[string]any{
			"fallback":  "go-local-browser-data-db",
			"procedure": "browserControls.queryHistory",
			"reason":    "upstream unavailable; using local metamcp browser history",
		},
	})
}

func (s *Server) handleBrowserControlsPushLogs(w http.ResponseWriter, r *http.Request) {
	s.handleTRPCBridgeBodyCall(w, r, "browserControls.pushConsoleLogs")
}

func (s *Server) handleBrowserControlsQueryLogs(w http.ResponseWriter, r *http.Request) {
	payload := map[string]any{}
	levelValue := ""
	searchValue := ""
	limitValue := 100
	if level := strings.TrimSpace(r.URL.Query().Get("level")); level != "" {
		levelValue = level
		payload["level"] = level
	}
	if search := strings.TrimSpace(r.URL.Query().Get("search")); search != "" {
		searchValue = search
		payload["search"] = search
	}
	if limit := strings.TrimSpace(r.URL.Query().Get("limit")); limit != "" {
		if parsed, err := strconv.Atoi(limit); err == nil {
			limitValue = parsed
			payload["limit"] = parsed
		}
	}
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "browserControls.queryConsoleLogs", payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "browserControls.queryConsoleLogs",
			},
		})
		return
	}

	logs, fallbackErr := s.localBrowserConsoleLogsQuery(levelValue, searchValue, limitValue)
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
		"data":    logs,
		"bridge": map[string]any{
			"fallback":  "go-local-browser-data-db",
			"procedure": "browserControls.queryConsoleLogs",
			"reason":    "upstream unavailable; using local metamcp browser console logs",
		},
	})
}

func (s *Server) handleBrowserControlsStats(w http.ResponseWriter, r *http.Request) {
	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), "browserControls.stats", nil, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    "browserControls.stats",
			},
		})
		return
	}

	stats, fallbackErr := s.localBrowserControlsStats()
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
		"data":    stats,
		"bridge": map[string]any{
			"fallback":  "go-local-browser-data-db",
			"procedure": "browserControls.stats",
			"reason":    "upstream unavailable; using local metamcp browser data stats",
		},
	})
}

func (s *Server) handleSessionBridgeBodyCall(w http.ResponseWriter, r *http.Request, procedure string) {
	s.handleTRPCBridgeBodyCall(w, r, procedure)
}

func (s *Server) handleSessionBridgeCall(w http.ResponseWriter, r *http.Request, method string, procedure string, payload any) {
	s.handleTRPCBridgeCall(w, r, method, procedure, payload)
}

func (s *Server) handleTRPCBridgeBodyCall(w http.ResponseWriter, r *http.Request, procedure string) {
	var payload any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "invalid JSON body",
		})
		return
	}
	s.handleTRPCBridgeCall(w, r, http.MethodPost, procedure, payload)
}

func (s *Server) handleTRPCBridgeCall(w http.ResponseWriter, r *http.Request, method string, procedure string, payload any) {
	if r.Method != method {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{
			"success": false,
			"error":   "method not allowed",
		})
		return
	}

	result, err := interop.CallTRPCProcedure(r.Context(), s.cfg.MainLockPath(), procedure, payload)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"success": false,
			"error":   "failed to call upstream procedure " + procedure + ": " + err.Error(),
		})
		return
	}

	var data any
	if err := json.Unmarshal(result.Data, &data); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"success": false,
			"error":   "invalid upstream JSON payload",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    data,
		"bridge": map[string]any{
			"upstreamBase": result.BaseURL,
			"procedure":    procedure,
		},
	})
}

func (s *Server) handleCLITools(w http.ResponseWriter, r *http.Request) {
	tools, err := s.detector.DetectAll(r.Context())
	if err != nil {
		writeJSON(w, http.StatusGatewayTimeout, map[string]any{
			"success": false,
			"error":   "failed to detect CLI tools: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    tools,
	})
}

func (s *Server) handleHarnesses(w http.ResponseWriter, r *http.Request) {
	tools, err := s.detector.DetectAll(r.Context())
	if err != nil {
		writeJSON(w, http.StatusGatewayTimeout, map[string]any{
			"success": false,
			"error":   "failed to detect CLI harnesses: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    harnesses.List(s.cfg.WorkspaceRoot, tools),
	})
}

func (s *Server) handleCLISummary(w http.ResponseWriter, r *http.Request) {
	tools, err := s.detector.DetectAll(r.Context())
	if err != nil {
		writeJSON(w, http.StatusGatewayTimeout, map[string]any{
			"success": false,
			"error":   "failed to detect CLI tools: " + err.Error(),
		})
		return
	}

	summary := summarizeCLI(s.cfg.WorkspaceRoot, tools)

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    summary,
	})
}

func (s *Server) handleImportSources(w http.ResponseWriter, _ *http.Request) {
	candidates, err := s.scanImportSources()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   "failed to scan import sources: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    candidates,
	})
}

func (s *Server) handleImportRoots(w http.ResponseWriter, _ *http.Request) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = s.cfg.MainConfigDir
	}

	scanner := sessionimport.NewScanner(s.cfg.WorkspaceRoot, homeDir, 50)
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    scanner.Roots(),
	})
}

func (s *Server) handleImportValidate(w http.ResponseWriter, r *http.Request) {
	targetPath := strings.TrimSpace(r.URL.Query().Get("path"))
	if targetPath == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "missing path query parameter",
		})
		return
	}

	info, err := os.Stat(targetPath)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{
			"success": false,
			"error":   "failed to stat import path: " + err.Error(),
		})
		return
	}

	result := sessionimport.ValidateCandidate(sessionimport.Candidate{
		SourceTool:     detectImportSourceTool(targetPath),
		SourcePath:     targetPath,
		SessionFormat:  sessionimport.DetectFormatFromPath(targetPath),
		LastModifiedAt: info.ModTime().UTC().Format(time.RFC3339),
		EstimatedSize:  info.Size(),
	})
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    result,
	})
}

func (s *Server) handleImportCandidates(w http.ResponseWriter, _ *http.Request) {
	candidates, err := s.scanValidatedImportSources()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   "failed to scan validated import candidates: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    candidates,
	})
}

func (s *Server) handleImportManifest(w http.ResponseWriter, _ *http.Request) {
	candidates, err := s.scanValidatedImportSources()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   "failed to build import manifest: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    sessionimport.BuildManifest(candidates),
	})
}

func (s *Server) handleImportSummary(w http.ResponseWriter, _ *http.Request) {
	candidates, err := s.scanValidatedImportSources()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   "failed to summarize validated import sources: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    sessionimport.BuildSummary(candidates),
	})
}

func (s *Server) handleMemoryStatus(w http.ResponseWriter, _ *http.Request) {
	status, err := memorystore.ReadStatus(s.cfg.WorkspaceRoot)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   "failed to read memory status: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    status,
	})
}

func (s *Server) handleConfiguredServerMutation(w http.ResponseWriter, r *http.Request, procedure string, fallback func(map[string]any) (any, error)) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{
			"success": false,
			"error":   "method not allowed",
		})
		return
	}

	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"error":   "invalid JSON body",
		})
		return
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), procedure, payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    procedure,
			},
		})
		return
	}

	fallbackResult, fallbackErr := fallback(payload)
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
		"data":    fallbackResult,
		"bridge": map[string]any{
			"fallback":  "go-local-jsonc",
			"procedure": procedure,
			"reason":    "upstream unavailable; applying local JSONC metadata placeholder fallback",
		},
	})
}

func (s *Server) handleRuntimeStatus(w http.ResponseWriter, r *http.Request) {
	candidates, err := s.scanImportSources()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   "failed to scan import sources: " + err.Error(),
		})
		return
	}

	tools, err := s.detector.DetectAll(r.Context())
	if err != nil {
		writeJSON(w, http.StatusGatewayTimeout, map[string]any{
			"success": false,
			"error":   "failed to detect CLI tools: " + err.Error(),
		})
		return
	}
	cliSummary := summarizeCLI(s.cfg.WorkspaceRoot, tools)
	rootStatuses := s.importRoots()
	existingRoots := 0
	for _, root := range rootStatuses {
		if root.Exists {
			existingRoots++
		}
	}

	instructions := interop.ReadImportedInstructions(s.cfg.ImportedInstructionsPath())
	configStatus := config.Snapshot(s.cfg)
	providerStatuses := providers.Snapshot()
	providerSummary := providers.BuildSummary(providerStatuses)
	configuredProviders := 0
	authenticatedProviders := 0
	for _, provider := range providerStatuses {
		if provider.Configured {
			configuredProviders++
		}
		if provider.Authenticated {
			authenticatedProviders++
		}
	}

	memoryStatus, err := memorystore.ReadStatus(s.cfg.WorkspaceRoot)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   "failed to read memory status: " + err.Error(),
		})
		return
	}

	discoveredSessions, err := s.discoveredSessions()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   "failed to summarize discovered sessions: " + err.Error(),
		})
		return
	}
	validSessions := 0
	for _, session := range discoveredSessions {
		if session.Valid {
			validSessions++
		}
	}

	sessionSummary := summarizeSessions(discoveredSessions)
	supervisorBridgeAvailable := false
	supervisorBridgeBase := ""
	bridgeCtx, cancelBridge := context.WithTimeout(r.Context(), 2*time.Second)
	bridgeResult, bridgeErr := interop.CallTRPCProcedure(bridgeCtx, s.cfg.MainLockPath(), "session.catalog", nil)
	cancelBridge()
	if bridgeErr == nil {
		supervisorBridgeAvailable = true
		supervisorBridgeBase = bridgeResult.BaseURL
	}

	validatedCandidates, err := s.scanValidatedImportSources()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"success": false,
			"error":   "failed to validate import sources: " + err.Error(),
		})
		return
	}
	importSummary := sessionimport.BuildSummary(validatedCandidates)

	lockStatuses := interop.DiscoverControlPlanes(s.cfg.MainLockPath(), s.cfg.LockPath())
	runningLocks := 0
	for _, status := range lockStatuses {
		if status.Running {
			runningLocks++
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": RuntimeStatus{
			Service:   "borg-go",
			Version:   buildinfo.Version,
			BaseURL:   s.cfg.BaseURL(),
			UptimeSec: int(time.Since(s.startedAt).Seconds()),
			Locks:     lockStatuses,
			LockSummary: LockRuntimeSummary{
				VisibleCount: len(lockStatuses),
				RunningCount: runningLocks,
			},
			Config: ConfigRuntimeSummary{
				WorkspaceRootAvailable: configStatus.WorkspaceRoot.Exists,
				ConfigDirAvailable:     configStatus.ConfigDir.Exists,
				MainConfigDirAvailable: configStatus.MainConfigDir.Exists,
				RepoConfigAvailable:    configStatus.BorgConfigFile.Exists,
				MCPConfigAvailable:     configStatus.MCPConfigFile.Exists,
				BorgSubmoduleAvailable: configStatus.BorgSubmodule.Exists,
			},
			CLI: CLIRuntimeSummary{
				ToolCount:                   cliSummary.ToolCount,
				AvailableToolCount:          cliSummary.AvailableToolCount,
				HarnessCount:                cliSummary.HarnessCount,
				InstalledHarnessCount:       cliSummary.InstalledHarnessCount,
				SourceBackedHarnessCount:    cliSummary.SourceBackedHarnessCount,
				MetadataOnlyHarnessCount:    cliSummary.MetadataOnlyHarnessCount,
				OperatorDefinedHarnessCount: cliSummary.OperatorDefinedHarnessCount,
				SourceBackedToolCount:       cliSummary.SourceBackedToolCount,
				PrimaryHarness:              cliSummary.PrimaryHarness,
			},
			Providers: ProviderRuntimeSummary{
				ProviderCount:           providerSummary.ProviderCount,
				ConfiguredCount:         configuredProviders,
				AuthenticatedCount:      authenticatedProviders,
				ExecutableCount:         providerSummary.ExecutableCount,
				RoutingPreviewAvailable: true,
				ByAuthMethod:            toHTTPBuckets(providerSummary.ByAuthMethod),
				ByPreferredTask:         toHTTPBuckets(providerSummary.ByPreferredTask),
				Statuses:                providerStatuses,
			},
			Memory: MemoryRuntimeSummary{
				Available:                  memoryStatus.Exists,
				StorePath:                  memoryStatus.StorePath,
				TotalEntries:               memoryStatus.TotalEntries,
				SectionCount:               memoryStatus.SectionCount,
				DefaultSectionCount:        memoryStatus.DefaultSectionCount,
				PopulatedSectionCount:      memoryStatus.PopulatedSectionCount,
				PresentDefaultSectionCount: memoryStatus.PresentDefaultSectionCount,
				MissingSections:            memoryStatus.MissingSections,
				Sections:                   memoryStatus.Sections,
				LastUpdatedAt:              memoryStatus.LastUpdatedAt,
			},
			Sessions: SessionRuntimeSummary{
				DiscoveredCount:           len(discoveredSessions),
				ValidCount:                validSessions,
				SupervisorBridgeAvailable: supervisorBridgeAvailable,
				SupervisorBridgeBase:      supervisorBridgeBase,
				ByCLIType:                 sessionSummary.ByCLIType,
				ByFormat:                  sessionSummary.ByFormat,
				ByTask:                    sessionSummary.ByTask,
				ByModelHint:               sessionSummary.ByModelHint,
			},
			ImportedInstructions: ImportedInstructionsSummary{
				Path:       instructions.Path,
				Available:  instructions.Available,
				ModifiedAt: instructions.ModifiedAt,
				Size:       instructions.Size,
			},
			ImportRoots: ImportRootsSummary{
				Count:         len(rootStatuses),
				ExistingCount: existingRoots,
				Roots:         rootStatuses,
			},
			ImportSources: ImportSourcesSummary{
				Count:              len(candidates),
				ValidCount:         importSummary.ValidCount,
				InvalidCount:       importSummary.InvalidCount,
				TotalEstimatedSize: importSummary.TotalEstimatedSize,
				Candidates:         candidates,
				BySourceTool:       toImportBuckets(importSummary.BySourceTool),
				BySourceType:       toImportBuckets(importSummary.BySourceType),
				ByFormat:           toImportBuckets(importSummary.ByFormat),
				ByModelHint:        toImportBuckets(importSummary.ByModelHint),
				ByError:            toImportBuckets(importSummary.ByError),
			},
		},
	})
}

func summarizeCLI(workspaceRoot string, tools []controlplane.Tool) CLISummary {
	availableTools := make([]controlplane.Tool, 0, len(tools))
	for _, tool := range tools {
		if tool.Available {
			availableTools = append(availableTools, tool)
		}
	}

	harnessDefinitions := harnesses.List(workspaceRoot, tools)
	harnessSummary := harnesses.Summarize(harnessDefinitions)
	installedHarnesses := make([]harnesses.Definition, 0, len(harnessDefinitions))
	primaryHarness := ""
	for _, harness := range harnessDefinitions {
		if harness.Primary {
			primaryHarness = harness.ID
		}
		if harness.Installed {
			installedHarnesses = append(installedHarnesses, harness)
		}
	}

	return CLISummary{
		ToolCount:                   len(tools),
		AvailableToolCount:          len(availableTools),
		HarnessCount:                len(harnessDefinitions),
		InstalledHarnessCount:       len(installedHarnesses),
		SourceBackedHarnessCount:    harnessSummary.SourceBackedHarnessCount,
		MetadataOnlyHarnessCount:    harnessSummary.MetadataOnlyHarnessCount,
		OperatorDefinedHarnessCount: harnessSummary.OperatorDefinedHarnessCount,
		SourceBackedToolCount:       harnessSummary.SourceBackedToolCount,
		PrimaryHarness:              primaryHarness,
		AvailableTools:              availableTools,
		InstalledHarnesses:          installedHarnesses,
	}
}

func localGitModules(workspaceRoot string) []map[string]any {
	content, err := os.ReadFile(filepath.Join(workspaceRoot, ".gitmodules"))
	if err != nil {
		return []map[string]any{}
	}

	regex := regexp.MustCompile(`\[submodule "(.*?)"\]\s*path = (.*?)\s*url = (.*?)\s`)
	matches := regex.FindAllStringSubmatch(string(content), -1)
	modules := make([]map[string]any, 0, len(matches))
	date := time.Now().Format("2006-01-02")
	for _, match := range matches {
		if len(match) < 4 {
			continue
		}
		modules = append(modules, map[string]any{
			"name":       match[1],
			"path":       match[2],
			"url":        match[3],
			"status":     "unknown",
			"branch":     "main",
			"lastCommit": "HEAD",
			"date":       date,
			"active":     false,
		})
	}
	return modules
}

func localGitStatus(workspaceRoot string) map[string]any {
	branchCommand := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	branchCommand.Dir = workspaceRoot
	branchOut, branchErr := branchCommand.Output()
	if branchErr != nil {
		return map[string]any{
			"branch":   "unknown",
			"clean":    false,
			"modified": []string{},
			"staged":   []string{},
		}
	}

	statusCommand := exec.Command("git", "status", "--porcelain")
	statusCommand.Dir = workspaceRoot
	statusOut, statusErr := statusCommand.Output()
	if statusErr != nil {
		return map[string]any{
			"branch":   "unknown",
			"clean":    false,
			"modified": []string{},
			"staged":   []string{},
		}
	}

	modified := []string{}
	staged := []string{}
	lines := strings.Split(strings.TrimSpace(string(statusOut)), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" || len(line) < 4 {
			continue
		}
		code := line[:2]
		file := strings.TrimSpace(line[3:])
		if strings.Contains(code, "M") || strings.Contains(code, "?") {
			modified = append(modified, file)
		}
		if strings.Contains(code, "A") || (strings.Contains(code, "M") && code[0] != ' ') {
			staged = append(staged, file)
		}
	}

	return map[string]any{
		"branch":   strings.TrimSpace(string(branchOut)),
		"clean":    strings.TrimSpace(string(statusOut)) == "",
		"modified": modified,
		"staged":   staged,
	}
}

func localGitLog(workspaceRoot string, limit int) []map[string]any {
	if limit <= 0 {
		limit = 20
	}
	command := exec.Command("git", "log", "-n", strconv.Itoa(limit), `--pretty=format:%H|%an|%aI|%s`)
	command.Dir = workspaceRoot
	output, err := command.Output()
	if err != nil {
		return []map[string]any{}
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	results := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 4)
		if len(parts) != 4 {
			continue
		}
		results = append(results, map[string]any{
			"hash":    parts[0],
			"author":  parts[1],
			"date":    parts[2],
			"message": parts[3],
		})
	}
	return results
}

func localSubmoduleList(workspaceRoot string) []map[string]any {
	modules := localGitModules(workspaceRoot)
	results := make([]map[string]any, 0, len(modules))
	for _, module := range modules {
		modulePath, _ := module["path"].(string)
		name, _ := module["name"].(string)
		url, _ := module["url"].(string)
		caps, startCommand := localSubmoduleCapabilitiesValues(workspaceRoot, modulePath)
		fullPath := filepath.Join(workspaceRoot, filepath.FromSlash(modulePath))
		results = append(results, map[string]any{
			"name":         coalesceSubmoduleName(name, modulePath),
			"path":         modulePath,
			"commit":       "HEAD",
			"branch":       "HEAD",
			"status":       submoduleStatusFromPath(fullPath),
			"url":          emptyStringToNilAny(url),
			"capabilities": caps,
			"isInstalled":  fileExists(filepath.Join(fullPath, "node_modules")) || fileExists(filepath.Join(fullPath, ".venv")),
			"isBuilt":      submoduleBuildExists(fullPath),
			"startCommand": emptyStringToNilAny(startCommand),
		})
	}
	return results
}

func localSubmoduleCapabilities(workspaceRoot string, submodulePath string) map[string]any {
	caps, startCommand := localSubmoduleCapabilitiesValues(workspaceRoot, submodulePath)
	result := map[string]any{
		"caps": caps,
	}
	if strings.TrimSpace(startCommand) != "" {
		result["startCommand"] = startCommand
	}
	return result
}

func (s *Server) localKnowledgeStats() map[string]any {
	contexts, err := s.localMemoryContexts()
	if err != nil {
		return map[string]any{"count": 0}
	}
	return map[string]any{"count": len(contexts)}
}

func localKnowledgeResources(workspaceRoot string) any {
	resourcePath := filepath.Join(workspaceRoot, "knowledge", "resources.json")
	raw, err := os.ReadFile(resourcePath)
	if err != nil {
		return map[string]any{
			"lastUpdated": "Never",
			"categories":  []any{},
		}
	}

	var parsed any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return map[string]any{
			"lastUpdated": "Never",
			"categories":  []any{},
		}
	}
	return parsed
}

func (s *Server) localPlanSandboxDir() string {
	return filepath.Join(s.cfg.WorkspaceRoot, ".borg", "sandbox")
}

func (s *Server) localPlanAllDiffs() []map[string]any {
	sandboxDir := s.localPlanSandboxDir()
	entries, err := os.ReadDir(sandboxDir)
	if err != nil {
		return []map[string]any{}
	}

	results := make([]map[string]any, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".json") || strings.EqualFold(entry.Name(), "checkpoints.json") {
			continue
		}

		raw, err := os.ReadFile(filepath.Join(sandboxDir, entry.Name()))
		if err != nil {
			continue
		}

		var parsed map[string]any
		if err := json.Unmarshal(raw, &parsed); err != nil {
			continue
		}
		results = append(results, parsed)
	}

	slices.SortStableFunc(results, func(a, b map[string]any) int {
		return strings.Compare(fmt.Sprint(a["id"]), fmt.Sprint(b["id"]))
	})
	return results
}

func (s *Server) localPlanDiffs() []map[string]any {
	all := s.localPlanAllDiffs()
	results := make([]map[string]any, 0, len(all))
	for _, diff := range all {
		if status, _ := diff["status"].(string); status == "pending" {
			results = append(results, diff)
		}
	}
	return results
}

func (s *Server) localPlanCheckpoints() []map[string]any {
	raw, err := os.ReadFile(filepath.Join(s.localPlanSandboxDir(), "checkpoints.json"))
	if err != nil {
		return []map[string]any{}
	}

	var parsed []map[string]any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return []map[string]any{}
	}
	return parsed
}

func (s *Server) localPlanSummary() string {
	diffs := s.localPlanAllDiffs()
	checkpoints := s.localPlanCheckpoints()
	pending := 0
	approved := 0
	applied := 0
	rejected := 0
	for _, diff := range diffs {
		switch fmt.Sprint(diff["status"]) {
		case "pending":
			pending++
		case "approved":
			approved++
		case "applied":
			applied++
		case "rejected":
			rejected++
		}
	}

	return strings.Join([]string{
		"Diff Sandbox Summary:",
		fmt.Sprintf("  Pending: %d", pending),
		fmt.Sprintf("  Approved: %d", approved),
		fmt.Sprintf("  Applied: %d", applied),
		fmt.Sprintf("  Rejected: %d", rejected),
		fmt.Sprintf("  Checkpoints: %d", len(checkpoints)),
	}, "\n")
}

func (s *Server) localMetaMCPDBPath() string {
	return filepath.Join(s.cfg.WorkspaceRoot, "metamcp.db")
}

func (s *Server) localPolicy(uuid string) (any, error) {
	db, err := sql.Open("sqlite", s.localMetaMCPDBPath())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var (
		policyUUID   string
		name         string
		description  sql.NullString
		rulesRaw     string
		createdAtRaw int64
		updatedAtRaw int64
	)

	row := db.QueryRow(`
		SELECT uuid, name, description, rules, created_at, updated_at
		FROM policies
		WHERE uuid = ?
		LIMIT 1
	`, uuid)
	if err := row.Scan(&policyUUID, &name, &description, &rulesRaw, &createdAtRaw, &updatedAtRaw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	var rules any
	if err := json.Unmarshal([]byte(rulesRaw), &rules); err != nil {
		rules = map[string]any{}
	}

	return map[string]any{
		"uuid":        policyUUID,
		"name":        name,
		"description": nullStringToAny(description),
		"rules":       rules,
		"createdAt":   unixTimestampToRFC3339(createdAtRaw),
		"updatedAt":   unixTimestampToRFC3339(updatedAtRaw),
	}, nil
}

func (s *Server) localPolicies() ([]map[string]any, error) {
	db, err := sql.Open("sqlite", s.localMetaMCPDBPath())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT uuid, name, description, rules, created_at, updated_at
		FROM policies
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]map[string]any, 0)
	for rows.Next() {
		var (
			policyUUID   string
			name         string
			description  sql.NullString
			rulesRaw     string
			createdAtRaw int64
			updatedAtRaw int64
		)
		if err := rows.Scan(&policyUUID, &name, &description, &rulesRaw, &createdAtRaw, &updatedAtRaw); err != nil {
			return nil, err
		}

		var rules any
		if err := json.Unmarshal([]byte(rulesRaw), &rules); err != nil {
			rules = map[string]any{}
		}

		results = append(results, map[string]any{
			"uuid":        policyUUID,
			"name":        name,
			"description": nullStringToAny(description),
			"rules":       rules,
			"createdAt":   unixTimestampToRFC3339(createdAtRaw),
			"updatedAt":   unixTimestampToRFC3339(updatedAtRaw),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (s *Server) localSecrets() ([]map[string]any, error) {
	db, err := sql.Open("sqlite", s.localMetaMCPDBPath())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT key, created_at, updated_at
		FROM workspace_secrets
		ORDER BY updated_at DESC, created_at DESC, key ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]map[string]any, 0)
	for rows.Next() {
		var (
			key          string
			createdAtRaw int64
			updatedAtRaw int64
		)
		if err := rows.Scan(&key, &createdAtRaw, &updatedAtRaw); err != nil {
			return nil, err
		}
		results = append(results, map[string]any{
			"key":        key,
			"created_at": unixTimestampToRFC3339(createdAtRaw),
			"updated_at": unixTimestampToRFC3339(updatedAtRaw),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func (s *Server) localAPIKeys() ([]map[string]any, error) {
	db, err := sql.Open("sqlite", s.localMetaMCPDBPath())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT uuid, name, key, created_at, is_active, user_id
		FROM api_keys
		WHERE user_id IS NULL
		ORDER BY created_at DESC
	`)
	if err != nil {
		if strings.Contains(err.Error(), "no such table: api_keys") {
			return []map[string]any{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	results := make([]map[string]any, 0)
	for rows.Next() {
		var (
			keyUUID      string
			name         string
			keyValue     string
			createdAtRaw int64
			isActive     bool
			userID       sql.NullString
		)
		if err := rows.Scan(&keyUUID, &name, &keyValue, &createdAtRaw, &isActive, &userID); err != nil {
			return nil, err
		}

		results = append(results, map[string]any{
			"uuid":       keyUUID,
			"name":       name,
			"key":        keyValue,
			"created_at": unixTimestampToRFC3339(createdAtRaw),
			"is_active":  isActive,
			"user_id":    nullStringToAny(userID),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (s *Server) localAPIKey(uuid string) (any, error) {
	db, err := sql.Open("sqlite", s.localMetaMCPDBPath())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var (
		keyUUID      string
		name         string
		keyValue     string
		createdAtRaw int64
		isActive     bool
		userID       sql.NullString
	)

	row := db.QueryRow(`
		SELECT uuid, name, key, created_at, is_active, user_id
		FROM api_keys
		WHERE uuid = ? AND user_id IS NULL
		LIMIT 1
	`, uuid)
	if err := row.Scan(&keyUUID, &name, &keyValue, &createdAtRaw, &isActive, &userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return map[string]any{
		"uuid":       keyUUID,
		"name":       name,
		"key":        keyValue,
		"created_at": unixTimestampToRFC3339(createdAtRaw),
		"is_active":  isActive,
		"user_id":    nullStringToAny(userID),
	}, nil
}

func (s *Server) localLinksBacklogItem(uuid string) (any, error) {
	db, err := sql.Open("sqlite", s.localMetaMCPDBPath())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	row := db.QueryRow(`
		SELECT uuid, url, normalized_url, title, description, tags, source, is_duplicate, duplicate_of,
		       research_status, http_status, page_title, page_description, favicon_url, researched_at,
		       cluster_id, bobbybookmarks_bookmark_id, import_session_id, raw_payload, synced_at, created_at, updated_at
		FROM links_backlog
		WHERE uuid = ?
		LIMIT 1
	`, uuid)
	item, err := scanLinksBacklogItem(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return item, nil
}

func (s *Server) localLinksBacklogStats() (any, error) {
	db, err := sql.Open("sqlite", s.localMetaMCPDBPath())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var (
		total      int64
		duplicates int64
		pending    int64
		researched int64
		failed     int64
		sources    int64
	)

	row := db.QueryRow(`
		SELECT
			count(*) AS total,
			coalesce(sum(case when is_duplicate = 1 then 1 else 0 end), 0) AS duplicates,
			coalesce(sum(case when research_status = 'pending' then 1 else 0 end), 0) AS pending,
			coalesce(sum(case when research_status = 'done' then 1 else 0 end), 0) AS researched,
			coalesce(sum(case when research_status = 'failed' then 1 else 0 end), 0) AS failed,
			count(distinct source) AS sources
		FROM links_backlog
	`)
	if err := row.Scan(&total, &duplicates, &pending, &researched, &failed, &sources); err != nil {
		return nil, err
	}

	unique := total - duplicates
	if unique < 0 {
		unique = 0
	}

	return map[string]any{
		"total":      total,
		"unique":     unique,
		"duplicates": duplicates,
		"pending":    pending,
		"researched": researched,
		"failed":     failed,
		"sources":    sources,
	}, nil
}

func (s *Server) localLinksBacklogList(limit, offset int, search, source, researchStatus, clusterID string, showDuplicates bool) (any, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	db, err := sql.Open("sqlite", s.localMetaMCPDBPath())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	items, err := listLinksBacklogRows(db, limit, offset, search, source, researchStatus, clusterID, showDuplicates)
	if err != nil {
		return nil, err
	}
	total, err := countLinksBacklogRows(db, search, source, researchStatus, clusterID, showDuplicates)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"items": items,
		"total": total,
	}, nil
}

func (s *Server) localOAuthClient(clientID string) (any, error) {
	db, err := sql.Open("sqlite", s.localMetaMCPDBPath())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var (
		foundClientID           string
		clientSecret            sql.NullString
		clientName              string
		redirectURIsRaw         string
		grantTypesRaw           string
		responseTypesRaw        string
		tokenEndpointAuthMethod string
		scope                   sql.NullString
		clientURI               sql.NullString
		logoURI                 sql.NullString
		contactsRaw             sql.NullString
		tosURI                  sql.NullString
		policyURI               sql.NullString
		softwareID              sql.NullString
		softwareVersion         sql.NullString
		createdAtRaw            int64
		updatedAtRaw            int64
	)

	row := db.QueryRow(`
		SELECT client_id, client_secret, client_name, redirect_uris, grant_types, response_types,
		       token_endpoint_auth_method, scope, client_uri, logo_uri, contacts, tos_uri,
		       policy_uri, software_id, software_version, created_at, updated_at
		FROM oauth_clients
		WHERE client_id = ?
		LIMIT 1
	`, clientID)
	if err := row.Scan(
		&foundClientID, &clientSecret, &clientName, &redirectURIsRaw, &grantTypesRaw, &responseTypesRaw,
		&tokenEndpointAuthMethod, &scope, &clientURI, &logoURI, &contactsRaw, &tosURI,
		&policyURI, &softwareID, &softwareVersion, &createdAtRaw, &updatedAtRaw,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	redirectURIs := jsonArrayOrEmpty(redirectURIsRaw)
	grantTypes := jsonArrayOrEmpty(grantTypesRaw)
	responseTypes := jsonArrayOrEmpty(responseTypesRaw)

	var contacts any
	if contactsRaw.Valid {
		contacts = jsonArrayOrEmpty(contactsRaw.String)
	}

	return map[string]any{
		"client_id":                  foundClientID,
		"client_secret":              nullStringToAny(clientSecret),
		"client_name":                clientName,
		"redirect_uris":              redirectURIs,
		"grant_types":                grantTypes,
		"response_types":             responseTypes,
		"token_endpoint_auth_method": tokenEndpointAuthMethod,
		"scope":                      nullStringToAny(scope),
		"client_uri":                 nullStringToAny(clientURI),
		"logo_uri":                   nullStringToAny(logoURI),
		"contacts":                   contacts,
		"tos_uri":                    nullStringToAny(tosURI),
		"policy_uri":                 nullStringToAny(policyURI),
		"software_id":                nullStringToAny(softwareID),
		"software_version":           nullStringToAny(softwareVersion),
		"created_at":                 unixTimestampToRFC3339(createdAtRaw),
		"updated_at":                 unixTimestampToRFC3339(updatedAtRaw),
	}, nil
}

func (s *Server) localOAuthSessionByServer(serverUUID string) (any, error) {
	db, err := sql.Open("sqlite", s.localMetaMCPDBPath())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var (
		uuid              string
		mcpServerUUID     string
		clientInformation string
		tokensRaw         sql.NullString
		codeVerifier      sql.NullString
		createdAtRaw      int64
		updatedAtRaw      int64
	)

	row := db.QueryRow(`
		SELECT uuid, mcp_server_uuid, client_information, tokens, code_verifier, created_at, updated_at
		FROM oauth_sessions
		WHERE mcp_server_uuid = ?
		LIMIT 1
	`, serverUUID)
	if err := row.Scan(
		&uuid, &mcpServerUUID, &clientInformation, &tokensRaw, &codeVerifier, &createdAtRaw, &updatedAtRaw,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	clientInfo := jsonObjectOrEmpty(clientInformation)

	var tokens any
	if tokensRaw.Valid {
		tokens = jsonObjectOrNil(tokensRaw.String)
	}

	return map[string]any{
		"uuid":               uuid,
		"mcp_server_uuid":    mcpServerUUID,
		"client_information": clientInfo,
		"tokens":             tokens,
		"code_verifier":      nullStringToAny(codeVerifier),
		"created_at":         unixTimestampToRFC3339(createdAtRaw),
		"updated_at":         unixTimestampToRFC3339(updatedAtRaw),
	}, nil
}

func (s *Server) localCatalogGet(uuid string) (any, error) {
	db, err := sql.Open("sqlite", s.localMetaMCPDBPath())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	server, err := localPublishedCatalogServer(db, uuid)
	if err != nil {
		return nil, err
	}
	if server == nil {
		return nil, nil
	}
	latestRun, err := localPublishedCatalogLatestRun(db, uuid)
	if err != nil {
		return nil, err
	}
	activeRecipe, err := localPublishedCatalogActiveRecipe(db, uuid)
	if err != nil {
		return nil, err
	}
	sources, err := localPublishedCatalogSources(db, uuid)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"server":       server,
		"latestRun":    latestRun,
		"activeRecipe": activeRecipe,
		"sources":      sources,
	}, nil
}

type marketplaceLegacyRegistryItem struct {
	Name string   `json:"name"`
	URL  string   `json:"url"`
	Tags []string `json:"tags"`
}

type marketplaceLegacyRegistryData struct {
	Directories []marketplaceLegacyRegistryItem `json:"directories"`
	Skills      []marketplaceLegacyRegistryItem `json:"skills"`
}

type marketplaceMCPRegistryDocument struct {
	Servers []marketplaceMCPRegistryItem `json:"servers"`
}

type marketplaceMCPRegistryItem struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Package     string   `json:"package"`
	Type        string   `json:"type"`
	Env         []string `json:"env"`
}

func (s *Server) localMarketplaceList(filter string) ([]map[string]any, error) {
	entries := []map[string]any{}

	legacyEntries, err := s.localMarketplaceLegacyEntries(filter)
	if err != nil {
		return nil, err
	}
	entries = append(entries, legacyEntries...)

	mcpEntries, err := s.localMarketplaceMCPRegistryEntries(filter)
	if err != nil {
		return nil, err
	}
	entries = append(entries, mcpEntries...)

	return entries, nil
}

func (s *Server) localMarketplaceLegacyEntries(filter string) ([]map[string]any, error) {
	registryPath := filepath.Join(s.cfg.WorkspaceRoot, "packages", "core", "data", "skills_registry.json")
	raw, err := os.ReadFile(registryPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []map[string]any{}, nil
		}
		return nil, err
	}

	var document marketplaceLegacyRegistryData
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, err
	}

	items := append([]marketplaceLegacyRegistryItem{}, document.Directories...)
	items = append(items, document.Skills...)
	filterLower := strings.ToLower(strings.TrimSpace(filter))
	entries := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.Name) == "" {
			continue
		}
		if filterLower != "" {
			matches := strings.Contains(strings.ToLower(item.Name), filterLower)
			if !matches {
				for _, tag := range item.Tags {
					if strings.Contains(strings.ToLower(tag), filterLower) {
						matches = true
						break
					}
				}
			}
			if !matches {
				continue
			}
		}
		entries = append(entries, map[string]any{
			"id":          item.Name,
			"name":        item.Name,
			"description": "Official Skill",
			"author":      "borg Ecosystem",
			"type":        "skill",
			"source":      "official",
			"url":         emptyStringToNilAny(item.URL),
			"verified":    true,
			"peerCount":   1,
			"installed":   s.localMarketplaceSkillInstalled(item.Name),
			"tags":        append([]string(nil), item.Tags...),
		})
	}
	return entries, nil
}

func (s *Server) localMarketplaceMCPRegistryEntries(filter string) ([]map[string]any, error) {
	registryPath := filepath.Join(s.cfg.WorkspaceRoot, "packages", "mcp-registry", "src", "registry.json")
	raw, err := os.ReadFile(registryPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []map[string]any{}, nil
		}
		return nil, err
	}

	var document marketplaceMCPRegistryDocument
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, err
	}

	filterLower := strings.ToLower(strings.TrimSpace(filter))
	entries := make([]map[string]any, 0, len(document.Servers))
	for _, server := range document.Servers {
		if strings.TrimSpace(server.Name) == "" || strings.TrimSpace(server.Package) == "" {
			continue
		}
		if filterLower != "" &&
			!strings.Contains(strings.ToLower(server.Name), filterLower) &&
			!strings.Contains(strings.ToLower(server.Description), filterLower) {
			continue
		}
		serverType := strings.TrimSpace(server.Type)
		if serverType == "" {
			serverType = "stdio"
		}
		entries = append(entries, map[string]any{
			"id":          server.Package,
			"name":        server.Name,
			"description": server.Description,
			"author":      "MCP Registry",
			"type":        "tool",
			"source":      "official",
			"url":         "https://www.npmjs.com/package/" + server.Package,
			"verified":    true,
			"peerCount":   1,
			"installed":   s.localMarketplaceMCPInstalled(server.Package),
			"tags":        []string{"mcp", serverType},
		})
	}
	return entries, nil
}

func (s *Server) localMarketplaceSkillInstalled(id string) bool {
	if strings.TrimSpace(id) == "" {
		return false
	}
	_, err := os.Stat(filepath.Join(s.cfg.WorkspaceRoot, ".borg", "skills", id))
	return err == nil
}

func (s *Server) localMarketplaceMCPInstalled(id string) bool {
	if strings.TrimSpace(id) == "" {
		return false
	}
	document, err := s.localMarketplaceMCPConfig()
	if err != nil {
		return false
	}
	servers, _ := document["mcpServers"].(map[string]any)
	for _, raw := range servers {
		serialized := prettyJSON(raw)
		if strings.Contains(serialized, id) {
			return true
		}
	}
	return false
}

func (s *Server) localMarketplaceMCPConfig() (map[string]any, error) {
	paths := []string{
		filepath.Join(s.cfg.MainConfigDir, "mcp.jsonc"),
		filepath.Join(s.cfg.MainConfigDir, "mcp.json"),
	}
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		content := string(raw)
		if strings.HasSuffix(strings.ToLower(path), ".jsonc") {
			content = stripJSONCLineComments(content)
		}
		var document map[string]any
		if err := json.Unmarshal([]byte(content), &document); err != nil {
			return nil, err
		}
		if _, ok := document["mcpServers"]; !ok {
			document["mcpServers"] = map[string]any{}
		}
		return document, nil
	}
	return map[string]any{"mcpServers": map[string]any{}}, nil
}

func (s *Server) localCatalogRuns(serverUUID string, limit int) ([]map[string]any, error) {
	if limit <= 0 {
		limit = 10
	}

	db, err := sql.Open("sqlite", s.localMetaMCPDBPath())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT uuid, server_uuid, run_mode, started_at, finished_at, outcome, failure_class, tool_count,
		       findings_summary, performed_by, created_at
		FROM published_mcp_validation_runs
		WHERE server_uuid = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, serverUUID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	runs := []map[string]any{}
	for rows.Next() {
		run, scanErr := scanPublishedCatalogRun(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		runs = append(runs, run)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return runs, nil
}

func (s *Server) localCatalogStats() (any, error) {
	db, err := sql.Open("sqlite", s.localMetaMCPDBPath())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	total, err := countPublishedCatalogServers(db, "")
	if err != nil {
		return nil, err
	}
	validated, err := countPublishedCatalogServers(db, "validated")
	if err != nil {
		return nil, err
	}
	broken, err := countPublishedCatalogServers(db, "broken")
	if err != nil {
		return nil, err
	}
	recentlyUpdated, err := countPublishedCatalogRecentlyUpdated(db, 24)
	if err != nil {
		return nil, err
	}

	statuses := []string{"discovered", "normalized", "probeable", "validated", "certified", "broken", "archived"}
	statusCounts := make([]map[string]any, 0, len(statuses))
	for _, status := range statuses {
		count, countErr := countPublishedCatalogServers(db, status)
		if countErr != nil {
			return nil, countErr
		}
		statusCounts = append(statusCounts, map[string]any{
			"status": status,
			"count":  count,
		})
	}

	return map[string]any{
		"total":           total,
		"validated":       validated,
		"broken":          broken,
		"recentlyUpdated": recentlyUpdated,
		"statusCounts":    statusCounts,
	}, nil
}

func (s *Server) localCatalogLinkedServers(publishedServerUUID string) ([]map[string]any, error) {
	db, err := sql.Open("sqlite", s.localMetaMCPDBPath())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT uuid, name, description, type, command, args, env, url, error_status, created_at,
		       bearer_token, headers, always_on, user_id, source_published_server_uuid
		FROM mcp_servers
		WHERE source_published_server_uuid = ?
	`, publishedServerUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	servers := []map[string]any{}
	for rows.Next() {
		server, scanErr := scanLocalMCPServer(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		servers = append(servers, server)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return servers, nil
}

func (s *Server) localCatalogList(limit, offset int, search, status, transport, installMethod string) (any, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	db, err := sql.Open("sqlite", s.localMetaMCPDBPath())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	items, err := listPublishedCatalogServers(db, limit, offset, search, status, transport, installMethod)
	if err != nil {
		return nil, err
	}
	total, err := countPublishedCatalogServersFiltered(db, search, status, transport, installMethod)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"servers": items,
		"total":   total,
	}, nil
}

func listLinksBacklogRows(db *sql.DB, limit, offset int, search, source, researchStatus, clusterID string, showDuplicates bool) ([]map[string]any, error) {
	args := []any{}
	where := buildLinksBacklogWhere(&args, search, source, researchStatus, clusterID, showDuplicates)
	query := `
		SELECT uuid, url, normalized_url, title, description, tags, source, is_duplicate, duplicate_of,
		       research_status, http_status, page_title, page_description, favicon_url, researched_at,
		       cluster_id, bobbybookmarks_bookmark_id, import_session_id, raw_payload, synced_at, created_at, updated_at
		FROM links_backlog`
	if where != "" {
		query += " WHERE " + where
	}
	query += " ORDER BY updated_at DESC, created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []map[string]any{}
	for rows.Next() {
		item, scanErr := scanLinksBacklogItem(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func countLinksBacklogRows(db *sql.DB, search, source, researchStatus, clusterID string, showDuplicates bool) (int64, error) {
	args := []any{}
	where := buildLinksBacklogWhere(&args, search, source, researchStatus, clusterID, showDuplicates)
	query := `SELECT count(*) FROM links_backlog`
	if where != "" {
		query += " WHERE " + where
	}

	row := db.QueryRow(query, args...)
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func buildLinksBacklogWhere(args *[]any, search, source, researchStatus, clusterID string, showDuplicates bool) string {
	conditions := []string{}
	if !showDuplicates {
		conditions = append(conditions, "is_duplicate = 0")
	}
	if strings.TrimSpace(source) != "" {
		conditions = append(conditions, "source = ?")
		*args = append(*args, source)
	}
	if strings.TrimSpace(researchStatus) != "" {
		conditions = append(conditions, "research_status = ?")
		*args = append(*args, researchStatus)
	}
	if strings.TrimSpace(clusterID) != "" {
		conditions = append(conditions, "cluster_id = ?")
		*args = append(*args, clusterID)
	}
	if trimmed := strings.TrimSpace(search); trimmed != "" {
		term := "%" + trimmed + "%"
		conditions = append(conditions, "(url LIKE ? OR normalized_url LIKE ? OR title LIKE ? OR description LIKE ?)")
		*args = append(*args, term, term, term, term)
	}
	return strings.Join(conditions, " AND ")
}

type linksBacklogScanner interface {
	Scan(dest ...any) error
}

func scanLinksBacklogItem(scanner linksBacklogScanner) (map[string]any, error) {
	var (
		itemUUID                 string
		urlValue                 string
		normalizedURL            string
		title                    sql.NullString
		description              sql.NullString
		tagsRaw                  string
		source                   string
		isDuplicate              bool
		duplicateOf              sql.NullString
		researchStatus           string
		httpStatus               sql.NullInt64
		pageTitle                sql.NullString
		pageDescription          sql.NullString
		faviconURL               sql.NullString
		researchedAt             sql.NullInt64
		clusterID                sql.NullString
		bobbyBookmarksBookmarkID sql.NullInt64
		importSessionID          sql.NullInt64
		rawPayloadText           sql.NullString
		syncedAt                 sql.NullInt64
		createdAtRaw             int64
		updatedAtRaw             int64
	)

	if err := scanner.Scan(
		&itemUUID, &urlValue, &normalizedURL, &title, &description, &tagsRaw, &source, &isDuplicate, &duplicateOf,
		&researchStatus, &httpStatus, &pageTitle, &pageDescription, &faviconURL, &researchedAt,
		&clusterID, &bobbyBookmarksBookmarkID, &importSessionID, &rawPayloadText, &syncedAt, &createdAtRaw, &updatedAtRaw,
	); err != nil {
		return nil, err
	}

	var tags any
	if err := json.Unmarshal([]byte(tagsRaw), &tags); err != nil {
		tags = []any{}
	}

	var rawPayload any
	if rawPayloadText.Valid {
		if err := json.Unmarshal([]byte(rawPayloadText.String), &rawPayload); err != nil {
			rawPayload = nil
		}
	}

	return map[string]any{
		"uuid":                       itemUUID,
		"url":                        urlValue,
		"normalized_url":             normalizedURL,
		"title":                      nullStringToAny(title),
		"description":                nullStringToAny(description),
		"tags":                       tags,
		"source":                     source,
		"is_duplicate":               isDuplicate,
		"duplicate_of":               nullStringToAny(duplicateOf),
		"research_status":            researchStatus,
		"http_status":                nullInt64ToAny(httpStatus),
		"page_title":                 nullStringToAny(pageTitle),
		"page_description":           nullStringToAny(pageDescription),
		"favicon_url":                nullStringToAny(faviconURL),
		"researched_at":              nullTimestampToAny(researchedAt),
		"cluster_id":                 nullStringToAny(clusterID),
		"bobbybookmarks_bookmark_id": nullInt64ToAny(bobbyBookmarksBookmarkID),
		"import_session_id":          nullInt64ToAny(importSessionID),
		"raw_payload":                rawPayload,
		"synced_at":                  nullTimestampToAny(syncedAt),
		"created_at":                 unixTimestampToRFC3339(createdAtRaw),
		"updated_at":                 unixTimestampToRFC3339(updatedAtRaw),
	}, nil
}

func (s *Server) localBrowserHistoryQuery(query string, limit int, since int64, domain string) (any, error) {
	if limit <= 0 {
		limit = 50
	}

	db, err := sql.Open("sqlite", s.localMetaMCPDBPath())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT id, url, title, domain, visited_at, visit_count
		FROM browser_history
		ORDER BY visited_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []map[string]any{}
	queryLower := strings.ToLower(strings.TrimSpace(query))
	for rows.Next() {
		var (
			id         string
			urlValue   string
			title      string
			domainName string
			visitedAt  int64
			visitCount int64
		)
		if err := rows.Scan(&id, &urlValue, &title, &domainName, &visitedAt, &visitCount); err != nil {
			return nil, err
		}
		if queryLower != "" && !strings.Contains(strings.ToLower(title), queryLower) && !strings.Contains(strings.ToLower(urlValue), queryLower) {
			continue
		}
		if strings.TrimSpace(domain) != "" && domainName != domain {
			continue
		}
		if since > 0 && visitedAt < since {
			continue
		}
		entries = append(entries, map[string]any{
			"id":         id,
			"url":        urlValue,
			"title":      title,
			"domain":     domainName,
			"visitedAt":  visitedAt,
			"visitCount": visitCount,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	total := len(entries)
	if len(entries) > limit {
		entries = entries[:limit]
	}

	return map[string]any{
		"entries": entries,
		"total":   total,
	}, nil
}

func (s *Server) localBrowserConsoleLogsQuery(level, search string, limit int) (any, error) {
	if limit <= 0 {
		limit = 100
	}

	db, err := sql.Open("sqlite", s.localMetaMCPDBPath())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	args := []any{}
	query := `
		SELECT id, level, message, source, url, line_number, timestamp
		FROM browser_console_logs`
	if strings.TrimSpace(level) != "" {
		query += " WHERE level = ?"
		args = append(args, level)
	}
	query += " ORDER BY timestamp DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := []map[string]any{}
	searchLower := strings.ToLower(strings.TrimSpace(search))
	for rows.Next() {
		var (
			id         string
			levelValue string
			message    string
			source     string
			urlValue   sql.NullString
			lineNumber sql.NullInt64
			timestamp  int64
		)
		if err := rows.Scan(&id, &levelValue, &message, &source, &urlValue, &lineNumber, &timestamp); err != nil {
			return nil, err
		}
		if searchLower != "" && !strings.Contains(strings.ToLower(message), searchLower) {
			continue
		}
		logs = append(logs, map[string]any{
			"id":         id,
			"level":      levelValue,
			"message":    message,
			"source":     source,
			"url":        nullStringToAny(urlValue),
			"lineNumber": nullInt64ToAny(lineNumber),
			"timestamp":  timestamp,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	total := len(logs)
	if len(logs) > limit {
		logs = logs[:limit]
	}

	return map[string]any{
		"logs":  logs,
		"total": total,
	}, nil
}

func (s *Server) localBrowserControlsStats() (any, error) {
	db, err := sql.Open("sqlite", s.localMetaMCPDBPath())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	historyRows, err := db.Query(`SELECT domain FROM browser_history`)
	if err != nil {
		return nil, err
	}
	defer historyRows.Close()

	historyCount := 0
	domains := map[string]struct{}{}
	for historyRows.Next() {
		var domain string
		if err := historyRows.Scan(&domain); err != nil {
			return nil, err
		}
		historyCount++
		domains[domain] = struct{}{}
	}
	if err := historyRows.Err(); err != nil {
		return nil, err
	}

	logRows, err := db.Query(`SELECT level FROM browser_console_logs`)
	if err != nil {
		return nil, err
	}
	defer logRows.Close()

	consoleLogCount := 0
	consoleErrors := 0
	for logRows.Next() {
		var level string
		if err := logRows.Scan(&level); err != nil {
			return nil, err
		}
		consoleLogCount++
		if level == "error" {
			consoleErrors++
		}
	}
	if err := logRows.Err(); err != nil {
		return nil, err
	}

	return map[string]any{
		"historyCount":    historyCount,
		"uniqueDomains":   len(domains),
		"consoleLogCount": consoleLogCount,
		"consoleErrors":   consoleErrors,
	}, nil
}

func (s *Server) localBrowserExtensionMemories(search, tag string, limit, offset int) (any, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	db, err := sql.Open("sqlite", s.localMetaMCPDBPath())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT id, url, normalized_url, title, content, selected_text, tags, favicon, source, content_hash, saved_at
		FROM web_memories
		ORDER BY saved_at DESC
	`)
	if err != nil {
		if strings.Contains(err.Error(), "no such table: web_memories") {
			return map[string]any{
				"items": []map[string]any{},
				"total": 0,
			}, nil
		}
		return nil, err
	}
	defer rows.Close()

	searchLower := strings.ToLower(strings.TrimSpace(search))
	tagValue := strings.TrimSpace(tag)
	items := make([]map[string]any, 0)
	for rows.Next() {
		var (
			id           string
			urlValue     string
			normalized   string
			title        string
			content      string
			selectedText sql.NullString
			tagsRaw      string
			favicon      sql.NullString
			source       string
			contentHash  string
			savedAt      int64
		)
		if err := rows.Scan(&id, &urlValue, &normalized, &title, &content, &selectedText, &tagsRaw, &favicon, &source, &contentHash, &savedAt); err != nil {
			return nil, err
		}

		tagsValue := jsonArrayOrEmpty(tagsRaw)
		tagList := jsonArrayStrings(tagsValue)
		if searchLower != "" &&
			!strings.Contains(strings.ToLower(title), searchLower) &&
			!strings.Contains(strings.ToLower(urlValue), searchLower) &&
			!strings.Contains(strings.ToLower(content), searchLower) {
			continue
		}
		if tagValue != "" && !stringSliceContains(tagList, tagValue) {
			continue
		}

		items = append(items, map[string]any{
			"id":            id,
			"url":           urlValue,
			"normalizedUrl": normalized,
			"title":         title,
			"content":       content,
			"selectedText":  nullStringToAny(selectedText),
			"tags":          tagsValue,
			"favicon":       nullStringToAny(favicon),
			"savedAt":       unixTimestampToRFC3339(savedAt),
			"source":        source,
			"contentHash":   contentHash,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	total := len(items)
	if offset >= total {
		return map[string]any{
			"items": []map[string]any{},
			"total": total,
		}, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return map[string]any{
		"items": items[offset:end],
		"total": total,
	}, nil
}

func (s *Server) localBrowserExtensionStats() (any, error) {
	db, err := sql.Open("sqlite", s.localMetaMCPDBPath())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT normalized_url, tags
		FROM web_memories
	`)
	if err != nil {
		if strings.Contains(err.Error(), "no such table: web_memories") {
			return map[string]any{
				"totalMemories": 0,
				"uniqueUrls":    0,
				"topTags":       []map[string]any{},
			}, nil
		}
		return nil, err
	}
	defer rows.Close()

	totalMemories := 0
	uniqueURLs := make(map[string]struct{})
	tagCounts := make(map[string]int)
	for rows.Next() {
		var (
			normalized string
			tagsRaw    string
		)
		if err := rows.Scan(&normalized, &tagsRaw); err != nil {
			return nil, err
		}
		totalMemories++
		uniqueURLs[normalized] = struct{}{}
		for _, tag := range jsonArrayStrings(jsonArrayOrEmpty(tagsRaw)) {
			tagCounts[tag]++
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	type tagCount struct {
		tag   string
		count int
	}
	topTags := make([]tagCount, 0, len(tagCounts))
	for tag, count := range tagCounts {
		topTags = append(topTags, tagCount{tag: tag, count: count})
	}
	sort.Slice(topTags, func(i, j int) bool {
		if topTags[i].count == topTags[j].count {
			return topTags[i].tag < topTags[j].tag
		}
		return topTags[i].count > topTags[j].count
	})
	if len(topTags) > 20 {
		topTags = topTags[:20]
	}

	serializedTopTags := make([]map[string]any, 0, len(topTags))
	for _, entry := range topTags {
		serializedTopTags = append(serializedTopTags, map[string]any{
			"tag":   entry.tag,
			"count": entry.count,
		})
	}

	return map[string]any{
		"totalMemories": totalMemories,
		"uniqueUrls":    len(uniqueURLs),
		"topTags":       serializedTopTags,
	}, nil
}

func (s *Server) localUnifiedDirectoryList(limit, offset int, search, source string, showDuplicates, duplicatesOnly bool, researchStatus string) (any, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	wantCatalog := source == "" || source == "all" || source == "catalog"
	wantBacklog := source == "" || source == "all" || source == "backlog"
	effectiveShowDuplicates := showDuplicates || duplicatesOnly
	fetchWindow := offset + limit*3
	if fetchWindow < 200 {
		fetchWindow = 200
	}
	if fetchWindow > 1000 {
		fetchWindow = 1000
	}

	catalogItems := []map[string]any{}
	backlogItems := []map[string]any{}
	catalogTotal := int64(0)
	backlogTotal := int64(0)

	if wantCatalog {
		catalogListRaw, err := s.localCatalogList(fetchWindow, 0, search, "", "", "")
		if err != nil {
			return nil, err
		}
		catalogList := catalogListRaw.(map[string]any)
		if servers, ok := catalogList["servers"].([]map[string]any); ok {
			for _, server := range servers {
				catalogItems = append(catalogItems, normalizeUnifiedCatalogItem(server))
			}
		}
		catalogTotal = anyInt64(catalogList["total"])
	}

	if wantBacklog {
		backlogListRaw, err := s.localLinksBacklogList(fetchWindow, 0, search, "", researchStatus, "", effectiveShowDuplicates)
		if err != nil {
			return nil, err
		}
		backlogList := backlogListRaw.(map[string]any)
		if items, ok := backlogList["items"].([]map[string]any); ok {
			for _, item := range items {
				if duplicatesOnly && !anyBool(item["is_duplicate"]) {
					continue
				}
				backlogItems = append(backlogItems, normalizeUnifiedBacklogItem(item))
			}
		}
		backlogTotal = anyInt64(backlogList["total"])
		if duplicatesOnly {
			backlogTotal = int64(len(backlogItems))
		}
	}

	merged := append([]map[string]any{}, catalogItems...)
	merged = append(merged, backlogItems...)
	slices.SortStableFunc(merged, func(a, b map[string]any) int {
		bTime := unifiedDirectoryTimeValue(b)
		aTime := unifiedDirectoryTimeValue(a)
		if bTime < aTime {
			return -1
		}
		if bTime > aTime {
			return 1
		}
		return 0
	})

	start := offset
	if start > len(merged) {
		start = len(merged)
	}
	end := offset + limit
	if end > len(merged) {
		end = len(merged)
	}

	return map[string]any{
		"items": merged[start:end],
		"total": catalogTotal + backlogTotal,
		"totals": map[string]any{
			"catalog": catalogTotal,
			"backlog": backlogTotal,
		},
	}, nil
}

func (s *Server) localUnifiedDirectoryStats() (any, error) {
	catalogStatsRaw, err := s.localCatalogStats()
	if err != nil {
		return nil, err
	}
	backlogStatsRaw, err := s.localLinksBacklogStats()
	if err != nil {
		return nil, err
	}

	catalogStats := catalogStatsRaw.(map[string]any)
	backlogStats := backlogStatsRaw.(map[string]any)

	return map[string]any{
		"catalog": map[string]any{
			"total":       catalogStats["total"],
			"validated":   catalogStats["validated"],
			"broken":      catalogStats["broken"],
			"updated_24h": catalogStats["recentlyUpdated"],
		},
		"backlog":        backlogStats,
		"combined_total": anyInt64(catalogStats["total"]) + anyInt64(backlogStats["total"]),
	}, nil
}

func (s *Server) localWorkflowCanvases() ([]map[string]any, error) {
	db, err := sql.Open("sqlite", s.localMetaMCPDBPath())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT id, name, description, nodes_json, edges_json, user_id, created_at, updated_at
		FROM workflows
		ORDER BY updated_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	workflows := []map[string]any{}
	for rows.Next() {
		workflow, scanErr := scanWorkflowCanvas(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		workflows = append(workflows, workflow)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return workflows, nil
}

func (s *Server) localWorkflowCanvas(id string) (any, error) {
	db, err := sql.Open("sqlite", s.localMetaMCPDBPath())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	row := db.QueryRow(`
		SELECT id, name, description, nodes_json, edges_json, user_id, created_at, updated_at
		FROM workflows
		WHERE id = ?
		LIMIT 1
	`, id)
	workflow, err := scanWorkflowCanvas(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return workflow, nil
}

func (s *Server) localAuditLogs(filter localAuditFilter) ([]map[string]any, error) {
	logPath := filepath.Join(s.cfg.WorkspaceRoot, ".borg", "audit", "audit-"+time.Now().UTC().Format("2006-01-02")+".jsonl")
	raw, err := os.ReadFile(logPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []map[string]any{}, nil
		}
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	if len(lines) == 1 && strings.TrimSpace(lines[0]) == "" {
		return []map[string]any{}, nil
	}

	results := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var entry map[string]any
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		if filter.level != "" {
			if level, _ := entry["level"].(string); level != filter.level {
				continue
			}
		}
		if filter.agentID != "" {
			if agentID, _ := entry["agentId"].(string); agentID != filter.agentID {
				continue
			}
		}
		if filter.action != "" {
			if action, _ := entry["action"].(string); action != filter.action {
				continue
			}
		}
		results = append(results, entry)
	}

	if filter.limit <= 0 {
		filter.limit = 100
	}
	if len(results) > filter.limit {
		results = results[len(results)-filter.limit:]
	}
	return results, nil
}

func unixTimestampToRFC3339(value int64) string {
	if value <= 0 {
		return time.Unix(0, 0).UTC().Format(time.RFC3339)
	}
	return time.Unix(value, 0).UTC().Format(time.RFC3339)
}

func nullStringToAny(value sql.NullString) any {
	if value.Valid {
		return value.String
	}
	return nil
}

func nullInt64ToAny(value sql.NullInt64) any {
	if value.Valid {
		return value.Int64
	}
	return nil
}

func nullTimestampToAny(value sql.NullInt64) any {
	if value.Valid {
		return unixTimestampToRFC3339(value.Int64)
	}
	return nil
}

func jsonArrayOrEmpty(raw string) any {
	var value any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return []any{}
	}
	return value
}

func jsonArrayStrings(value any) []string {
	items, ok := value.([]any)
	if !ok {
		return []string{}
	}
	results := make([]string, 0, len(items))
	for _, item := range items {
		text, ok := item.(string)
		if !ok {
			continue
		}
		results = append(results, text)
	}
	return results
}

func stringSliceContains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func jsonObjectOrEmpty(raw string) any {
	var value any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return map[string]any{}
	}
	return value
}

func jsonObjectOrNil(raw string) any {
	var value any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return nil
	}
	return value
}

func localPublishedCatalogServer(db *sql.DB, uuid string) (any, error) {
	row := db.QueryRow(`
		SELECT uuid, canonical_id, display_name, description, author, repository_url, homepage_url, icon_url,
		       transport, install_method, auth_model, status, confidence, tags, categories, stars,
		       last_seen_at, last_verified_at, created_at, updated_at
		FROM published_mcp_servers
		WHERE uuid = ?
		LIMIT 1
	`, uuid)
	server, err := scanPublishedCatalogServer(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return server, nil
}

func localPublishedCatalogLatestRun(db *sql.DB, serverUUID string) (any, error) {
	row := db.QueryRow(`
		SELECT uuid, server_uuid, run_mode, started_at, finished_at, outcome, failure_class, tool_count,
		       findings_summary, performed_by, created_at
		FROM published_mcp_validation_runs
		WHERE server_uuid = ?
		ORDER BY created_at DESC
		LIMIT 1
	`, serverUUID)
	run, err := scanPublishedCatalogRun(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return run, nil
}

func localPublishedCatalogActiveRecipe(db *sql.DB, serverUUID string) (any, error) {
	var (
		uuid             string
		recipeServerUUID string
		recipeVersion    int64
		templateRaw      string
		requiredSecrets  string
		requiredEnvRaw   string
		confidence       int64
		explanation      sql.NullString
		isActive         bool
		generatedBy      string
		createdAtRaw     int64
		updatedAtRaw     int64
	)

	row := db.QueryRow(`
		SELECT uuid, server_uuid, recipe_version, template, required_secrets, required_env, confidence,
		       explanation, is_active, generated_by, created_at, updated_at
		FROM published_mcp_config_recipes
		WHERE server_uuid = ? AND is_active = 1
		ORDER BY recipe_version DESC
		LIMIT 1
	`, serverUUID)
	if err := row.Scan(
		&uuid, &recipeServerUUID, &recipeVersion, &templateRaw, &requiredSecrets, &requiredEnvRaw, &confidence,
		&explanation, &isActive, &generatedBy, &createdAtRaw, &updatedAtRaw,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return map[string]any{
		"uuid":             uuid,
		"server_uuid":      recipeServerUUID,
		"recipe_version":   recipeVersion,
		"template":         jsonObjectOrEmpty(templateRaw),
		"required_secrets": jsonArrayOrEmpty(requiredSecrets),
		"required_env":     jsonObjectOrEmpty(requiredEnvRaw),
		"confidence":       confidence,
		"explanation":      nullStringToAny(explanation),
		"is_active":        isActive,
		"generated_by":     generatedBy,
		"created_at":       unixTimestampToRFC3339(createdAtRaw),
		"updated_at":       unixTimestampToRFC3339(updatedAtRaw),
	}, nil
}

func localPublishedCatalogSources(db *sql.DB, serverUUID string) ([]map[string]any, error) {
	rows, err := db.Query(`
		SELECT uuid, source_name, source_url, first_seen_at, last_seen_at
		FROM published_mcp_server_sources
		WHERE server_uuid = ?
	`, serverUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sources := []map[string]any{}
	for rows.Next() {
		var (
			uuid        string
			sourceName  string
			sourceURL   sql.NullString
			firstSeenAt int64
			lastSeenAt  int64
		)
		if err := rows.Scan(&uuid, &sourceName, &sourceURL, &firstSeenAt, &lastSeenAt); err != nil {
			return nil, err
		}
		sources = append(sources, map[string]any{
			"uuid":          uuid,
			"source_name":   sourceName,
			"source_url":    nullStringToAny(sourceURL),
			"first_seen_at": unixTimestampToRFC3339(firstSeenAt),
			"last_seen_at":  unixTimestampToRFC3339(lastSeenAt),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return sources, nil
}

func countPublishedCatalogServers(db *sql.DB, status string) (int64, error) {
	var row *sql.Row
	if strings.TrimSpace(status) == "" {
		row = db.QueryRow(`SELECT count(*) FROM published_mcp_servers`)
	} else {
		row = db.QueryRow(`SELECT count(*) FROM published_mcp_servers WHERE status = ?`, status)
	}

	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func countPublishedCatalogServersFiltered(db *sql.DB, search, status, transport, installMethod string) (int64, error) {
	args := []any{}
	where := buildPublishedCatalogWhere(&args, search, status, transport, installMethod, false)
	query := `SELECT count(*) FROM published_mcp_servers`
	if where != "" {
		query += " WHERE " + where
	}

	row := db.QueryRow(query, args...)
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func listPublishedCatalogServers(db *sql.DB, limit, offset int, search, status, transport, installMethod string) ([]map[string]any, error) {
	args := []any{}
	where := buildPublishedCatalogWhere(&args, search, status, transport, installMethod, true)
	query := `
		SELECT uuid, canonical_id, display_name, description, author, repository_url, homepage_url, icon_url,
		       transport, install_method, auth_model, status, confidence, tags, categories, stars,
		       last_seen_at, last_verified_at, created_at, updated_at
		FROM published_mcp_servers`
	if where != "" {
		query += " WHERE " + where
	}
	query += " ORDER BY confidence DESC, updated_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	servers := []map[string]any{}
	for rows.Next() {
		server, scanErr := scanPublishedCatalogServer(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		servers = append(servers, server)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return servers, nil
}

func buildPublishedCatalogWhere(args *[]any, search, status, transport, installMethod string, includeAuthorInSearch bool) string {
	conditions := []string{}
	if strings.TrimSpace(status) != "" {
		conditions = append(conditions, "status = ?")
		*args = append(*args, status)
	}
	if strings.TrimSpace(transport) != "" {
		conditions = append(conditions, "transport = ?")
		*args = append(*args, transport)
	}
	if strings.TrimSpace(installMethod) != "" {
		conditions = append(conditions, "install_method = ?")
		*args = append(*args, installMethod)
	}
	if trimmed := strings.TrimSpace(search); trimmed != "" {
		term := "%" + trimmed + "%"
		searchConditions := []string{"display_name LIKE ?", "description LIKE ?", "canonical_id LIKE ?"}
		searchArgs := []any{term, term, term}
		if includeAuthorInSearch {
			searchConditions = append(searchConditions, "author LIKE ?")
			searchArgs = append(searchArgs, term)
		}
		conditions = append(conditions, "("+strings.Join(searchConditions, " OR ")+")")
		*args = append(*args, searchArgs...)
	}
	return strings.Join(conditions, " AND ")
}

func countPublishedCatalogRecentlyUpdated(db *sql.DB, hours int) (int64, error) {
	if hours < 1 {
		hours = 1
	}
	threshold := time.Now().UTC().Add(-time.Duration(hours) * time.Hour).Unix()
	row := db.QueryRow(`SELECT count(*) FROM published_mcp_servers WHERE updated_at >= ?`, threshold)

	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func normalizeUnifiedCatalogItem(server map[string]any) map[string]any {
	return map[string]any{
		"source":         "catalog",
		"id":             anyString(server["uuid"]),
		"updated_at":     server["updated_at"],
		"created_at":     server["created_at"],
		"title":          anyString(server["display_name"]),
		"subtitle":       anyStringOrNil(server["canonical_id"]),
		"description":    anyStringOrNil(server["description"]),
		"status":         anyStringOrNil(server["status"]),
		"transport":      anyStringOrNil(server["transport"]),
		"install_method": anyStringOrNil(server["install_method"]),
		"url":            firstNonEmptyString(server["repository_url"], server["homepage_url"]),
		"tags":           mergeStringLists(server["categories"], server["tags"]),
		"confidence":     server["confidence"],
		"is_duplicate":   nil,
		"duplicate_of":   nil,
	}
}

func normalizeUnifiedBacklogItem(item map[string]any) map[string]any {
	title := firstNonEmptyString(item["title"], item["page_title"], item["normalized_url"])
	description := firstNonEmptyString(item["description"], item["page_description"])
	return map[string]any{
		"source":         "backlog",
		"id":             anyString(item["uuid"]),
		"updated_at":     item["updated_at"],
		"created_at":     item["created_at"],
		"title":          title,
		"subtitle":       anyStringOrNil(item["normalized_url"]),
		"description":    description,
		"status":         anyStringOrNil(item["research_status"]),
		"transport":      nil,
		"install_method": nil,
		"url":            anyStringOrNil(item["url"]),
		"tags":           normalizeStringList(item["tags"]),
		"confidence":     nil,
		"is_duplicate":   item["is_duplicate"],
		"duplicate_of":   item["duplicate_of"],
	}
}

func unifiedDirectoryTimeValue(item map[string]any) int64 {
	if updated := parseRFC3339Any(item["updated_at"]); updated != 0 {
		return updated
	}
	return parseRFC3339Any(item["created_at"])
}

func parseRFC3339Any(value any) int64 {
	if s, ok := value.(string); ok && strings.TrimSpace(s) != "" {
		parsed, err := time.Parse(time.RFC3339, s)
		if err == nil {
			return parsed.UnixMilli()
		}
	}
	return 0
}

func firstNonEmptyString(values ...any) any {
	for _, value := range values {
		if s, ok := value.(string); ok && strings.TrimSpace(s) != "" {
			return s
		}
	}
	return nil
}

func normalizeStringList(value any) []string {
	switch typed := value.(type) {
	case []string:
		return typed
	case []any:
		out := []string{}
		for _, item := range typed {
			if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
				out = append(out, s)
			}
		}
		return out
	default:
		return []string{}
	}
}

func mergeStringLists(a, b any) []string {
	out := append([]string{}, normalizeStringList(a)...)
	out = append(out, normalizeStringList(b)...)
	return out
}

func anyString(value any) string {
	if s, ok := value.(string); ok {
		return s
	}
	return ""
}

func anyStringOrNil(value any) any {
	if s, ok := value.(string); ok && strings.TrimSpace(s) != "" {
		return s
	}
	return nil
}

func anyInt64(value any) int64 {
	switch typed := value.(type) {
	case int64:
		return typed
	case int:
		return int64(typed)
	case float64:
		return int64(typed)
	default:
		return 0
	}
}

func anyBool(value any) bool {
	if b, ok := value.(bool); ok {
		return b
	}
	return false
}

type workflowCanvasScanner interface {
	Scan(dest ...any) error
}

func scanWorkflowCanvas(scanner workflowCanvasScanner) (map[string]any, error) {
	var (
		id          string
		name        string
		description sql.NullString
		nodesRaw    string
		edgesRaw    string
		userID      string
		createdAt   int64
		updatedAt   int64
	)
	if err := scanner.Scan(&id, &name, &description, &nodesRaw, &edgesRaw, &userID, &createdAt, &updatedAt); err != nil {
		return nil, err
	}

	return map[string]any{
		"id":          id,
		"name":        name,
		"description": nullStringToAny(description),
		"nodes_json":  jsonArrayOrEmpty(nodesRaw),
		"edges_json":  jsonArrayOrEmpty(edgesRaw),
		"user_id":     userID,
		"created_at":  unixTimestampToRFC3339(createdAt),
		"updated_at":  unixTimestampToRFC3339(updatedAt),
	}, nil
}

type localMCPServerScanner interface {
	Scan(dest ...any) error
}

func scanLocalMCPServer(scanner localMCPServerScanner) (map[string]any, error) {
	var (
		uuid                      string
		name                      string
		description               sql.NullString
		serverType                string
		command                   sql.NullString
		argsRaw                   string
		envRaw                    string
		urlValue                  sql.NullString
		errorStatus               string
		createdAtRaw              int64
		bearerToken               sql.NullString
		headersRaw                string
		alwaysOn                  bool
		userID                    string
		sourcePublishedServerUUID sql.NullString
	)

	if err := scanner.Scan(
		&uuid, &name, &description, &serverType, &command, &argsRaw, &envRaw, &urlValue, &errorStatus, &createdAtRaw,
		&bearerToken, &headersRaw, &alwaysOn, &userID, &sourcePublishedServerUUID,
	); err != nil {
		return nil, err
	}

	return map[string]any{
		"uuid":                         uuid,
		"name":                         name,
		"description":                  nullStringToAny(description),
		"type":                         serverType,
		"command":                      nullStringToAny(command),
		"args":                         jsonArrayOrEmpty(argsRaw),
		"env":                          jsonObjectOrEmpty(envRaw),
		"url":                          nullStringToAny(urlValue),
		"error_status":                 errorStatus,
		"created_at":                   unixTimestampToRFC3339(createdAtRaw),
		"bearerToken":                  nullStringToAny(bearerToken),
		"headers":                      jsonObjectOrEmpty(headersRaw),
		"always_on":                    alwaysOn,
		"user_id":                      userID,
		"source_published_server_uuid": nullStringToAny(sourcePublishedServerUUID),
	}, nil
}

type publishedCatalogRunScanner interface {
	Scan(dest ...any) error
}

type publishedCatalogServerScanner interface {
	Scan(dest ...any) error
}

func scanPublishedCatalogServer(scanner publishedCatalogServerScanner) (map[string]any, error) {
	var (
		serverUUID     string
		canonicalID    string
		displayName    string
		description    sql.NullString
		author         sql.NullString
		repositoryURL  sql.NullString
		homepageURL    sql.NullString
		iconURL        sql.NullString
		transport      string
		installMethod  string
		authModel      string
		status         string
		confidence     int64
		tagsRaw        string
		categoriesRaw  string
		stars          sql.NullInt64
		lastSeenAt     sql.NullInt64
		lastVerifiedAt sql.NullInt64
		createdAtRaw   int64
		updatedAtRaw   int64
	)

	if err := scanner.Scan(
		&serverUUID, &canonicalID, &displayName, &description, &author, &repositoryURL, &homepageURL, &iconURL,
		&transport, &installMethod, &authModel, &status, &confidence, &tagsRaw, &categoriesRaw, &stars,
		&lastSeenAt, &lastVerifiedAt, &createdAtRaw, &updatedAtRaw,
	); err != nil {
		return nil, err
	}

	return map[string]any{
		"uuid":             serverUUID,
		"canonical_id":     canonicalID,
		"display_name":     displayName,
		"description":      nullStringToAny(description),
		"author":           nullStringToAny(author),
		"repository_url":   nullStringToAny(repositoryURL),
		"homepage_url":     nullStringToAny(homepageURL),
		"icon_url":         nullStringToAny(iconURL),
		"transport":        transport,
		"install_method":   installMethod,
		"auth_model":       authModel,
		"status":           status,
		"confidence":       confidence,
		"tags":             jsonArrayOrEmpty(tagsRaw),
		"categories":       jsonArrayOrEmpty(categoriesRaw),
		"stars":            nullInt64ToAny(stars),
		"last_seen_at":     nullTimestampToAny(lastSeenAt),
		"last_verified_at": nullTimestampToAny(lastVerifiedAt),
		"created_at":       unixTimestampToRFC3339(createdAtRaw),
		"updated_at":       unixTimestampToRFC3339(updatedAtRaw),
	}, nil
}

func scanPublishedCatalogRun(scanner publishedCatalogRunScanner) (map[string]any, error) {
	var (
		uuid          string
		runServerUUID string
		runMode       string
		startedAtRaw  int64
		finishedAt    sql.NullInt64
		outcome       string
		failureClass  sql.NullString
		toolCount     sql.NullInt64
		findingsRaw   sql.NullString
		performedBy   string
		createdAtRaw  int64
	)

	if err := scanner.Scan(
		&uuid, &runServerUUID, &runMode, &startedAtRaw, &finishedAt, &outcome, &failureClass, &toolCount,
		&findingsRaw, &performedBy, &createdAtRaw,
	); err != nil {
		return nil, err
	}

	var findings any
	if findingsRaw.Valid {
		findings = jsonObjectOrNil(findingsRaw.String)
	}

	return map[string]any{
		"uuid":             uuid,
		"server_uuid":      runServerUUID,
		"run_mode":         runMode,
		"started_at":       unixTimestampToRFC3339(startedAtRaw),
		"finished_at":      nullTimestampToAny(finishedAt),
		"outcome":          outcome,
		"failure_class":    nullStringToAny(failureClass),
		"tool_count":       nullInt64ToAny(toolCount),
		"findings_summary": findings,
		"performed_by":     performedBy,
		"created_at":       unixTimestampToRFC3339(createdAtRaw),
	}, nil
}

func (s *Server) localServerHealth(serverUUID string) map[string]any {
	servers, err := s.localConfiguredMCPServers()
	if err != nil {
		return map[string]any{
			"status":      "HEALTHY",
			"crashCount":  0,
			"maxAttempts": 0,
		}
	}

	for _, server := range servers {
		uuid, _ := server["uuid"].(string)
		meta, _ := server["_meta"].(map[string]any)
		metaUUID, _ := meta["uuid"].(string)
		requestedUUID := strings.TrimSpace(serverUUID)
		if strings.TrimSpace(uuid) != requestedUUID && strings.TrimSpace(metaUUID) != requestedUUID {
			continue
		}
		crashCount := intNumber(meta["crashCount"])
		maxAttempts := intNumber(meta["maxAttempts"])
		statusValue, _ := meta["status"].(string)
		status := "HEALTHY"
		if strings.EqualFold(strings.TrimSpace(statusValue), "failed") || strings.EqualFold(strings.TrimSpace(statusValue), "error") {
			status = "ERROR"
		}
		return map[string]any{
			"status":      status,
			"crashCount":  crashCount,
			"maxAttempts": maxAttempts,
		}
	}

	return map[string]any{
		"status":      "HEALTHY",
		"crashCount":  0,
		"maxAttempts": 0,
	}
}

func localSubmoduleCapabilitiesValues(workspaceRoot string, submodulePath string) ([]string, string) {
	fullPath := filepath.Join(workspaceRoot, filepath.FromSlash(submodulePath))
	caps := []string{}
	startCommand := ""

	packageJSONPath := filepath.Join(fullPath, "package.json")
	if raw, err := os.ReadFile(packageJSONPath); err == nil {
		var parsed map[string]any
		if json.Unmarshal(raw, &parsed) == nil {
			if keywords, ok := parsed["keywords"].([]any); ok {
				for _, rawKeyword := range keywords {
					keyword, _ := rawKeyword.(string)
					if keyword == "mcp-server" {
						caps = append(caps, "mcp-server")
						break
					}
				}
			}
			if dependencies, ok := parsed["dependencies"].(map[string]any); ok {
				if _, exists := dependencies["@modelcontextprotocol/sdk"]; exists {
					caps = appendIfMissing(caps, "mcp-sdk")
				}
			}
			if scripts, ok := parsed["scripts"].(map[string]any); ok {
				if _, exists := scripts["build"]; exists {
					caps = appendIfMissing(caps, "build")
				}
				if _, exists := scripts["start"]; exists {
					startCommand = "npm start"
				}
			}
			if startCommand == "" {
				if binValue, ok := parsed["bin"]; ok {
					switch typed := binValue.(type) {
					case string:
						if strings.TrimSpace(typed) != "" {
							startCommand = "node " + typed
						}
					case map[string]any:
						for _, value := range typed {
							if entry, ok := value.(string); ok && strings.TrimSpace(entry) != "" {
								startCommand = "node " + entry
								break
							}
						}
					}
				}
			}
			if startCommand == "" {
				if mainEntry, ok := parsed["main"].(string); ok && strings.TrimSpace(mainEntry) != "" {
					startCommand = "node " + mainEntry
				}
			}
		}
	}

	requirementsPath := filepath.Join(fullPath, "requirements.txt")
	if fileExists(requirementsPath) {
		caps = appendIfMissing(caps, "python")
		if startCommand == "" {
			if fileExists(filepath.Join(fullPath, "main.py")) {
				startCommand = "python main.py"
			} else if fileExists(filepath.Join(fullPath, "app.py")) {
				startCommand = "python app.py"
			}
		}
	}

	return caps, startCommand
}

func appendIfMissing(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func submoduleStatusFromPath(fullPath string) string {
	if !fileExists(fullPath) {
		return "missing"
	}
	return "clean"
}

func submoduleBuildExists(fullPath string) bool {
	for _, candidate := range []string{"dist", "build", "out"} {
		if fileExists(filepath.Join(fullPath, candidate)) {
			return true
		}
	}
	return false
}

func coalesceSubmoduleName(name string, modulePath string) string {
	if strings.TrimSpace(name) != "" {
		return name
	}
	parts := strings.Split(strings.ReplaceAll(modulePath, "\\", "/"), "/")
	if len(parts) == 0 {
		return modulePath
	}
	return parts[len(parts)-1]
}

func localProjectContext(workspaceRoot string) string {
	const defaultContent = "# Project Context\n\nDefine your repository rules and architectural vision here."
	content, err := os.ReadFile(filepath.Join(workspaceRoot, ".borg", "project_context.md"))
	if err != nil {
		return defaultContent
	}
	return string(content)
}

func localProjectHandoffs(workspaceRoot string) []map[string]any {
	entries, err := os.ReadDir(filepath.Join(workspaceRoot, ".borg", "handoffs"))
	if err != nil {
		return []map[string]any{}
	}

	type handoffFile struct {
		name string
	}
	files := make([]handoffFile, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, "handoff_") {
			files = append(files, handoffFile{name: name})
		}
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].name > files[j].name
	})

	handoffs := make([]map[string]any, 0, len(files))
	for _, file := range files {
		rawTimestamp := strings.TrimSuffix(strings.TrimPrefix(file.name, "handoff_"), ".json")
		timestamp, err := strconv.ParseInt(rawTimestamp, 10, 64)
		if err != nil {
			timestamp = 0
		}
		handoffs = append(handoffs, map[string]any{
			"id":        file.name,
			"timestamp": timestamp,
			"path":      file.name,
		})
	}

	return handoffs
}

func localInfrastructureStatus(workspaceRoot string) map[string]any {
	infraBinary := strings.TrimSpace(os.Getenv("BORG_INFRA_BINARY"))
	if infraBinary == "" {
		infraBinary = "mcpetes"
	}

	infraSubmoduleDir := strings.TrimSpace(os.Getenv("BORG_INFRA_SUBMODULE"))
	if infraSubmoduleDir == "" {
		infraSubmoduleDir = infraBinary
	}

	binPath := filepath.Join(workspaceRoot, "..", "..", "submodules", infraSubmoduleDir, "bin", infraBinary)
	_, binErr := os.Stat(binPath)
	isInstalled := binErr == nil

	configPath := filepath.Join(os.Getenv("USERPROFILE"), ".config", "mcpetes", "config.yaml")
	if strings.TrimSpace(os.Getenv("USERPROFILE")) == "" {
		if home, err := os.UserHomeDir(); err == nil {
			configPath = filepath.Join(home, ".config", "mcpetes", "config.yaml")
		}
	}
	_, configErr := os.Stat(configPath)

	return map[string]any{
		"installed":    isInstalled,
		"hasConfig":    configErr == nil,
		"daemonActive": false,
		"version": func() any {
			if isInstalled {
				return "latest"
			}
			return nil
		}(),
	}
}

func localSettingsEnvironment() map[string]any {
	return map[string]any{
		"nodeVersion": "",
		"platform":    runtime.GOOS,
		"arch":        runtime.GOARCH,
		"cwd":         mustGetwd(),
		"env": map[string]any{
			"NODE_ENV": coalesceEnv("NODE_ENV", "development"),
			"PORT":     coalesceEnv("PORT", "3000"),
		},
	}
}

func localSettingsConfig(workspaceRoot string) map[string]any {
	configPath := filepath.Join(workspaceRoot, ".borg", "config.json")
	raw, err := os.ReadFile(configPath)
	if err != nil {
		return map[string]any{}
	}

	var parsed map[string]any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return map[string]any{}
	}
	if parsed == nil {
		return map[string]any{}
	}
	return parsed
}

func writeLocalSettingsConfig(workspaceRoot string, config map[string]any) error {
	configDir := filepath.Join(workspaceRoot, ".borg")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return err
	}
	configPath := filepath.Join(configDir, "config.json")
	encoded, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, append(encoded, '\n'), 0o644)
}

func (s *Server) localSavedScripts() ([]map[string]any, error) {
	config := localSettingsConfig(s.cfg.WorkspaceRoot)
	rawScripts, ok := config["scripts"].([]any)
	if !ok {
		return []map[string]any{}, nil
	}

	scripts := make([]map[string]any, 0, len(rawScripts))
	for _, entry := range rawScripts {
		script, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		scripts = append(scripts, cloneMap(script))
	}

	return scripts, nil
}

func (s *Server) localCreateSavedScript(payload map[string]any) (any, error) {
	name := strings.TrimSpace(stringValue(payload["name"]))
	code := stringValue(payload["code"])
	if name == "" || strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("missing script name or code")
	}
	description := nullableString(payload["description"])
	config := localSettingsConfig(s.cfg.WorkspaceRoot)
	rawScripts, _ := config["scripts"].([]any)
	newScript := map[string]any{
		"uuid":        uuid.NewString(),
		"name":        name,
		"description": description,
		"code":        code,
	}
	config["scripts"] = append(rawScripts, newScript)
	if err := writeLocalSettingsConfig(s.cfg.WorkspaceRoot, config); err != nil {
		return nil, err
	}
	return newScript, nil
}

func (s *Server) localUpdateSavedScript(payload map[string]any) (any, error) {
	targetUUID := strings.TrimSpace(stringValue(payload["uuid"]))
	if targetUUID == "" {
		return nil, fmt.Errorf("missing script uuid")
	}
	config := localSettingsConfig(s.cfg.WorkspaceRoot)
	rawScripts, _ := config["scripts"].([]any)
	updated := false
	var updatedScript map[string]any
	for index, entry := range rawScripts {
		script, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		if stringValue(script["uuid"]) != targetUUID {
			continue
		}

		nextScript := cloneMap(script)
		if name := strings.TrimSpace(stringValue(payload["name"])); name != "" {
			nextScript["name"] = name
		}
		if _, exists := payload["description"]; exists {
			nextScript["description"] = nullableString(payload["description"])
		}
		if _, exists := payload["code"]; exists {
			code := stringValue(payload["code"])
			if strings.TrimSpace(code) == "" {
				return nil, fmt.Errorf("missing script code")
			}
			nextScript["code"] = code
		}
		rawScripts[index] = nextScript
		updated = true
		updatedScript = nextScript
		break
	}
	if !updated {
		return nil, fmt.Errorf("script not found")
	}
	config["scripts"] = rawScripts
	if err := writeLocalSettingsConfig(s.cfg.WorkspaceRoot, config); err != nil {
		return nil, err
	}
	return updatedScript, nil
}

func (s *Server) localDeleteSavedScript(targetUUID string) (any, error) {
	if strings.TrimSpace(targetUUID) == "" {
		return nil, fmt.Errorf("missing script uuid")
	}
	config := localSettingsConfig(s.cfg.WorkspaceRoot)
	rawScripts, _ := config["scripts"].([]any)
	filtered := make([]any, 0, len(rawScripts))
	deleted := false
	for _, entry := range rawScripts {
		script, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		if stringValue(script["uuid"]) == targetUUID {
			deleted = true
			continue
		}
		filtered = append(filtered, script)
	}
	config["scripts"] = filtered
	if err := writeLocalSettingsConfig(s.cfg.WorkspaceRoot, config); err != nil {
		return nil, err
	}
	return map[string]any{"success": deleted}, nil
}

func (s *Server) localExecuteSavedScript(targetUUID string) (any, error) {
	if strings.TrimSpace(targetUUID) == "" {
		return nil, fmt.Errorf("missing script uuid")
	}
	scripts, err := s.localSavedScripts()
	if err != nil {
		return nil, err
	}
	var script map[string]any
	for _, entry := range scripts {
		if stringValue(entry["uuid"]) == targetUUID {
			script = entry
			break
		}
	}
	if script == nil {
		return nil, fmt.Errorf("script not found")
	}
	if _, err := exec.LookPath("node"); err != nil {
		return nil, fmt.Errorf("node runtime not available for local script execution")
	}

	code := stringValue(script["code"])
	startedAt := time.Now().UTC()
	wrapper := "(async () => {\n" + code + "\n})().catch((error) => {\n  console.error(error instanceof Error ? error.stack || error.message : String(error));\n  process.exitCode = 1;\n});\n"
	cmd := exec.Command("node", "-e", wrapper)
	cmd.Dir = s.cfg.WorkspaceRoot
	output, runErr := cmd.CombinedOutput()
	finishedAt := time.Now().UTC()
	message := strings.TrimSpace(string(output))
	if message == "" && runErr == nil {
		message = "Script executed without console output."
	}
	result := map[string]any{
		"success": runErr == nil,
		"result":  message,
		"execution": map[string]any{
			"scriptUuid": stringValue(script["uuid"]),
			"scriptName": stringValue(script["name"]),
			"startedAt":  startedAt.Format(time.RFC3339),
			"finishedAt": finishedAt.Format(time.RFC3339),
			"durationMs": finishedAt.Sub(startedAt).Milliseconds(),
		},
	}
	if runErr != nil {
		result["error"] = runErr.Error()
	}
	return result, nil
}

func (s *Server) localToolSets() ([]map[string]any, error) {
	db, err := sql.Open("sqlite", s.localMetaMCPDBPath())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT uuid, name, description
		FROM tool_sets
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	toolSets := make([]map[string]any, 0)
	for rows.Next() {
		var (
			uuid        string
			name        string
			description sql.NullString
		)
		if err := rows.Scan(&uuid, &name, &description); err != nil {
			return nil, err
		}

		toolRows, err := db.Query(`
			SELECT tool_uuid
			FROM tool_set_items
			WHERE tool_set_uuid = ?
			ORDER BY created_at ASC, uuid ASC
		`, uuid)
		if err != nil {
			return nil, err
		}

		tools := make([]string, 0)
		for toolRows.Next() {
			var toolUUID string
			if err := toolRows.Scan(&toolUUID); err != nil {
				toolRows.Close()
				return nil, err
			}
			tools = append(tools, toolUUID)
		}
		if err := toolRows.Err(); err != nil {
			toolRows.Close()
			return nil, err
		}
		toolRows.Close()

		toolSets = append(toolSets, map[string]any{
			"uuid":        uuid,
			"name":        name,
			"description": nullStringToAny(description),
			"tools":       tools,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return toolSets, nil
}

func (s *Server) localToolChains() ([]map[string]any, error) {
	db, err := sql.Open("sqlite", s.localMetaMCPDBPath())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT id, name, description, created_at
		FROM tool_chains
		ORDER BY created_at DESC
	`)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no such table") {
			return []map[string]any{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	chains := make([]map[string]any, 0)
	for rows.Next() {
		var (
			id          string
			name        string
			description sql.NullString
			createdAt   int64
		)
		if err := rows.Scan(&id, &name, &description, &createdAt); err != nil {
			return nil, err
		}

		steps, err := localToolChainSteps(db, id)
		if err != nil {
			return nil, err
		}

		chains = append(chains, map[string]any{
			"id":            id,
			"name":          name,
			"description":   nullStringToAny(description),
			"steps":         steps,
			"failurePolicy": "stop",
			"maxRetries":    1,
			"createdAt":     createdAt * 1000,
			"runCount":      0,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return chains, nil
}

func (s *Server) localToolChain(id string) (any, error) {
	db, err := sql.Open("sqlite", s.localMetaMCPDBPath())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var (
		chainID     string
		name        string
		description sql.NullString
		createdAt   int64
	)
	row := db.QueryRow(`
		SELECT id, name, description, created_at
		FROM tool_chains
		WHERE id = ?
		LIMIT 1
	`, id)
	if err := row.Scan(&chainID, &name, &description, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		if strings.Contains(strings.ToLower(err.Error()), "no such table") {
			return nil, nil
		}
		return nil, err
	}

	steps, err := localToolChainSteps(db, chainID)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"id":            chainID,
		"name":          name,
		"description":   nullStringToAny(description),
		"steps":         steps,
		"failurePolicy": "stop",
		"maxRetries":    1,
		"createdAt":     createdAt * 1000,
		"runCount":      0,
	}, nil
}

func localToolChainSteps(db *sql.DB, chainID string) ([]map[string]any, error) {
	rows, err := db.Query(`
		SELECT tool_name, arguments_template
		FROM tool_chain_steps
		WHERE chain_id = ?
		ORDER BY step_order ASC
	`, chainID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no such table") {
			return []map[string]any{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	steps := make([]map[string]any, 0)
	for rows.Next() {
		var (
			toolName             string
			argumentsTemplateRaw string
		)
		if err := rows.Scan(&toolName, &argumentsTemplateRaw); err != nil {
			return nil, err
		}

		inputMapping := map[string]any{}
		if strings.TrimSpace(argumentsTemplateRaw) != "" {
			if err := json.Unmarshal([]byte(argumentsTemplateRaw), &inputMapping); err != nil {
				inputMapping = map[string]any{}
			}
		}

		steps = append(steps, map[string]any{
			"toolName":     toolName,
			"inputMapping": inputMapping,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return steps, nil
}

func (s *Server) localToolAliases() ([]map[string]any, error) {
	db, err := sql.Open("sqlite", s.localMetaMCPDBPath())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT alias, target_tool, description, created_at
		FROM tool_aliases
		ORDER BY created_at DESC
	`)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no such table") {
			return []map[string]any{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	aliases := make([]map[string]any, 0)
	for rows.Next() {
		var (
			alias       string
			targetTool  string
			description sql.NullString
			createdAt   int64
		)
		if err := rows.Scan(&alias, &targetTool, &description, &createdAt); err != nil {
			return nil, err
		}

		aliases = append(aliases, map[string]any{
			"serverId":     "unknown",
			"originalName": targetTool,
			"alias":        alias,
			"description":  nullStringToAny(description),
			"createdAt":    createdAt * 1000,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return aliases, nil
}

func (s *Server) localToolAlias(name string) (any, error) {
	db, err := sql.Open("sqlite", s.localMetaMCPDBPath())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var (
		alias       string
		targetTool  string
		description sql.NullString
		createdAt   int64
	)
	row := db.QueryRow(`
		SELECT alias, target_tool, description, created_at
		FROM tool_aliases
		WHERE alias = ?
		LIMIT 1
	`, name)
	if err := row.Scan(&alias, &targetTool, &description, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return map[string]any{"resolved": false}, nil
		}
		if strings.Contains(strings.ToLower(err.Error()), "no such table") {
			return map[string]any{"resolved": false}, nil
		}
		return nil, err
	}

	return map[string]any{
		"resolved":     true,
		"serverId":     "unknown",
		"originalName": targetTool,
		"alias":        alias,
		"description":  nullStringToAny(description),
		"createdAt":    createdAt * 1000,
	}, nil
}

type localDBToolScanner interface {
	Scan(dest ...any) error
}

func scanLocalDBTool(scanner localDBToolScanner) (map[string]any, error) {
	var (
		uuid          string
		name          string
		description   sql.NullString
		toolSchemaRaw string
		isDeferred    bool
		alwaysOn      bool
		serverUUID    string
		serverName    sql.NullString
	)
	if err := scanner.Scan(&uuid, &name, &description, &toolSchemaRaw, &isDeferred, &alwaysOn, &serverUUID, &serverName); err != nil {
		return nil, err
	}

	inputSchema := jsonObjectOrEmpty(toolSchemaRaw)
	schemaParamCount := 0
	if inputSchemaMap, ok := inputSchema.(map[string]any); ok {
		if properties, ok := inputSchemaMap["properties"].(map[string]any); ok {
			schemaParamCount = len(properties)
		}
	}

	server := "unknown"
	if serverName.Valid && strings.TrimSpace(serverName.String) != "" {
		server = serverName.String
	}

	return map[string]any{
		"uuid":             name,
		"name":             name,
		"description":      nullStringToAny(description),
		"server":           server,
		"inputSchema":      inputSchema,
		"isDeferred":       isDeferred,
		"schemaParamCount": schemaParamCount,
		"mcpServerUuid":    serverUUID,
		"always_on":        alwaysOn,
	}, nil
}

func (s *Server) localDBTools() ([]map[string]any, error) {
	db, err := sql.Open("sqlite", s.localMetaMCPDBPath())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT t.uuid, t.name, t.description, t.tool_schema, t.is_deferred, t.always_on, t.mcp_server_uuid, s.name
		FROM tools t
		LEFT JOIN mcp_servers s ON s.uuid = t.mcp_server_uuid
		ORDER BY t.name
	`)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no such table") {
			return []map[string]any{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	tools := make([]map[string]any, 0)
	for rows.Next() {
		tool, err := scanLocalDBTool(rows)
		if err != nil {
			return nil, err
		}
		tools = append(tools, tool)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tools, nil
}

func (s *Server) localDBToolsByServer(serverID string) ([]map[string]any, error) {
	tools, err := s.localDBTools()
	if err != nil {
		return nil, err
	}
	filtered := make([]map[string]any, 0)
	for _, tool := range tools {
		if stringValue(tool["mcpServerUuid"]) == serverID || stringValue(tool["server"]) == serverID {
			filtered = append(filtered, tool)
		}
	}
	return filtered, nil
}

func (s *Server) localDBToolSearch(query string, limit int) ([]map[string]any, error) {
	tools, err := s.localDBTools()
	if err != nil {
		return nil, err
	}
	queryLower := strings.ToLower(strings.TrimSpace(query))
	results := make([]map[string]any, 0)
	for _, tool := range tools {
		name := strings.ToLower(stringValue(tool["name"]))
		description := strings.ToLower(stringValue(tool["description"]))
		server := strings.ToLower(stringValue(tool["server"]))
		if strings.Contains(name, queryLower) || strings.Contains(description, queryLower) || strings.Contains(server, queryLower) {
			results = append(results, tool)
			if limit > 0 && len(results) >= limit {
				break
			}
		}
	}
	return results, nil
}

func (s *Server) localDBTool(uuid string) (any, error) {
	tools, err := s.localDBTools()
	if err != nil {
		return nil, err
	}
	for _, tool := range tools {
		if stringValue(tool["uuid"]) == uuid || stringValue(tool["name"]) == uuid {
			return tool, nil
		}
	}
	return nil, nil
}

func (s *Server) localShellQueryHistory(query string, limit int) ([]map[string]any, error) {
	historyPath := filepath.Join(s.cfg.WorkspaceRoot, ".borg", "shell_history.json")
	raw, err := os.ReadFile(historyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []map[string]any{}, nil
		}
		return nil, err
	}

	var entries []map[string]any
	if err := json.Unmarshal(raw, &entries); err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	filtered := make([]map[string]any, 0, len(entries))
	for index := len(entries) - 1; index >= 0; index-- {
		entry := entries[index]
		command := strings.ToLower(stringValue(entry["command"]))
		output := strings.ToLower(stringValue(entry["outputSnippet"]))
		if strings.Contains(command, query) || strings.Contains(output, query) {
			filtered = append(filtered, cloneMap(entry))
			if len(filtered) >= limit {
				break
			}
		}
	}

	return filtered, nil
}

func (s *Server) localShellSystemHistory(limit int) ([]string, error) {
	historyPath := shellHistoryPath()
	raw, err := os.ReadFile(historyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	lines := strings.Split(strings.ReplaceAll(string(raw), "\r\n", "\n"), "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			filtered = append(filtered, line)
		}
	}
	if len(filtered) <= limit {
		return filtered, nil
	}
	return filtered[len(filtered)-limit:], nil
}

func shellHistoryPath() string {
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return ".bash_history"
	}
	if runtime.GOOS == "windows" {
		return filepath.Join(home, "AppData", "Roaming", "Microsoft", "Windows", "PowerShell", "PSReadLine", "ConsoleHost_history.txt")
	}
	return filepath.Join(home, ".bash_history")
}

func mustGetwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return cwd
}

func coalesceEnv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func normalizeResearchURL(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return strings.TrimSpace(raw)
	}
	parsed.Fragment = ""
	cleanPath := strings.TrimRight(parsed.Path, "/")
	if cleanPath == "" {
		cleanPath = "/"
	}
	parsed.Path = cleanPath
	normalized := parsed.String()
	return strings.TrimRight(normalized, "/")
}

func (s *Server) localResearchQueue() (map[string]any, error) {
	statusPath := filepath.Join(s.cfg.WorkspaceRoot, "scripts", "ingestion-status.json")
	indexPath := filepath.Join(s.cfg.WorkspaceRoot, "BORG_MASTER_INDEX.jsonc")

	var statusDoc struct {
		Processed []string `json:"processed"`
		Pending   []string `json:"pending"`
		Failed    []struct {
			URL           string `json:"url"`
			Error         string `json:"error"`
			Source        string `json:"source"`
			Attempts      int    `json:"attempts"`
			LastAttemptAt string `json:"last_attempt_at"`
		} `json:"failed"`
	}
	if raw, err := os.ReadFile(statusPath); err == nil {
		_ = json.Unmarshal(raw, &statusDoc)
	}

	metaByURL := map[string]map[string]any{}
	if raw, err := os.ReadFile(indexPath); err == nil {
		var parsed struct {
			Categories map[string][]map[string]any `json:"categories"`
		}
		if err := json.Unmarshal([]byte(stripJSONCLineComments(string(raw))), &parsed); err == nil {
			for category, items := range parsed.Categories {
				for _, item := range items {
					rawURL, _ := item["url"].(string)
					normalized := normalizeResearchURL(rawURL)
					if normalized == "" {
						continue
					}
					name, _ := item["name"].(string)
					if strings.TrimSpace(name) == "" {
						name = normalized
					}
					metaByURL[normalized] = map[string]any{
						"id":       item["id"],
						"name":     name,
						"category": category,
					}
				}
			}
		}
	}

	buildEntry := func(rawURL string) map[string]any {
		normalized := normalizeResearchURL(rawURL)
		meta := metaByURL[normalized]
		if meta == nil {
			meta = map[string]any{
				"id":       nil,
				"name":     normalized,
				"category": nil,
			}
		}
		return map[string]any{
			"url":      normalized,
			"name":     meta["name"],
			"id":       meta["id"],
			"category": meta["category"],
		}
	}

	processed := make([]map[string]any, 0, len(statusDoc.Processed))
	for _, item := range statusDoc.Processed {
		processed = append(processed, buildEntry(item))
	}

	pending := make([]map[string]any, 0, len(statusDoc.Pending))
	for _, item := range statusDoc.Pending {
		pending = append(pending, buildEntry(item))
	}

	failed := make([]map[string]any, 0, len(statusDoc.Failed))
	for _, item := range statusDoc.Failed {
		entry := buildEntry(item.URL)
		if strings.TrimSpace(item.Error) == "" {
			entry["error"] = "Unknown fetch failure"
		} else {
			entry["error"] = item.Error
		}
		if item.Attempts <= 0 {
			entry["attempts"] = 1
		} else {
			entry["attempts"] = item.Attempts
		}
		if strings.TrimSpace(item.LastAttemptAt) == "" {
			entry["lastAttemptAt"] = nil
		} else {
			entry["lastAttemptAt"] = item.LastAttemptAt
		}
		if strings.TrimSpace(item.Source) == "" {
			entry["source"] = "unknown"
		} else {
			entry["source"] = item.Source
		}
		failed = append(failed, entry)
	}

	return map[string]any{
		"queue": map[string]any{
			"processed": processed,
			"pending":   pending,
			"failed":    failed,
		},
		"totals": map[string]any{
			"processed": len(processed),
			"pending":   len(pending),
			"failed":    len(failed),
		},
		"updatedAt": time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (s *Server) localMCPSummary(ctx context.Context) ([]controlplane.Tool, CLISummary, error) {
	tools, err := s.detector.DetectAll(ctx)
	if err != nil {
		return nil, CLISummary{}, err
	}
	summary := summarizeCLI(s.cfg.WorkspaceRoot, tools)
	return tools, summary, nil
}

func (s *Server) localDetectedCliHarnesses(ctx context.Context) ([]map[string]any, error) {
	tools, err := s.detector.DetectAll(ctx)
	if err != nil {
		return nil, err
	}
	definitions := harnesses.List(s.cfg.WorkspaceRoot, tools)
	results := make([]map[string]any, 0, len(definitions))
	for _, definition := range definitions {
		command := definition.ID
		switch definition.ID {
		case "claude":
			command = "claude-code"
		case "factory-droid":
			command = "droid"
		case "copilot":
			command = "github-copilot-cli"
		}
		result := map[string]any{
			"id":             definition.ID,
			"name":           definitionDescriptionName(definition.ID, definition.Description),
			"command":        command,
			"homepage":       harnessHomepage(definition.ID),
			"docsUrl":        harnessDocsURL(definition.ID),
			"installHint":    harnessInstallHint(definition.ID),
			"category":       harnessCategory(definition.ID),
			"sessionCapable": harnessSessionCapable(definition.ID),
			"installed":      definition.Installed,
			"resolvedPath":   nil,
			"version":        nil,
			"detectionError": nil,
		}
		if definition.ID == "antigravity" {
			result["detectionError"] = "Manual install surface; no PATH-detectable command is currently modeled for this harness."
		} else if !definition.Installed {
			result["detectionError"] = "Command not found on PATH."
		}
		for _, tool := range tools {
			if toolToHarnessID(tool.Type) != definition.ID {
				continue
			}
			if strings.TrimSpace(tool.Path) != "" {
				result["resolvedPath"] = tool.Path
			}
			if strings.TrimSpace(tool.Version) != "" {
				result["version"] = tool.Version
			}
			break
		}
		results = append(results, result)
	}
	sort.Slice(results, func(i, j int) bool {
		leftInstalled, _ := results[i]["installed"].(bool)
		rightInstalled, _ := results[j]["installed"].(bool)
		if leftInstalled != rightInstalled {
			return leftInstalled
		}
		leftName, _ := results[i]["name"].(string)
		rightName, _ := results[j]["name"].(string)
		return leftName < rightName
	})
	return results, nil
}

func (s *Server) localExecutionEnvironment(ctx context.Context) (map[string]any, error) {
	tools, summary, err := s.localMCPSummary(ctx)
	if err != nil {
		return nil, err
	}
	harnessesData, err := s.localDetectedCliHarnesses(ctx)
	if err != nil {
		return nil, err
	}

	shells := localExecutionShells()
	preferredShellID := selectPreferredShellID(shells)
	preferredShellLabel := ""
	verifiedShellCount := 0
	supportsPowerShell := false
	supportsPosixShell := false
	for index := range shells {
		if shells[index]["verified"] == true {
			verifiedShellCount++
			family, _ := shells[index]["family"].(string)
			if family == "powershell" {
				supportsPowerShell = true
			}
			if family == "posix" || family == "wsl" {
				supportsPosixShell = true
			}
		}
		if id, _ := shells[index]["id"].(string); id == preferredShellID {
			shells[index]["preferred"] = true
			preferredShellLabel, _ = shells[index]["name"].(string)
		}
	}

	verifiedToolCount := 0
	toolRows := make([]map[string]any, 0, len(tools))
	for _, tool := range tools {
		if tool.Available && strings.TrimSpace(tool.Version) != "" && tool.Version != "unknown" {
			verifiedToolCount++
		}
		toolRows = append(toolRows, map[string]any{
			"id":           tool.Type,
			"name":         tool.Name,
			"installed":    tool.Available,
			"verified":     tool.Available && strings.TrimSpace(tool.Version) != "",
			"resolvedPath": nullableString(tool.Path),
			"version":      nullableString(tool.Version),
			"capabilities": append([]string(nil), tool.Capabilities...),
			"notes":        toolNotes(tool),
		})
	}

	verifiedHarnessCount := 0
	for _, harness := range harnessesData {
		installed, _ := harness["installed"].(bool)
		detectionError, _ := harness["detectionError"].(string)
		if installed && strings.TrimSpace(detectionError) == "" {
			verifiedHarnessCount++
		}
	}

	notes := []string{}
	if strings.TrimSpace(preferredShellLabel) != "" {
		notes = append(notes, "Prefer "+preferredShellLabel+" for default borg shell execution on this host.")
	}
	if supportsPosixShell {
		for _, shell := range shells {
			verified, _ := shell["verified"].(bool)
			family, _ := shell["family"].(string)
			if verified && (family == "posix" || family == "wsl") {
				name, _ := shell["name"].(string)
				notes = append(notes, name+" is available for POSIX-style pipelines and Unix-first tooling.")
				break
			}
		}
	} else if runtime.GOOS == "windows" {
		notes = append(notes, "No verified POSIX shell detected. Recommendation: Install Cygwin or WSL to ensure 1:1 compatibility with AI model tool training (e.g. bash, grep, sed).")
	}

	return map[string]any{
		"os": runtime.GOOS,
		"summary": map[string]any{
			"ready":                verifiedShellCount > 0,
			"preferredShellId":     emptyStringToNilAny(preferredShellID),
			"preferredShellLabel":  emptyStringToNilAny(preferredShellLabel),
			"shellCount":           installedShellCount(shells),
			"verifiedShellCount":   verifiedShellCount,
			"toolCount":            summary.AvailableToolCount,
			"verifiedToolCount":    verifiedToolCount,
			"harnessCount":         summary.InstalledHarnessCount,
			"verifiedHarnessCount": verifiedHarnessCount,
			"supportsPowerShell":   supportsPowerShell,
			"supportsPosixShell":   supportsPosixShell,
			"notes":                notes,
		},
		"shells":    shells,
		"tools":     toolRows,
		"harnesses": harnessesData,
	}, nil
}

func (s *Server) localInstallSurfaces() []map[string]any {
	workspaceRoot := s.cfg.WorkspaceRoot
	chromiumPath := firstExistingRelativePath(workspaceRoot, []string{
		filepath.Join("apps", "borg-extension", "dist-chromium"),
		filepath.Join("apps", "extension", "dist"),
		filepath.Join("apps", "borg-extension", "dist"),
	})
	firefoxBundlePath := firstExistingRelativePath(workspaceRoot, []string{
		filepath.Join("apps", "borg-extension", "dist-firefox"),
	})
	firefoxManifestPath := firstExistingRelativePath(workspaceRoot, []string{
		filepath.Join("apps", "extension", "manifest.firefox.json"),
	})
	vscodeBuildPath := firstExistingRelativePath(workspaceRoot, []string{
		filepath.Join("packages", "vscode", "dist", "extension.js"),
		filepath.Join("packages", "vscode", "dist"),
	})
	mcpConfigPath := firstExistingRelativePath(workspaceRoot, []string{"mcp.jsonc", "mcp.json"})

	return []map[string]any{
		installSurfaceArtifact("browser-extension-chromium", chromiumPath != "", chromiumPath, chromiumArtifactKind(chromiumPath), map[bool]string{
			true:  "Unpacked Chromium-compatible browser extension output is available.",
			false: "Build the browser extension to generate a Chromium/Edge unpacked bundle.",
		}, packageVersion(workspaceRoot, filepath.Join("apps", "borg-extension", "package.json")), workspaceRoot),
		firefoxInstallSurface(workspaceRoot, firefoxBundlePath, firefoxManifestPath),
		vscodeInstallSurface(workspaceRoot, vscodeBuildPath),
		installSurfaceArtifact("mcp-client-sync", mcpConfigPath != "", mcpConfigPath, mcpConfigArtifactKind(mcpConfigPath), map[bool]string{
			true:  "borg-managed MCP config source is present for dashboard sync and preview flows.",
			false: "No borg MCP config source file was detected yet.",
		}, "", workspaceRoot),
	}
}

func sourceBackedInstalledHarnesses(definitions []harnesses.Definition) []harnesses.Definition {
	filtered := make([]harnesses.Definition, 0, len(definitions))
	for _, definition := range definitions {
		if !definition.Installed || definition.ToolInventoryStatus != "source-backed" {
			continue
		}
		filtered = append(filtered, definition)
	}
	return filtered
}

func sourceBackedToolCount(definitions []harnesses.Definition) int {
	total := 0
	for _, definition := range definitions {
		if !definition.Installed || definition.ToolInventoryStatus != "source-backed" {
			continue
		}
		total += definition.ToolCallCount
	}
	return total
}

func fallbackRuntimeServers(definitions []harnesses.Definition) []map[string]any {
	servers := make([]map[string]any, 0)
	for _, definition := range sourceBackedInstalledHarnesses(definitions) {
		servers = append(servers, map[string]any{
			"name":                definition.ID,
			"runtimeConnected":    false,
			"toolCount":           definition.ToolCallCount,
			"toolInventoryStatus": definition.ToolInventoryStatus,
			"integrationLevel":    definition.IntegrationLevel,
			"source":              definition.ToolSource,
		})
	}
	return servers
}

func fallbackMCPTools(definitions []harnesses.Definition) []map[string]any {
	tools := make([]map[string]any, 0)
	for _, definition := range sourceBackedInstalledHarnesses(definitions) {
		for _, name := range definition.ToolCallNames {
			tools = append(tools, map[string]any{
				"name":         name,
				"server":       definition.ID,
				"alwaysOn":     false,
				"alwaysShow":   false,
				"source":       definition.ToolSource,
				"availability": "source-backed",
			})
		}
	}
	sort.Slice(tools, func(i, j int) bool {
		leftServer, _ := tools[i]["server"].(string)
		rightServer, _ := tools[j]["server"].(string)
		if leftServer == rightServer {
			leftName, _ := tools[i]["name"].(string)
			rightName, _ := tools[j]["name"].(string)
			return leftName < rightName
		}
		return leftServer < rightServer
	})
	return tools
}

func fallbackControlTools(definitions []harnesses.Definition) []map[string]any {
	tools := make([]map[string]any, 0)
	for _, definition := range sourceBackedInstalledHarnesses(definitions) {
		for _, name := range definition.ToolCallNames {
			tools = append(tools, map[string]any{
				"uuid":             name,
				"name":             name,
				"description":      definition.Description,
				"server":           definition.ID,
				"inputSchema":      map[string]any{"type": "object"},
				"isDeferred":       false,
				"schemaParamCount": 0,
				"mcpServerUuid":    definition.ID,
				"always_on":        false,
			})
		}
	}
	sort.Slice(tools, func(i, j int) bool {
		leftServer, _ := tools[i]["server"].(string)
		rightServer, _ := tools[j]["server"].(string)
		if leftServer == rightServer {
			leftName, _ := tools[i]["name"].(string)
			rightName, _ := tools[j]["name"].(string)
			return leftName < rightName
		}
		return leftServer < rightServer
	})
	return tools
}

func definitionDescriptionName(id string, description string) string {
	switch id {
	case "borg":
		return "borg"
	case "claude":
		return "Claude Code"
	case "codex":
		return "OpenAI Codex CLI"
	case "gemini":
		return "Gemini CLI"
	case "opencode":
		return "OpenCode"
	case "cursor":
		return "Cursor CLI"
	case "copilot":
		return "GitHub Copilot CLI"
	case "factory-droid":
		return "Factory Droid"
	case "qwen":
		return "Qwen Code CLI"
	case "superai-cli":
		return "SuperAI CLI"
	case "codebuff":
		return "Codebuff"
	case "codemachine":
		return "CodeMachine"
	case "antigravity":
		return "Antigravity"
	default:
		if strings.TrimSpace(description) == "" {
			return id
		}
		return description
	}
}

func harnessHomepage(id string) string {
	switch id {
	case "borg":
		return "https://github.com/robertpelloni/borg"
	case "aider":
		return "https://aider.chat/"
	case "antigravity":
		return "https://antigravity.google/"
	case "claude":
		return "https://www.anthropic.com/claude-code"
	case "codex":
		return "https://platform.openai.com/docs/guides/codex"
	case "gemini":
		return "https://ai.google.dev/gemini-api/docs/cli"
	case "opencode":
		return "https://opencode.ai/"
	case "cursor":
		return "https://cursor.com/"
	case "copilot":
		return "https://github.com/features/copilot"
	case "goose":
		return "https://block.github.io/goose/"
	case "qwen":
		return "https://chat.qwen.ai/"
	case "superai-cli":
		return "https://superai.dev/"
	case "codebuff":
		return "https://codebuff.com/"
	case "codemachine":
		return "https://codemachine.dev/"
	case "factory-droid":
		return "https://factory.ai/"
	default:
		return "#"
	}
}

func harnessDocsURL(id string) string {
	switch id {
	case "borg":
		return "https://github.com/robertpelloni/borg"
	case "aider":
		return "https://aider.chat/docs/"
	case "antigravity":
		return "https://antigravity.google/docs/home"
	case "claude":
		return "https://docs.anthropic.com/en/docs/claude-code/overview"
	case "codex":
		return "https://platform.openai.com/docs/guides/codex"
	case "gemini":
		return "https://ai.google.dev/gemini-api/docs"
	case "opencode":
		return "https://opencode.ai/docs"
	case "cursor":
		return "https://cursor.com/docs"
	case "copilot":
		return "https://docs.github.com/en/copilot"
	case "goose":
		return "https://block.github.io/goose/docs/"
	case "qwen":
		return "https://qwen.readthedocs.io/"
	case "superai-cli":
		return "https://superai.dev/docs"
	case "codebuff":
		return "https://codebuff.com/docs"
	case "codemachine":
		return "https://codemachine.dev/docs"
	case "factory-droid":
		return "https://factory.ai/docs"
	default:
		return "#"
	}
}

func harnessInstallHint(id string) string {
	switch id {
	case "borg":
		return "Use borg's tracked `submodules/borg` checkout or install borg and ensure `borg` is on PATH."
	case "aider":
		return "pip install aider-chat"
	case "antigravity":
		return "Download the Antigravity desktop app from https://antigravity.google/download and launch it directly; borg does not currently detect it as a PATH CLI."
	case "claude":
		return "npm install -g @anthropic-ai/claude-code"
	case "codex":
		return "Install the Codex CLI binary and make sure `codex` is on PATH."
	case "gemini":
		return "Install the Gemini CLI and ensure `gemini` is on PATH."
	case "opencode":
		return "Install OpenCode and ensure `opencode` is on PATH."
	case "cursor":
		return "Install Cursor and enable its shell command so `cursor` is available on PATH."
	case "copilot":
		return "Install GitHub Copilot CLI and ensure `github-copilot-cli` or `gh copilot` is available."
	case "goose":
		return "Install Goose and ensure `goose` is on PATH."
	case "qwen":
		return "Install the Qwen CLI and ensure `qwen` is on PATH."
	case "superai-cli":
		return "npm install -g superai-cli"
	case "codebuff":
		return "npm install -g codebuff"
	case "codemachine":
		return "npm install -g codemachine"
	case "factory-droid":
		return "Install Factory Droid and ensure `droid` is on PATH."
	default:
		return "Installation instructions unavailable."
	}
}

func harnessCategory(id string) string {
	switch id {
	case "cursor", "antigravity":
		return "editor"
	default:
		return "cli"
	}
}

func harnessSessionCapable(id string) bool {
	switch id {
	case "borg", "aider", "claude", "codex", "gemini", "opencode", "superai-cli", "codebuff", "codemachine", "factory-droid":
		return true
	default:
		return false
	}
}

func toolToHarnessID(toolType string) string {
	switch toolType {
	case "claude-code":
		return "claude"
	case "factory-droid":
		return "factory-droid"
	case "copilot":
		return "copilot"
	default:
		return toolType
	}
}

func localExecutionShells() []map[string]any {
	isWindows := runtime.GOOS == "windows"
	gitBashPath := filepath.Join("C:\\", "Program Files", "Git", "bin", "bash.exe")
	shells := []map[string]any{
		{
			"id":           "pwsh",
			"name":         "PowerShell 7",
			"family":       "powershell",
			"installed":    lookupPathExists("pwsh"),
			"verified":     lookupPathExists("pwsh"),
			"resolvedPath": lookupPath("pwsh"),
			"version":      nil,
			"preferred":    false,
			"notes":        []string{"Preferred borg shell on Windows for general command execution and structured scripting."},
		},
		{
			"id":           "powershell",
			"name":         "Windows PowerShell",
			"family":       "powershell",
			"installed":    lookupPathExists("powershell"),
			"verified":     lookupPathExists("powershell"),
			"resolvedPath": lookupPath("powershell"),
			"version":      nil,
			"preferred":    false,
			"notes":        []string{"Useful legacy fallback when PowerShell 7 is unavailable."},
		},
		{
			"id":           "cmd",
			"name":         "Command Prompt",
			"family":       "cmd",
			"installed":    isWindows,
			"verified":     isWindows,
			"resolvedPath": lookupPath("cmd"),
			"version":      nil,
			"preferred":    false,
			"notes":        []string{"Lowest-common-denominator Windows shell for compatibility-only flows."},
		},
		{
			"id":           "git-bash",
			"name":         "Git Bash",
			"family":       "posix",
			"installed":    fileExists(gitBashPath),
			"verified":     fileExists(gitBashPath),
			"resolvedPath": nullableExistingPath(gitBashPath),
			"version":      nil,
			"preferred":    false,
			"notes":        []string{"POSIX-friendly shell for lightweight Unix tooling without a full Cygwin install."},
		},
		{
			"id":           "wsl",
			"name":         "Windows Subsystem for Linux",
			"family":       "wsl",
			"installed":    lookupPathExists("wsl"),
			"verified":     lookupPathExists("wsl"),
			"resolvedPath": lookupPath("wsl"),
			"version":      nil,
			"preferred":    false,
			"notes":        []string{"Best fit for Linux-native commands when WSL is installed and configured."},
		},
	}
	return shells
}

func selectPreferredShellID(shells []map[string]any) string {
	order := []string{"pwsh", "powershell", "cmd", "git-bash", "wsl"}
	for _, shellID := range order {
		for _, shell := range shells {
			id, _ := shell["id"].(string)
			verified, _ := shell["verified"].(bool)
			if id == shellID && verified {
				return shellID
			}
		}
	}
	return ""
}

func installedShellCount(shells []map[string]any) int {
	count := 0
	for _, shell := range shells {
		installed, _ := shell["installed"].(bool)
		if installed {
			count++
		}
	}
	return count
}

func toolNotes(tool controlplane.Tool) []string {
	if tool.Available && strings.TrimSpace(tool.Version) != "" {
		return []string{}
	}
	if tool.Available {
		return []string{"Binary resolved on PATH, but version verification failed."}
	}
	return []string{"Binary not detected on PATH."}
}

func firstExistingRelativePath(workspaceRoot string, candidates []string) string {
	for _, candidate := range candidates {
		if fileExists(filepath.Join(workspaceRoot, candidate)) {
			return candidate
		}
	}
	return ""
}

func installSurfaceArtifact(id string, ready bool, artifactPath string, artifactKind string, detail map[bool]string, declaredVersion string, workspaceRoot string) map[string]any {
	status := "missing"
	if ready {
		status = "ready"
	}
	return map[string]any{
		"id":              id,
		"status":          status,
		"artifactPath":    emptyStringToNilAny(artifactPath),
		"artifactKind":    emptyStringToNilAny(artifactKind),
		"detail":          detail[ready],
		"declaredVersion": emptyStringToNilAny(declaredVersion),
		"lastModifiedAt":  lastModifiedAtIfPresent(workspaceRoot, artifactPath),
	}
}

func firefoxInstallSurface(workspaceRoot string, bundlePath string, manifestPath string) map[string]any {
	status := "missing"
	artifactPath := manifestPath
	artifactKind := ""
	detail := "No Firefox-ready browser extension artifact was detected yet."
	if strings.TrimSpace(bundlePath) != "" {
		status = "ready"
		artifactPath = bundlePath
		artifactKind = "Firefox unpacked bundle"
		detail = "Firefox-specific browser extension output is available."
	} else if strings.TrimSpace(manifestPath) != "" {
		status = "partial"
		artifactKind = "Firefox manifest source"
		detail = "Firefox manifest source is present, but no packaged Firefox bundle was detected yet."
	}
	return map[string]any{
		"id":              "browser-extension-firefox",
		"status":          status,
		"artifactPath":    emptyStringToNilAny(artifactPath),
		"artifactKind":    emptyStringToNilAny(artifactKind),
		"detail":          detail,
		"declaredVersion": emptyStringToNilAny(packageVersion(workspaceRoot, filepath.Join("apps", "borg-extension", "package.json"))),
		"lastModifiedAt":  lastModifiedAtIfPresent(workspaceRoot, artifactPath),
	}
}

func vscodeInstallSurface(workspaceRoot string, buildPath string) map[string]any {
	status := "missing"
	artifactKind := ""
	detail := "Build and package the VS Code extension to generate an installable `.vsix`."
	if strings.TrimSpace(buildPath) != "" {
		status = "partial"
		artifactKind = "Compiled extension output"
		detail = "VS Code extension is compiled, but no `.vsix` package was detected yet."
	}
	return map[string]any{
		"id":              "vscode-extension",
		"status":          status,
		"artifactPath":    emptyStringToNilAny(buildPath),
		"artifactKind":    emptyStringToNilAny(artifactKind),
		"detail":          detail,
		"declaredVersion": emptyStringToNilAny(packageVersion(workspaceRoot, filepath.Join("packages", "vscode", "package.json"))),
		"lastModifiedAt":  lastModifiedAtIfPresent(workspaceRoot, buildPath),
	}
}

func chromiumArtifactKind(path string) string {
	switch path {
	case filepath.Join("apps", "borg-extension", "dist-chromium"):
		return "Chromium unpacked bundle"
	case filepath.Join("apps", "extension", "dist"):
		return "Legacy extension dist bundle"
	case filepath.Join("apps", "borg-extension", "dist"):
		return "Generic borg-extension dist bundle"
	default:
		return ""
	}
}

func mcpConfigArtifactKind(path string) string {
	switch path {
	case "mcp.jsonc":
		return "JSONC config source"
	case "mcp.json":
		return "JSON config source"
	default:
		return ""
	}
}

func packageVersion(workspaceRoot string, relativePath string) string {
	raw, err := os.ReadFile(filepath.Join(workspaceRoot, relativePath))
	if err != nil {
		return ""
	}
	var parsed struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return ""
	}
	return parsed.Version
}

func lastModifiedAtIfPresent(workspaceRoot string, relativePath string) any {
	if strings.TrimSpace(relativePath) == "" {
		return nil
	}
	info, err := os.Stat(filepath.Join(workspaceRoot, relativePath))
	if err != nil {
		return nil
	}
	return info.ModTime().UTC().Format(time.RFC3339)
}

func lookupPath(command string) any {
	path, err := exec.LookPath(command)
	if err != nil {
		return nil
	}
	return path
}

func lookupPathExists(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

func nullableExistingPath(path string) any {
	if !fileExists(path) {
		return nil
	}
	return path
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func fallbackSearchMCPTools(definitions []harnesses.Definition, query string) []map[string]any {
	allTools := fallbackMCPTools(definitions)
	terms := strings.Fields(strings.ToLower(query))
	if len(terms) == 0 {
		if len(allTools) > 8 {
			return allTools[:8]
		}
		return allTools
	}

	type rankedTool struct {
		tool  map[string]any
		score int
	}
	ranked := make([]rankedTool, 0, len(allTools))
	for _, tool := range allTools {
		name, _ := tool["name"].(string)
		server, _ := tool["server"].(string)
		haystack := strings.ToLower(name + " " + server)
		score := 0
		for _, term := range terms {
			if term == "" {
				continue
			}
			if strings.Contains(haystack, term) {
				score++
			}
		}
		if score == 0 {
			continue
		}
		ranked = append(ranked, rankedTool{
			tool:  tool,
			score: score,
		})
	}

	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].score == ranked[j].score {
			leftName, _ := ranked[i].tool["name"].(string)
			rightName, _ := ranked[j].tool["name"].(string)
			return leftName < rightName
		}
		return ranked[i].score > ranked[j].score
	})

	results := make([]map[string]any, 0, len(ranked))
	for _, item := range ranked {
		name, _ := item.tool["name"].(string)
		server, _ := item.tool["server"].(string)
		results = append(results, map[string]any{
			"name":        name,
			"server":      server,
			"alwaysShow":  false,
			"matchReason": "Matched local source-backed tool inventory.",
			"score":       item.score,
		})
	}
	return results
}

func fallbackControlToolSearch(definitions []harnesses.Definition, query string, limit int) []map[string]any {
	allTools := fallbackControlTools(definitions)
	terms := strings.Fields(strings.ToLower(query))
	if len(terms) == 0 {
		if limit > 0 && len(allTools) > limit {
			return allTools[:limit]
		}
		return allTools
	}

	type rankedTool struct {
		tool  map[string]any
		score int
	}
	ranked := make([]rankedTool, 0, len(allTools))
	for _, tool := range allTools {
		name, _ := tool["name"].(string)
		server, _ := tool["server"].(string)
		description, _ := tool["description"].(string)
		haystack := strings.ToLower(name + " " + server + " " + description)
		score := 0
		for _, term := range terms {
			if term == "" {
				continue
			}
			if strings.Contains(haystack, term) {
				score++
			}
		}
		if score == 0 {
			continue
		}
		ranked = append(ranked, rankedTool{tool: tool, score: score})
	}

	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].score == ranked[j].score {
			leftName, _ := ranked[i].tool["name"].(string)
			rightName, _ := ranked[j].tool["name"].(string)
			return leftName < rightName
		}
		return ranked[i].score > ranked[j].score
	})

	results := make([]map[string]any, 0, len(ranked))
	for _, item := range ranked {
		results = append(results, item.tool)
		if limit > 0 && len(results) >= limit {
			break
		}
	}
	return results
}

func (s *Server) localCallMCPMetaTool(r *http.Request, payload map[string]any) (map[string]any, error) {
	name, _ := payload["name"].(string)
	args, _ := payload["args"].(map[string]any)
	if args == nil {
		args = map[string]any{}
	}

	switch name {
	case "search_tools":
		query, _ := args["query"].(string)
		_, summary, err := s.localMCPSummary(r.Context())
		if err != nil {
			return nil, err
		}
		results := fallbackSearchMCPTools(summary.InstalledHarnesses, query)
		return map[string]any{
			"ok": true,
			"result": map[string]any{
				"content": []map[string]any{
					{
						"type": "text",
						"text": prettyJSON(results),
					},
				},
			},
		}, nil
	case "list_all_tools":
		_, summary, err := s.localMCPSummary(r.Context())
		if err != nil {
			return nil, err
		}
		results := fallbackMCPTools(summary.InstalledHarnesses)
		return map[string]any{
			"ok": true,
			"result": map[string]any{
				"content": []map[string]any{
					{
						"type": "text",
						"text": prettyJSON(results),
					},
				},
			},
		}, nil
	case "auto_call_tool":
		normalized := normalizeAutoCallArgs(args)
		objective, _ := normalized["objective"].(string)
		if strings.TrimSpace(objective) == "" {
			return map[string]any{
				"ok": false,
				"result": map[string]any{
					"content": []map[string]any{
						{
							"type": "text",
							"text": "Objective is required for auto_call_tool.",
						},
					},
				},
			}, nil
		}
		_, summary, err := s.localMCPSummary(r.Context())
		if err != nil {
			return nil, err
		}
		searchResults := fallbackSearchMCPTools(summary.InstalledHarnesses, objective)
		chosen := "list_all_tools"
		if len(searchResults) > 0 {
			if toolName, ok := searchResults[0]["name"].(string); ok && strings.TrimSpace(toolName) != "" {
				chosen = toolName
			}
		}
		return map[string]any{
			"ok": true,
			"result": map[string]any{
				"content": []map[string]any{
					{
						"type": "text",
						"text": "[Auto-Execution Logic: Chose " + chosen + "]\n--- Result ---\nLocal fallback can recommend tools but cannot execute non-meta MCP tools without the TypeScript bridge.",
					},
				},
			},
		}, nil
	default:
		return nil, errors.New("unsupported tool fallback: " + name)
	}
}

func localFallbackToolSchema(payload map[string]any) (map[string]any, error) {
	name, _ := payload["name"].(string)
	switch name {
	case "search_tools":
		return map[string]any{
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "Tool intent or keyword search query.",
					},
					"limit": map[string]any{
						"type":        "number",
						"description": "Maximum number of results to return (default 10).",
					},
				},
				"required": []string{"query"},
			},
			"evictedHydratedTools": []any{},
		}, nil
	case "list_all_tools":
		return map[string]any{
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "Optional keyword filter applied across tool names, descriptions, server names, and advertised names.",
					},
					"limit": map[string]any{
						"type":        "number",
						"description": "Maximum number of tools to return after filtering. Defaults to 100.",
					},
					"category": map[string]any{
						"type":        "string",
						"description": "Optional category filter.",
						"enum":        []string{"all", "meta", "compatibility", "native", "saved-script", "downstream"},
					},
				},
			},
			"evictedHydratedTools": []any{},
		}, nil
	case "auto_call_tool":
		return map[string]any{
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"objective": map[string]any{
						"type":        "string",
						"description": "The objective or task you want to accomplish using a tool.",
					},
					"context": map[string]any{
						"type":        "string",
						"description": "Any necessary variables, file paths, or text snippets required to fill the tool arguments.",
					},
				},
				"required": []string{"objective", "context"},
			},
			"evictedHydratedTools": []any{},
		}, nil
	default:
		return nil, errors.New("unsupported tool schema fallback: " + name)
	}
}

func (s *Server) localMCPRegistrySnapshot() ([]map[string]any, error) {
	indexPath := filepath.Join(s.cfg.WorkspaceRoot, "BORG_MASTER_INDEX.jsonc")
	content, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, err
	}

	sanitized := stripJSONCLineComments(string(content))
	var parsed struct {
		Categories map[string][]map[string]any `json:"categories"`
	}
	if err := json.Unmarshal([]byte(sanitized), &parsed); err != nil {
		return nil, err
	}

	results := make([]map[string]any, 0)
	seenByURL := make(map[string]struct{})
	for category, items := range parsed.Categories {
		for _, item := range items {
			url, _ := item["url"].(string)
			if strings.TrimSpace(url) == "" {
				continue
			}
			if _, seen := seenByURL[url]; seen {
				continue
			}
			if !isMCPLikeRegistryEntry(category, item) {
				continue
			}
			seenByURL[url] = struct{}{}

			name, _ := item["name"].(string)
			if strings.TrimSpace(name) == "" {
				name, _ = item["id"].(string)
			}
			if strings.TrimSpace(name) == "" {
				name = url
			}
			description, _ := item["description"].(string)
			if strings.TrimSpace(description) == "" {
				description, _ = item["summary"].(string)
			}
			if strings.TrimSpace(description) == "" {
				description = "No description available."
			}

			tags := make([]string, 0)
			if rawTags, ok := item["tags"].([]any); ok {
				for _, rawTag := range rawTags {
					tag, _ := rawTag.(string)
					if strings.TrimSpace(tag) == "" {
						continue
					}
					tags = append(tags, strings.ToLower(tag))
				}
			}
			id, _ := item["id"].(string)
			if strings.TrimSpace(id) == "" {
				id = url
			}
			results = append(results, map[string]any{
				"id":          id,
				"name":        name,
				"url":         url,
				"category":    category,
				"description": description,
				"tags":        tags,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		leftName, _ := results[i]["name"].(string)
		rightName, _ := results[j]["name"].(string)
		return leftName < rightName
	})
	if len(results) > 300 {
		results = results[:300]
	}
	return results, nil
}

func isMCPLikeRegistryEntry(category string, item map[string]any) bool {
	url, _ := item["url"].(string)
	kind, _ := item["kind"].(string)
	if strings.Contains(strings.ToLower(category), "mcp") || strings.Contains(strings.ToLower(kind), "mcp") || strings.Contains(strings.ToLower(url), "modelcontextprotocol") || strings.Contains(strings.ToLower(url), "mcp") {
		return true
	}
	if rawTags, ok := item["tags"].([]any); ok {
		for _, rawTag := range rawTags {
			tag, _ := rawTag.(string)
			if strings.Contains(strings.ToLower(tag), "mcp") {
				return true
			}
		}
	}
	return false
}

func (s *Server) localMCPJsoncEditor() (map[string]any, error) {
	jsoncPath := filepath.Join(s.cfg.MainConfigDir, "mcp.jsonc")
	content, err := os.ReadFile(jsoncPath)
	if err == nil {
		return map[string]any{
			"path":    jsoncPath,
			"content": string(content),
		}, nil
	}
	if !os.IsNotExist(err) {
		return nil, err
	}

	fallback := map[string]any{
		"mcpServers": map[string]any{},
	}
	return map[string]any{
		"path":    jsoncPath,
		"content": "// borg MCP configuration\n" + prettyJSON(fallback) + "\n",
	}, nil
}

func (s *Server) saveLocalMCPJsonc(content string) error {
	sanitized := stripJSONCLineComments(content)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(sanitized), &parsed); err != nil {
		return err
	}
	if _, ok := parsed["mcpServers"]; !ok {
		parsed["mcpServers"] = map[string]any{}
	}

	jsoncPath := filepath.Join(s.cfg.MainConfigDir, "mcp.jsonc")
	jsonPath := filepath.Join(s.cfg.MainConfigDir, "mcp.json")
	if err := os.MkdirAll(s.cfg.MainConfigDir, 0o755); err != nil {
		return err
	}

	jsoncBody := "// borg MCP configuration\n// This file is borg-owned and may include cached server metadata under mcpServers.<name>._meta.\n" + prettyJSON(parsed) + "\n"
	if err := os.WriteFile(jsoncPath, []byte(jsoncBody), 0o644); err != nil {
		return err
	}

	compatibility := make(map[string]any, len(parsed))
	for key, value := range parsed {
		if key == "settings" {
			continue
		}
		if key == "mcpServers" {
			compatibility[key] = stripServerMeta(value)
			continue
		}
		compatibility[key] = value
	}
	return os.WriteFile(jsonPath, []byte(prettyJSON(compatibility)+"\n"), 0o644)
}

func (s *Server) localMemoryContexts() ([]map[string]any, error) {
	contextsPath := filepath.Join(s.cfg.WorkspaceRoot, ".borg", "memory", "contexts.json")
	raw, err := os.ReadFile(contextsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []map[string]any{}, nil
		}
		return nil, err
	}

	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" {
		return []map[string]any{}, nil
	}

	var contexts []map[string]any
	if err := json.Unmarshal(raw, &contexts); err != nil {
		return nil, err
	}
	return contexts, nil
}

type localAgentMemorySnapshot struct {
	Memories []localAgentMemoryRecord `json:"memories"`
}

type localAgentMemoryRecord struct {
	ID          string         `json:"id"`
	Content     string         `json:"content"`
	Type        string         `json:"type"`
	Namespace   string         `json:"namespace"`
	Metadata    map[string]any `json:"metadata"`
	CreatedAt   string         `json:"createdAt"`
	AccessedAt  string         `json:"accessedAt"`
	AccessCount int            `json:"accessCount"`
	TTL         *float64       `json:"ttl"`
}

func localAgentMemoryZeroStats() map[string]any {
	return map[string]any{
		"totalCount":             0,
		"sessionCount":           0,
		"workingCount":           0,
		"longTermCount":          0,
		"observationCount":       0,
		"uniqueObservationCount": 0,
		"promptCount":            0,
		"sessionSummaryCount":    0,
		"session":                0,
		"working":                0,
		"long_term":              0,
		"user":                   0,
		"agent":                  0,
		"project":                0,
		"discovery":              0,
		"decision":               0,
		"progress":               0,
		"warning":                0,
		"fix":                    0,
	}
}

func (s *Server) localAgentMemoryStats() (map[string]any, error) {
	records, err := s.localAgentMemories()
	if err != nil {
		return nil, err
	}
	stats := localAgentMemoryZeroStats()
	observationHashes := map[string]struct{}{}

	for _, memory := range records {
		stats["totalCount"] = stats["totalCount"].(int) + 1

		switch memory.Type {
		case "session":
			stats["sessionCount"] = stats["sessionCount"].(int) + 1
			stats["session"] = stats["session"].(int) + 1
		case "working":
			stats["workingCount"] = stats["workingCount"].(int) + 1
			stats["working"] = stats["working"].(int) + 1
		case "long_term":
			stats["longTermCount"] = stats["longTermCount"].(int) + 1
			stats["long_term"] = stats["long_term"].(int) + 1
		}

		switch memory.Namespace {
		case "user":
			stats["user"] = stats["user"].(int) + 1
		case "agent":
			stats["agent"] = stats["agent"].(int) + 1
		case "project":
			stats["project"] = stats["project"].(int) + 1
		}

		metadata := memory.Metadata
		if metadata == nil {
			continue
		}

		if observation, ok := metadata["structuredObservation"].(map[string]any); ok {
			stats["observationCount"] = stats["observationCount"].(int) + 1
			if observationType, ok := observation["type"].(string); ok {
				switch observationType {
				case "discovery", "decision", "progress", "warning", "fix":
					stats[observationType] = stats[observationType].(int) + 1
				}
			}
			if contentHash, ok := observation["contentHash"].(string); ok && strings.TrimSpace(contentHash) != "" {
				observationHashes[contentHash] = struct{}{}
			}
		}

		if _, ok := metadata["structuredUserPrompt"].(map[string]any); ok {
			stats["promptCount"] = stats["promptCount"].(int) + 1
		}

		if _, ok := metadata["structuredSessionSummary"].(map[string]any); ok {
			stats["sessionSummaryCount"] = stats["sessionSummaryCount"].(int) + 1
		}
	}

	stats["uniqueObservationCount"] = len(observationHashes)
	return stats, nil
}

func (s *Server) localAgentMemories() ([]localAgentMemoryRecord, error) {
	memoriesPath := filepath.Join(s.cfg.WorkspaceRoot, ".borg", "agent_memory", "memories.json")
	raw, err := os.ReadFile(memoriesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []localAgentMemoryRecord{}, nil
		}
		return nil, err
	}

	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" {
		return []localAgentMemoryRecord{}, nil
	}

	var snapshot localAgentMemorySnapshot
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		return nil, err
	}

	now := time.Now()
	records := make([]localAgentMemoryRecord, 0, len(snapshot.Memories))
	for _, memory := range snapshot.Memories {
		if memory.Type == "session" && memory.TTL != nil {
			createdAt, ok := localAgentMemoryTime(memory.CreatedAt)
			if ok && now.Sub(createdAt) > time.Duration(*memory.TTL)*time.Millisecond {
				continue
			}
		}
		if memory.Metadata == nil {
			memory.Metadata = map[string]any{}
		}
		records = append(records, memory)
	}
	return records, nil
}

func (s *Server) localRecentObservations(limit int, namespace, observationType string) ([]map[string]any, error) {
	records, err := s.localAgentMemories()
	if err != nil {
		return nil, err
	}

	filtered := make([]localAgentMemoryRecord, 0)
	for _, record := range records {
		observation, ok := localStructuredObservation(record.Metadata)
		if !ok {
			continue
		}
		if namespace != "" && record.Namespace != namespace {
			continue
		}
		if observationType != "" && stringValue(observation["type"]) != observationType {
			continue
		}
		filtered = append(filtered, record)
	}

	sort.Slice(filtered, func(i, j int) bool {
		return localAgentMemorySortTime(filtered[i]).After(localAgentMemorySortTime(filtered[j]))
	})

	if limit <= 0 || limit > len(filtered) {
		limit = len(filtered)
	}
	results := make([]map[string]any, 0, limit)
	for _, record := range filtered[:limit] {
		results = append(results, localAgentMemoryMap(record))
	}
	return results, nil
}

func (s *Server) localRecentUserPrompts(limit int, role string) ([]map[string]any, error) {
	records, err := s.localAgentMemories()
	if err != nil {
		return nil, err
	}

	filtered := make([]localAgentMemoryRecord, 0)
	for _, record := range records {
		prompt, ok := localStructuredUserPrompt(record.Metadata)
		if !ok {
			continue
		}
		if record.Type != "long_term" || record.Namespace != "project" {
			continue
		}
		if role != "" && stringValue(prompt["role"]) != role {
			continue
		}
		filtered = append(filtered, record)
	}

	sort.Slice(filtered, func(i, j int) bool {
		return localAgentMemorySortTime(filtered[i]).After(localAgentMemorySortTime(filtered[j]))
	})

	if limit <= 0 || limit > len(filtered) {
		limit = len(filtered)
	}
	results := make([]map[string]any, 0, limit)
	for _, record := range filtered[:limit] {
		results = append(results, localAgentMemoryMap(record))
	}
	return results, nil
}

func (s *Server) localRecentSessionSummaries(limit int) ([]map[string]any, error) {
	records, err := s.localAgentMemories()
	if err != nil {
		return nil, err
	}

	filtered := make([]localAgentMemoryRecord, 0)
	for _, record := range records {
		if _, ok := localStructuredSessionSummary(record.Metadata); !ok {
			continue
		}
		if record.Type != "long_term" {
			continue
		}
		filtered = append(filtered, record)
	}

	sort.Slice(filtered, func(i, j int) bool {
		return localAgentMemorySortTime(filtered[i]).After(localAgentMemorySortTime(filtered[j]))
	})

	if limit <= 0 || limit > len(filtered) {
		limit = len(filtered)
	}
	results := make([]map[string]any, 0, limit)
	for _, record := range filtered[:limit] {
		results = append(results, localAgentMemoryMap(record))
	}
	return results, nil
}

func (s *Server) localSessionBootstrapPayload(activeGoal, lastObjective string) (map[string]any, error) {
	summaries, err := s.localRecentSessionSummaries(3)
	if err != nil {
		return nil, err
	}
	observations, err := s.localRecentObservations(5, "", "")
	if err != nil {
		return nil, err
	}

	summaryContents := make([]string, 0, len(summaries))
	for _, summary := range summaries {
		summaryContents = append(summaryContents, stringValue(summary["content"]))
	}

	observationLines := make([]map[string]any, 0, len(observations))
	for _, observation := range observations {
		metadata, _ := observation["metadata"].(map[string]any)
		structured, _ := localStructuredObservation(metadata)
		observationLines = append(observationLines, map[string]any{
			"title":     stringValue(structured["title"]),
			"narrative": stringValue(structured["narrative"]),
			"type":      stringValue(structured["type"]),
			"toolName":  stringValue(structured["toolName"]),
			"content":   stringValue(observation["content"]),
		})
	}

	return map[string]any{
		"activeGoal":             nullableString(activeGoal),
		"lastObjective":          nullableString(lastObjective),
		"goal":                   nullableString(activeGoal),
		"objective":              nullableString(lastObjective),
		"summaryCount":           len(summaryContents),
		"observationCount":       len(observationLines),
		"toolAdvertisementCount": 0,
		"prompt":                 strings.Join(localBootstrapLines(activeGoal, lastObjective, summaryContents, observationLines), "\n"),
	}, nil
}

func (s *Server) localToolContextPayload(toolName, activeGoal, lastObjective string) (map[string]any, error) {
	records, err := s.localAgentMemories()
	if err != nil {
		return nil, err
	}

	observations := make([]map[string]any, 0)
	summaries := make([]map[string]any, 0)
	for _, record := range records {
		if structured, ok := localStructuredObservation(record.Metadata); ok {
			observations = append(observations, map[string]any{
				"title":         stringValue(structured["title"]),
				"narrative":     stringValue(structured["narrative"]),
				"content":       record.Content,
				"type":          stringValue(structured["type"]),
				"toolName":      stringValue(structured["toolName"]),
				"concepts":      stringArray(structured["concepts"]),
				"filesRead":     stringArray(structured["filesRead"]),
				"filesModified": stringArray(structured["filesModified"]),
				"recordedAt":    localNumericValue(structured["recordedAt"]),
			})
		}
		if structured, ok := localStructuredSessionSummary(record.Metadata); ok {
			recordedAt := localNumericValue(structured["stoppedAt"])
			if recordedAt <= 0 {
				recordedAt = localTimeToMillis(localAgentMemorySortTime(record))
			}
			summaries = append(summaries, map[string]any{
				"content":    record.Content,
				"cliType":    stringValue(structured["cliType"]),
				"status":     stringValue(structured["status"]),
				"sessionId":  stringValue(structured["sessionId"]),
				"recordedAt": recordedAt,
			})
		}
	}

	query, tokens := localToolContextQuery(toolName)
	matchedObservations := localMatchedToolObservations(observations, toolName, tokens)
	matchedSummaries := localMatchedToolSummaries(summaries, tokens)
	fallbackSummaries := matchedSummaries
	if len(fallbackSummaries) == 0 && len(summaries) > 0 {
		sorted := append([]map[string]any{}, summaries...)
		sort.Slice(sorted, func(i, j int) bool {
			return localNumericValue(sorted[i]["recordedAt"]) > localNumericValue(sorted[j]["recordedAt"])
		})
		fallbackSummaries = sorted[:1]
	}

	matchedPaths := localUniqueStrings(nil,
		func() []string {
			paths := make([]string, 0)
			for _, observation := range matchedObservations {
				paths = append(paths, stringArray(observation["filesRead"])...)
				paths = append(paths, stringArray(observation["filesModified"])...)
			}
			return paths
		}()...,
	)

	return map[string]any{
		"toolName":         localTrimLine(toolName, 120),
		"query":            query,
		"matchedPaths":     matchedPaths,
		"observationCount": len(matchedObservations),
		"summaryCount":     len(fallbackSummaries),
		"prompt":           strings.Join(localToolContextLines(toolName, query, activeGoal, lastObjective, matchedPaths, matchedObservations, fallbackSummaries), "\n"),
	}, nil
}

func localAgentMemoryMap(record localAgentMemoryRecord) map[string]any {
	data := map[string]any{
		"id":          record.ID,
		"content":     record.Content,
		"type":        record.Type,
		"namespace":   record.Namespace,
		"metadata":    cloneMap(record.Metadata),
		"createdAt":   record.CreatedAt,
		"accessedAt":  nullableString(record.AccessedAt),
		"accessCount": record.AccessCount,
	}
	if record.TTL != nil {
		data["ttl"] = *record.TTL
	}
	return data
}

func localStructuredObservation(metadata map[string]any) (map[string]any, bool) {
	value, ok := metadata["structuredObservation"].(map[string]any)
	if !ok || strings.TrimSpace(stringValue(value["title"])) == "" || strings.TrimSpace(stringValue(value["contentHash"])) == "" {
		return nil, false
	}
	observationType := stringValue(value["type"])
	switch observationType {
	case "discovery", "decision", "progress", "warning", "fix":
	default:
		return nil, false
	}
	return value, true
}

func localStructuredUserPrompt(metadata map[string]any) (map[string]any, bool) {
	value, ok := metadata["structuredUserPrompt"].(map[string]any)
	if !ok || strings.TrimSpace(stringValue(value["content"])) == "" || localNumericValue(value["recordedAt"]) <= 0 {
		return nil, false
	}
	return value, true
}

func localStructuredSessionSummary(metadata map[string]any) (map[string]any, bool) {
	value, ok := metadata["structuredSessionSummary"].(map[string]any)
	if !ok || strings.TrimSpace(stringValue(value["sessionId"])) == "" || strings.TrimSpace(stringValue(value["status"])) == "" {
		return nil, false
	}
	return value, true
}

func localAgentMemoryTime(value string) (time.Time, bool) {
	if strings.TrimSpace(value) == "" {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err == nil {
		return parsed, true
	}
	parsed, err = time.Parse(time.RFC3339, value)
	if err == nil {
		return parsed, true
	}
	return time.Time{}, false
}

func localAgentMemorySortTime(record localAgentMemoryRecord) time.Time {
	if parsed, ok := localAgentMemoryTime(record.CreatedAt); ok {
		return parsed
	}
	return time.Time{}
}

func localBootstrapLines(activeGoal, lastObjective string, summaries []string, observations []map[string]any) []string {
	lines := []string{"Memory bootstrap:"}
	if strings.TrimSpace(activeGoal) != "" {
		lines = append(lines, "Current goal: "+localTrimLine(activeGoal, 180))
	}
	if strings.TrimSpace(lastObjective) != "" {
		lines = append(lines, "Last objective: "+localTrimLine(lastObjective, 180))
	}

	summaryLines := make([]string, 0, minInt(len(summaries), 3))
	for _, summary := range summaries {
		summaryLines = append(summaryLines, "- "+localTrimLine(summary, 240))
		if len(summaryLines) >= 3 {
			break
		}
	}
	observationLines := make([]string, 0, minInt(len(observations), 5))
	for _, observation := range observations {
		title := localTrimLine(stringValue(firstNonEmptyString(
			stringValue(observation["title"]),
			stringValue(observation["narrative"]),
			stringValue(observation["content"]),
			"Recent observation available",
		)), 220)
		tags := strings.Join(localUniqueStrings(nil, localTrimLine(stringValue(observation["type"]), 40), localTrimLine(stringValue(observation["toolName"]), 60)), " · ")
		if tags != "" {
			title += " (" + tags + ")"
		}
		observationLines = append(observationLines, "- "+title)
		if len(observationLines) >= 5 {
			break
		}
	}

	if len(summaryLines) > 0 {
		lines = append(lines, "Recent session summaries:")
		lines = append(lines, summaryLines...)
	}
	if len(observationLines) > 0 {
		lines = append(lines, "Relevant observations:")
		lines = append(lines, observationLines...)
	}
	return lines
}

func localToolContextLines(toolName, query, activeGoal, lastObjective string, matchedPaths []string, observations []map[string]any, summaries []map[string]any) []string {
	lines := []string{"JIT tool context for " + localTrimLine(toolName, 120) + ":"}
	if strings.TrimSpace(activeGoal) != "" {
		lines = append(lines, "Current goal: "+localTrimLine(activeGoal, 180))
	}
	if strings.TrimSpace(lastObjective) != "" {
		lines = append(lines, "Last objective: "+localTrimLine(lastObjective, 180))
	}
	if strings.TrimSpace(query) != "" {
		lines = append(lines, "Focus query: "+localTrimLine(query, 240))
	}
	if len(matchedPaths) > 0 {
		names := make([]string, 0, len(matchedPaths))
		for _, matchedPath := range matchedPaths {
			names = append(names, localBasenameLike(matchedPath))
		}
		lines = append(lines, "Relevant files: "+strings.Join(localUniqueStrings(nil, names...), ", "))
	}
	if len(observations) > 0 {
		lines = append(lines, "Potentially relevant observations:")
		for _, observation := range observations {
			title := localTrimLine(stringValue(firstNonEmptyString(stringValue(observation["title"]), stringValue(observation["narrative"]), stringValue(observation["content"]), "Relevant observation")), 160)
			detail := localTrimLine(stringValue(firstNonEmptyString(stringValue(observation["narrative"]), stringValue(observation["content"]))), 220)
			tags := strings.Join(localUniqueStrings(nil, stringValue(observation["type"]), stringValue(observation["toolName"])), " · ")
			line := title
			if detail != "" && detail != title {
				line += " — " + detail
			}
			if tags != "" {
				line += " (" + tags + ")"
			}
			lines = append(lines, "- "+line)
		}
	}
	if len(summaries) > 0 {
		lines = append(lines, "Potentially relevant session summaries:")
		for _, summary := range summaries {
			lines = append(lines, "- "+localTrimLine(stringValue(firstNonEmptyString(stringValue(summary["content"]), "Relevant prior session summary")), 240))
		}
	}
	if len(observations) == 0 && len(summaries) == 0 {
		lines = append(lines, "No strongly relevant prior memory was found for this tool call.")
	} else {
		lines = append(lines, "Use only the parts that still match the current intent; ignore stale context.")
	}
	return lines
}

func localToolContextQuery(toolName string) (string, []string) {
	query := localTrimLine(toolName, 240)
	return query, localTokenize(query)
}

func localMatchedToolObservations(observations []map[string]any, toolName string, tokens []string) []map[string]any {
	type scoredObservation struct {
		observation map[string]any
		score       int
		recordedAt  float64
	}

	scored := make([]scoredObservation, 0)
	for _, observation := range observations {
		score := 0
		if stringValue(observation["toolName"]) == toolName {
			score += 8
		} else if strings.Contains(strings.ToLower(stringValue(observation["toolName"])), strings.ToLower(toolName)) && strings.TrimSpace(toolName) != "" {
			score += 4
		}
		score += localScoreText(strings.Join(append(localUniqueStrings(nil,
			stringValue(observation["title"]),
			stringValue(observation["narrative"]),
			stringValue(observation["content"]),
		), stringArray(observation["concepts"])...), " "), tokens)
		if score > 0 {
			scored = append(scored, scoredObservation{
				observation: observation,
				score:       score,
				recordedAt:  localNumericValue(observation["recordedAt"]),
			})
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		if scored[i].score == scored[j].score {
			return scored[i].recordedAt > scored[j].recordedAt
		}
		return scored[i].score > scored[j].score
	})

	limit := minInt(len(scored), 4)
	results := make([]map[string]any, 0, limit)
	for _, item := range scored[:limit] {
		results = append(results, item.observation)
	}
	return results
}

func localMatchedToolSummaries(summaries []map[string]any, tokens []string) []map[string]any {
	type scoredSummary struct {
		summary    map[string]any
		score      int
		recordedAt float64
	}

	scored := make([]scoredSummary, 0)
	for _, summary := range summaries {
		score := localScoreText(strings.Join(localUniqueStrings(nil,
			stringValue(summary["content"]),
			stringValue(summary["cliType"]),
			stringValue(summary["status"]),
		), " "), tokens)
		if score > 0 {
			scored = append(scored, scoredSummary{
				summary:    summary,
				score:      score,
				recordedAt: localNumericValue(summary["recordedAt"]),
			})
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		if scored[i].score == scored[j].score {
			return scored[i].recordedAt > scored[j].recordedAt
		}
		return scored[i].score > scored[j].score
	})

	limit := minInt(len(scored), 2)
	results := make([]map[string]any, 0, limit)
	for _, item := range scored[:limit] {
		results = append(results, item.summary)
	}
	return results
}

func localScoreText(text string, tokens []string) int {
	lowered := strings.ToLower(text)
	score := 0
	for _, token := range tokens {
		if token != "" && strings.Contains(lowered, token) {
			score += 2
		}
	}
	return score
}

func localTokenize(value string) []string {
	parts := strings.FieldsFunc(strings.ToLower(value), func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9')
	})
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		switch part {
		case "", "tool", "with", "from", "that", "this", "into", "using", "args":
			continue
		}
		if len(part) >= 3 {
			filtered = append(filtered, part)
		}
	}
	return localUniqueStrings(nil, filtered...)
}

func localUniqueStrings(seed []string, values ...string) []string {
	seen := make(map[string]struct{})
	normalized := make([]string, 0, len(seed)+len(values))
	for _, value := range seed {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, trimmed)
	}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, trimmed)
	}
	return normalized
}

func localTrimLine(value string, maxLength int) string {
	normalized := strings.Join(strings.Fields(value), " ")
	if maxLength > 0 && len(normalized) > maxLength {
		return normalized[:maxLength]
	}
	return normalized
}

func localBasenameLike(filePath string) string {
	normalized := strings.ReplaceAll(strings.TrimSpace(filePath), "\\", "/")
	if normalized == "" {
		return ""
	}
	parts := strings.Split(normalized, "/")
	return parts[len(parts)-1]
}

func localNumericValue(value any) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	case int32:
		return float64(typed)
	case json.Number:
		parsed, _ := typed.Float64()
		return parsed
	default:
		return 0
	}
}

func localTimeToMillis(value time.Time) float64 {
	if value.IsZero() {
		return 0
	}
	return float64(value.UnixMilli())
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func (s *Server) localMemoryExport(userID, format string) (string, error) {
	if strings.TrimSpace(userID) == "" {
		userID = "default"
	}

	if format == "" {
		format = "json"
	}

	if format == "json-provider" {
		snapshot, err := s.tryReadLocalMemorySnapshot()
		if err != nil {
			return "", err
		}
		if snapshot != "" {
			return snapshot, nil
		}
	}

	memories, err := s.localMemoryExportRecords(userID)
	if err != nil {
		return "", err
	}

	switch format {
	case "json", "json-provider":
		data, err := json.MarshalIndent(memories, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data), nil
	case "jsonl":
		lines := make([]string, 0, len(memories))
		for _, memory := range memories {
			data, err := json.Marshal(memory)
			if err != nil {
				return "", err
			}
			lines = append(lines, string(data))
		}
		return strings.Join(lines, "\n"), nil
	case "csv":
		return serializeLocalMemoryCSV(memories), nil
	case "sectioned-memory-store":
		data, err := json.MarshalIndent(localMemorySectionedStore(memories), "", "  ")
		if err != nil {
			return "", err
		}
		return string(data), nil
	default:
		return "", errors.New("unsupported memory export format: " + format)
	}
}

func (s *Server) tryReadLocalMemorySnapshot() (string, error) {
	path := filepath.Join(s.cfg.WorkspaceRoot, "memory.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(raw)), nil
}

func (s *Server) localMemoryExportRecords(userID string) ([]map[string]any, error) {
	snapshot, err := s.tryReadLocalMemorySnapshot()
	if err != nil {
		return nil, err
	}
	if snapshot != "" {
		var memories []map[string]any
		if err := json.Unmarshal([]byte(snapshot), &memories); err == nil {
			return ensureLocalMemoryExportFields(memories, userID), nil
		}
	}

	contexts, err := s.localMemoryContexts()
	if err != nil {
		return nil, err
	}

	memories := make([]map[string]any, 0, len(contexts))
	for index, context := range contexts {
		uuid := stringValue(context["uuid"])
		if uuid == "" {
			uuid = stringValue(context["id"])
		}
		if uuid == "" {
			uuid = "context-" + strconv.Itoa(index+1)
		}

		createdAt := time.Now().UTC().Format(time.RFC3339)
		if raw := context["createdAt"]; raw != nil {
			switch value := raw.(type) {
			case string:
				if strings.TrimSpace(value) != "" {
					createdAt = value
				}
			case float64:
				createdAt = time.UnixMilli(int64(value)).UTC().Format(time.RFC3339)
			case int64:
				createdAt = time.UnixMilli(value).UTC().Format(time.RFC3339)
			case int:
				createdAt = time.UnixMilli(int64(value)).UTC().Format(time.RFC3339)
			}
		}

		metadata, _ := context["metadata"].(map[string]any)
		if metadata == nil {
			metadata = map[string]any{}
		}

		memories = append(memories, map[string]any{
			"uuid":      uuid,
			"content":   stringValue(context["content"]),
			"metadata":  metadata,
			"userId":    userID,
			"createdAt": createdAt,
		})
	}

	return memories, nil
}

func ensureLocalMemoryExportFields(memories []map[string]any, userID string) []map[string]any {
	normalized := make([]map[string]any, 0, len(memories))
	for index, memory := range memories {
		item := cloneMap(memory)
		uuid := stringValue(item["uuid"])
		if uuid == "" {
			uuid = stringValue(item["id"])
		}
		if uuid == "" {
			uuid = "memory-" + strconv.Itoa(index+1)
		}
		item["uuid"] = uuid

		if stringValue(item["userId"]) == "" {
			item["userId"] = userID
		}

		if _, ok := item["metadata"].(map[string]any); !ok {
			item["metadata"] = map[string]any{}
		}

		if stringValue(item["createdAt"]) == "" {
			item["createdAt"] = time.Now().UTC().Format(time.RFC3339)
		}

		if _, ok := item["content"]; !ok {
			item["content"] = ""
		}

		normalized = append(normalized, item)
	}
	return normalized
}

func serializeLocalMemoryCSV(memories []map[string]any) string {
	rows := []string{"uuid,content,userId,agentId,createdAt,metadata"}
	for _, memory := range memories {
		metadataJSON, _ := json.Marshal(memory["metadata"])
		rows = append(rows, strings.Join([]string{
			stringValue(memory["uuid"]),
			csvQuote(stringValue(memory["content"])),
			stringValue(memory["userId"]),
			stringValue(memory["agentId"]),
			stringValue(memory["createdAt"]),
			csvQuote(string(metadataJSON)),
		}, ","))
	}
	return strings.Join(rows, "\n")
}

func localMemorySectionedStore(memories []map[string]any) map[string]any {
	defaultSections := []string{"project_context", "user_facts", "style_preferences", "commands", "general"}
	sections := map[string][]map[string]any{}
	for _, section := range defaultSections {
		sections[section] = []map[string]any{}
	}

	for _, memory := range memories {
		metadata, _ := memory["metadata"].(map[string]any)
		sectionName := "general"
		if metadata != nil && stringValue(metadata["section"]) != "" {
			sectionName = stringValue(metadata["section"])
		}
		entry := map[string]any{
			"uuid":      stringValue(memory["uuid"]),
			"content":   stringValue(memory["content"]),
			"tags":      stringArray(metadata["tags"]),
			"createdAt": stringValue(memory["createdAt"]),
			"source":    localMemorySource(metadata, stringValue(memory["agentId"])),
		}
		sections[sectionName] = append(sections[sectionName], entry)
	}

	ordered := make([]map[string]any, 0, len(sections))
	seen := map[string]bool{}
	for _, section := range defaultSections {
		ordered = append(ordered, map[string]any{"section": section, "entries": sections[section]})
		seen[section] = true
	}
	extra := make([]string, 0)
	for section := range sections {
		if !seen[section] {
			extra = append(extra, section)
		}
	}
	sort.Strings(extra)
	for _, section := range extra {
		ordered = append(ordered, map[string]any{"section": section, "entries": sections[section]})
	}

	return map[string]any{
		"version":  "1.0.0",
		"sections": ordered,
	}
}

func csvQuote(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func localMemorySource(metadata map[string]any, agentID string) string {
	if metadata != nil && stringValue(metadata["source"]) != "" {
		return stringValue(metadata["source"])
	}
	if agentID != "" {
		return "agent"
	}
	return "user"
}

func stringValue(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case []byte:
		return string(typed)
	case nil:
		return ""
	default:
		return fmt.Sprint(value)
	}
}

func stringArray(value any) []string {
	switch typed := value.(type) {
	case []string:
		return typed
	case []any:
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			items = append(items, stringValue(item))
		}
		return items
	default:
		return []string{}
	}
}

func cloneMap(input map[string]any) map[string]any {
	output := make(map[string]any, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}

func (s *Server) handleReadOnlyMemoryBodyFallback(w http.ResponseWriter, r *http.Request, procedure string) {
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), procedure, payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    procedure,
			},
		})
		return
	}

	writeJSON(w, http.StatusServiceUnavailable, map[string]any{
		"success": false,
		"error":   "memory route unavailable",
		"detail":  "upstream unavailable; local memory fallback has no persisted body results for " + procedure,
		"bridge": map[string]any{
			"fallback":  "go-local-memory",
			"procedure": procedure,
			"reason":    "upstream unavailable; local memory fallback has no persisted body results for " + procedure,
		},
	})
}

func (s *Server) handleMCPManualToolMutation(w http.ResponseWriter, r *http.Request, procedure string) {
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "invalid JSON body"})
		return
	}

	var result any
	upstreamBase, err := s.callUpstreamJSON(r.Context(), procedure, payload, &result)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
			"bridge": map[string]any{
				"upstreamBase": upstreamBase,
				"procedure":    procedure,
			},
		})
		return
	}

	writeJSON(w, http.StatusServiceUnavailable, map[string]any{
		"success": false,
		"error":   "MCP working-set mutation is unavailable: upstream MCP router is unavailable and the local MCP working set manager is not initialized.",
		"data": map[string]any{
			"ok":      false,
			"message": "MCP Server not initialized",
		},
		"bridge": map[string]any{
			"fallback":  "go-local-mcp",
			"procedure": procedure,
			"reason":    "upstream unavailable; local MCP working set manager is not initialized",
		},
	})
}

func (s *Server) localConfiguredMCPServers() ([]map[string]any, error) {
	editor, err := s.localMCPJsoncEditor()
	if err != nil {
		return nil, err
	}
	content, _ := editor["content"].(string)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(stripJSONCLineComments(content)), &parsed); err != nil {
		return nil, err
	}

	rawServers, _ := parsed["mcpServers"].(map[string]any)
	results := make([]map[string]any, 0, len(rawServers))
	for name, rawServer := range rawServers {
		serverMap, _ := rawServer.(map[string]any)
		serverType, _ := serverMap["type"].(string)
		if serverType == "" {
			if url, _ := serverMap["url"].(string); strings.TrimSpace(url) != "" {
				serverType = "STREAMABLE_HTTP"
			} else {
				serverType = "STDIO"
			}
		}

		command := nullableString(serverMap["command"])
		url := nullableString(serverMap["url"])
		description := nullableString(serverMap["description"])
		args := stringSlice(serverMap["args"])
		env := stringMap(serverMap["env"])
		headers := stringMap(serverMap["headers"])
		alwaysOn, _ := serverMap["always_on"].(bool)
		if !alwaysOn {
			if metaMap, ok := serverMap["_meta"].(map[string]any); ok {
				if metaAlwaysOn, ok := metaMap["alwaysOn"].(bool); ok {
					alwaysOn = metaAlwaysOn
				}
			}
		}

		results = append(results, map[string]any{
			"uuid":                         syntheticServerUUID(name),
			"name":                         name,
			"description":                  description,
			"type":                         serverType,
			"command":                      command,
			"args":                         args,
			"env":                          env,
			"url":                          url,
			"error_status":                 "unknown",
			"created_at":                   nil,
			"bearerToken":                  nullableString(serverMap["bearerToken"]),
			"headers":                      headers,
			"always_on":                    alwaysOn,
			"user_id":                      nil,
			"source_published_server_uuid": nullableString(serverMap["source_published_server_uuid"]),
			"_meta":                        serverMap["_meta"],
		})
	}

	sort.Slice(results, func(i, j int) bool {
		leftName, _ := results[i]["name"].(string)
		rightName, _ := results[j]["name"].(string)
		return leftName < rightName
	})
	return results, nil
}

func (s *Server) localConfiguredMCPServersFromDB() ([]map[string]any, error) {
	db, err := sql.Open("sqlite", s.localMetaMCPDBPath())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT uuid, name, description, type, command, args, env, url, error_status, created_at,
		       bearer_token, headers, always_on, user_id, source_published_server_uuid
		FROM mcp_servers
		ORDER BY name ASC
	`)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no such table") {
			return []map[string]any{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	servers := make([]map[string]any, 0)
	for rows.Next() {
		server, scanErr := scanLocalMCPServer(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		servers = append(servers, server)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	secretEnv, err := s.localWorkspaceSecretEnv()
	if err != nil {
		return nil, err
	}
	metaByName, err := s.localConfiguredMCPServerMetaByName()
	if err != nil {
		return nil, err
	}

	for _, server := range servers {
		mergeServerEnv(server, secretEnv)
		if name, _ := server["name"].(string); strings.TrimSpace(name) != "" {
			server["_meta"] = metaByName[name]
		}
	}

	return servers, nil
}

func (s *Server) localConfiguredMCPServerFromDB(uuid string) (map[string]any, error) {
	db, err := sql.Open("sqlite", s.localMetaMCPDBPath())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	row := db.QueryRow(`
		SELECT uuid, name, description, type, command, args, env, url, error_status, created_at,
		       bearer_token, headers, always_on, user_id, source_published_server_uuid
		FROM mcp_servers
		WHERE uuid = ?
		LIMIT 1
	`, uuid)
	server, err := scanLocalMCPServer(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	secretEnv, err := s.localWorkspaceSecretEnv()
	if err != nil {
		return nil, err
	}
	mergeServerEnv(server, secretEnv)

	metaByName, err := s.localConfiguredMCPServerMetaByName()
	if err != nil {
		return nil, err
	}
	if name, _ := server["name"].(string); strings.TrimSpace(name) != "" {
		server["_meta"] = metaByName[name]
	}

	return server, nil
}

func (s *Server) localConfiguredMCPServerMetaByName() (map[string]any, error) {
	editor, err := s.localMCPJsoncEditor()
	if err != nil {
		return nil, err
	}
	content, _ := editor["content"].(string)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(stripJSONCLineComments(content)), &parsed); err != nil {
		return nil, err
	}

	rawServers, _ := parsed["mcpServers"].(map[string]any)
	results := make(map[string]any, len(rawServers))
	for name, rawServer := range rawServers {
		serverMap, _ := rawServer.(map[string]any)
		if serverMap == nil {
			results[name] = nil
			continue
		}
		results[name] = serverMap["_meta"]
	}
	return results, nil
}

func (s *Server) localWorkspaceSecretEnv() (map[string]string, error) {
	db, err := sql.Open("sqlite", s.localMetaMCPDBPath())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT key, value
		FROM workspace_secrets
	`)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no such table") {
			return map[string]string{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	secretEnv := make(map[string]string)
	for rows.Next() {
		var (
			key   string
			value string
		)
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		secretEnv[key] = value
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return secretEnv, nil
}

func mergeServerEnv(server map[string]any, secretEnv map[string]string) {
	envValue, _ := server["env"].(map[string]any)
	merged := make(map[string]any, len(secretEnv)+len(envValue))
	for key, value := range secretEnv {
		merged[key] = value
	}
	for key, value := range envValue {
		merged[key] = value
	}
	server["env"] = merged
}

func (s *Server) localSkillsMetadata() ([]map[string]any, error) {
	skills, err := s.scanLocalSkills()
	if err != nil {
		return nil, err
	}
	results := make([]map[string]any, 0, len(skills))
	for _, skill := range skills {
		results = append(results, map[string]any{
			"id":          skill.ID,
			"name":        skill.Name,
			"description": skill.Description,
			"content":     skill.Content,
			"path":        skill.Path,
		})
	}
	return results, nil
}

func (s *Server) localSkillSummaries(query string) ([]SkillSummary, error) {
	skills, err := s.scanLocalSkills()
	if err != nil {
		return nil, err
	}
	query = strings.ToLower(strings.TrimSpace(query))
	results := make([]SkillSummary, 0, len(skills))
	for _, skill := range skills {
		summary := SkillSummary{
			ID:     skill.ID,
			Name:   skill.Name,
			Folder: filepath.Base(filepath.Dir(skill.Path)),
		}
		if query != "" {
			haystack := strings.ToLower(strings.Join([]string{summary.ID, summary.Name, summary.Folder}, " "))
			if !strings.Contains(haystack, query) {
				continue
			}
		}
		results = append(results, summary)
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Folder == results[j].Folder {
			return results[i].Name < results[j].Name
		}
		return results[i].Folder < results[j].Folder
	})
	return results, nil
}

func (s *Server) localReadSkill(name string) (map[string]any, error) {
	skills, err := s.scanLocalSkills()
	if err != nil {
		return nil, err
	}
	for _, skill := range skills {
		if skill.ID == name || skill.Name == name || filepath.Base(filepath.Dir(skill.Path)) == name {
			return map[string]any{
				"content": []map[string]any{
					{
						"type": "text",
						"text": skill.Content,
					},
				},
			}, nil
		}
	}
	return map[string]any{
		"content": []map[string]any{
			{
				"type": "text",
				"text": "Skill '" + name + "' not found.",
			},
		},
	}, nil
}

func (s *Server) localCreateSkill(payload map[string]any) (map[string]any, error) {
	id, _ := payload["id"].(string)
	name, _ := payload["name"].(string)
	description, _ := payload["description"].(string)
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("missing skill id")
	}
	if strings.TrimSpace(name) == "" {
		name = id
	}
	skillDir := filepath.Join(s.localSkillRoots()[1], id)
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		return nil, err
	}
	skillFile := filepath.Join(skillDir, "SKILL.md")
	body := strings.TrimSpace(description)
	if body == "" {
		body = name
	}
	content := "---\nname: " + name + "\ndescription: " + description + "\n---\n\n# " + name + "\n\n" + body + "\n\n## Instructions\n1. ...\n"
	if err := os.WriteFile(skillFile, []byte(content), 0o644); err != nil {
		return nil, err
	}
	return map[string]any{
		"content": []map[string]any{
			{
				"type": "text",
				"text": "Created skill '" + name + "' at " + skillFile,
			},
		},
	}, nil
}

func (s *Server) localSaveSkill(payload map[string]any) (map[string]any, error) {
	id, _ := payload["id"].(string)
	content, _ := payload["content"].(string)
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("missing skill id")
	}
	skills, err := s.scanLocalSkills()
	if err != nil {
		return nil, err
	}
	for _, skill := range skills {
		if skill.ID != id {
			continue
		}
		if err := os.WriteFile(skill.Path, []byte(content), 0o644); err != nil {
			return nil, err
		}
		return map[string]any{
			"content": []map[string]any{
				{
					"type": "text",
					"text": "Saved skill '" + id + "'.",
				},
			},
		}, nil
	}
	return map[string]any{
		"content": []map[string]any{
			{
				"type": "text",
				"text": "Skill '" + id + "' not found.",
			},
		},
	}, nil
}

func (s *Server) localSkillRoots() []string {
	return []string{
		filepath.Join(s.cfg.WorkspaceRoot, "packages", "core", "src", "skills"),
		filepath.Join(s.cfg.WorkspaceRoot, ".borg", "skills"),
	}
}

type localSkillRecord struct {
	ID          string
	Name        string
	Description string
	Content     string
	Path        string
}

func (s *Server) scanLocalSkills() ([]localSkillRecord, error) {
	results := make([]localSkillRecord, 0)
	seen := make(map[string]struct{})
	for _, root := range s.localSkillRoots() {
		if strings.TrimSpace(root) == "" {
			continue
		}
		if _, err := os.Stat(root); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if !strings.EqualFold(d.Name(), "SKILL.md") {
				return nil
			}
			record, err := parseLocalSkill(path)
			if err != nil {
				return nil
			}
			if _, ok := seen[record.Path]; ok {
				return nil
			}
			seen[record.Path] = struct{}{}
			results = append(results, record)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].ID == results[j].ID {
			return results[i].Path < results[j].Path
		}
		return results[i].ID < results[j].ID
	})
	return results, nil
}

func parseLocalSkill(path string) (localSkillRecord, error) {
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		return localSkillRecord{}, err
	}
	raw := string(contentBytes)
	name := filepath.Base(filepath.Dir(path))
	description := "No description provided"
	body := raw
	if strings.HasPrefix(raw, "---\n") {
		parts := strings.SplitN(raw, "\n---\n", 2)
		if len(parts) == 2 {
			frontmatter := strings.TrimPrefix(parts[0], "---\n")
			body = parts[1]
			for _, line := range strings.Split(frontmatter, "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "name:") {
					value := strings.TrimSpace(strings.TrimPrefix(line, "name:"))
					if value != "" {
						name = value
					}
				}
				if strings.HasPrefix(line, "description:") {
					value := strings.TrimSpace(strings.TrimPrefix(line, "description:"))
					if value != "" {
						description = value
					}
				}
			}
		}
	}
	return localSkillRecord{
		ID:          name,
		Name:        name,
		Description: description,
		Content:     body,
		Path:        path,
	}, nil
}

func (s *Server) localCreateConfiguredServer(payload map[string]any) (any, error) {
	name, _ := payload["name"].(string)
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("missing server name")
	}
	config, err := s.readLocalMCPConfigObject()
	if err != nil {
		return nil, err
	}
	servers, _ := config["mcpServers"].(map[string]any)
	if servers == nil {
		servers = map[string]any{}
	}
	servers[name] = configuredServerEntryFromPayload(payload)
	config["mcpServers"] = servers
	if err := s.writeLocalMCPConfigObject(config); err != nil {
		return nil, err
	}
	return s.localConfiguredServerByName(name)
}

func (s *Server) localUpdateConfiguredServer(payload map[string]any) (any, error) {
	uuid, _ := payload["uuid"].(string)
	config, err := s.readLocalMCPConfigObject()
	if err != nil {
		return nil, err
	}
	servers, _ := config["mcpServers"].(map[string]any)
	if servers == nil {
		return nil, errors.New("no configured servers")
	}
	targetName := strings.TrimSpace(nameForSyntheticUUID(servers, uuid))
	if targetName == "" {
		return nil, errors.New("configured server not found")
	}

	current, _ := servers[targetName].(map[string]any)
	if current == nil {
		current = map[string]any{}
	}
	nextName := targetName
	if updatedName, ok := payload["name"].(string); ok && strings.TrimSpace(updatedName) != "" {
		nextName = updatedName
	}
	updated := mergeConfiguredServerEntry(current, payload)
	delete(servers, targetName)
	servers[nextName] = updated
	config["mcpServers"] = servers
	if err := s.writeLocalMCPConfigObject(config); err != nil {
		return nil, err
	}
	return s.localConfiguredServerByName(nextName)
}

func (s *Server) localDeleteConfiguredServer(payload map[string]any) (any, error) {
	uuid, _ := payload["uuid"].(string)
	config, err := s.readLocalMCPConfigObject()
	if err != nil {
		return nil, err
	}
	servers, _ := config["mcpServers"].(map[string]any)
	if servers == nil {
		return map[string]any{"ok": true}, nil
	}
	targetName := strings.TrimSpace(nameForSyntheticUUID(servers, uuid))
	if targetName == "" {
		return map[string]any{"ok": true}, nil
	}
	delete(servers, targetName)
	config["mcpServers"] = servers
	if err := s.writeLocalMCPConfigObject(config); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true}, nil
}

func (s *Server) localReloadConfiguredServerMetadata(payload map[string]any) (any, error) {
	uuid, _ := payload["uuid"].(string)
	mode, _ := payload["mode"].(string)
	if strings.TrimSpace(mode) == "" {
		mode = "binary"
	}
	config, err := s.readLocalMCPConfigObject()
	if err != nil {
		return nil, err
	}
	servers, _ := config["mcpServers"].(map[string]any)
	if servers == nil {
		return nil, errors.New("no configured servers")
	}
	targetName := strings.TrimSpace(nameForSyntheticUUID(servers, uuid))
	if targetName == "" {
		return nil, errors.New("configured server not found")
	}
	entry, _ := servers[targetName].(map[string]any)
	if entry == nil {
		entry = map[string]any{}
	}
	now := time.Now().UTC().Format(time.RFC3339)
	meta, _ := entry["_meta"].(map[string]any)
	if meta == nil {
		meta = map[string]any{}
	}
	meta["status"] = "pending"
	meta["metadataVersion"] = 2
	meta["metadataSource"] = "derived"
	meta["cacheHydratedAt"] = now
	meta["lastAttemptedBinaryLoadAt"] = now
	meta["reloadableFromCache"] = false
	if _, ok := meta["toolCount"]; !ok {
		meta["toolCount"] = 0
	}
	if _, ok := meta["tools"]; !ok {
		meta["tools"] = []any{}
	}
	meta["error"] = "Go fallback refreshed metadata cache placeholder using local configuration only (" + mode + ")."
	entry["_meta"] = meta
	servers[targetName] = entry
	config["mcpServers"] = servers
	if err := s.writeLocalMCPConfigObject(config); err != nil {
		return nil, err
	}
	server, err := s.localConfiguredServerByName(targetName)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"server":         server,
		"metadata":       meta,
		"toolCount":      meta["toolCount"],
		"reloadDecision": "go-local-placeholder",
		"ok":             true,
	}, nil
}

func (s *Server) localClearConfiguredServerMetadata(payload map[string]any) (any, error) {
	uuid, _ := payload["uuid"].(string)
	config, err := s.readLocalMCPConfigObject()
	if err != nil {
		return nil, err
	}
	servers, _ := config["mcpServers"].(map[string]any)
	if servers == nil {
		return nil, errors.New("no configured servers")
	}
	targetName := strings.TrimSpace(nameForSyntheticUUID(servers, uuid))
	if targetName == "" {
		return nil, errors.New("configured server not found")
	}
	entry, _ := servers[targetName].(map[string]any)
	if entry == nil {
		entry = map[string]any{}
	}
	clearedAt := time.Now().UTC().Format(time.RFC3339)
	meta := map[string]any{
		"status":                     "pending",
		"metadataVersion":            2,
		"metadataSource":             "derived",
		"reloadableFromCache":        false,
		"toolCount":                  0,
		"tools":                      []any{},
		"error":                      "Cache cleared at " + clearedAt,
		"cacheHydratedAt":            nil,
		"discoveredAt":               nil,
		"lastAttemptedBinaryLoadAt":  nil,
		"lastSuccessfulBinaryLoadAt": nil,
	}
	entry["_meta"] = meta
	servers[targetName] = entry
	config["mcpServers"] = servers
	if err := s.writeLocalMCPConfigObject(config); err != nil {
		return nil, err
	}
	server, err := s.localConfiguredServerByName(targetName)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"server":    server,
		"metadata":  meta,
		"toolCount": 0,
		"ok":        true,
	}, nil
}

func (s *Server) localBulkImportConfiguredServers(payload []map[string]any) (any, error) {
	config, err := s.readLocalMCPConfigObject()
	if err != nil {
		return nil, err
	}
	servers, _ := config["mcpServers"].(map[string]any)
	if servers == nil {
		servers = map[string]any{}
	}

	if len(payload) == 0 {
		return []map[string]any{}, nil
	}

	importedNames := make([]string, 0, len(payload))
	for _, item := range payload {
		name, _ := item["name"].(string)
		if strings.TrimSpace(name) == "" {
			continue
		}
		servers[name] = configuredServerEntryFromPayload(item)
		importedNames = append(importedNames, name)
	}
	config["mcpServers"] = servers
	if err := s.writeLocalMCPConfigObject(config); err != nil {
		return nil, err
	}

	results := make([]map[string]any, 0, len(importedNames))
	for _, name := range importedNames {
		server, err := s.localConfiguredServerByName(name)
		if err != nil {
			return nil, err
		}
		results = append(results, server)
	}
	return results, nil
}

func (s *Server) localConfiguredServerByName(name string) (map[string]any, error) {
	servers, err := s.localConfiguredMCPServers()
	if err != nil {
		return nil, err
	}
	for _, server := range servers {
		serverName, _ := server["name"].(string)
		if serverName == name {
			return server, nil
		}
	}
	return nil, errors.New("configured server not found")
}

func (s *Server) readLocalMCPConfigObject() (map[string]any, error) {
	editor, err := s.localMCPJsoncEditor()
	if err != nil {
		return nil, err
	}
	content, _ := editor["content"].(string)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(stripJSONCLineComments(content)), &parsed); err != nil {
		return nil, err
	}
	if _, ok := parsed["mcpServers"]; !ok {
		parsed["mcpServers"] = map[string]any{}
	}
	return parsed, nil
}

func (s *Server) writeLocalMCPConfigObject(config map[string]any) error {
	return s.saveLocalMCPJsonc(prettyJSON(config))
}

func configuredServerEntryFromPayload(payload map[string]any) map[string]any {
	entry := map[string]any{}
	for _, key := range []string{"description", "command", "url", "bearerToken", "source_published_server_uuid", "type"} {
		if value, ok := payload[key]; ok && value != nil {
			entry[key] = value
		}
	}
	if args := stringSlice(payload["args"]); len(args) > 0 {
		entry["args"] = args
	}
	if env := stringMap(payload["env"]); len(env) > 0 {
		envMap := make(map[string]any, len(env))
		for key, value := range env {
			envMap[key] = value
		}
		entry["env"] = envMap
	}
	if headers := stringMap(payload["headers"]); len(headers) > 0 {
		headerMap := make(map[string]any, len(headers))
		for key, value := range headers {
			headerMap[key] = value
		}
		entry["headers"] = headerMap
	}
	if alwaysOn, ok := payload["always_on"].(bool); ok {
		entry["always_on"] = alwaysOn
	}
	return entry
}

func mergeConfiguredServerEntry(current map[string]any, patch map[string]any) map[string]any {
	next := make(map[string]any, len(current))
	for key, value := range current {
		next[key] = value
	}
	for key, value := range configuredServerEntryFromPayload(patch) {
		next[key] = value
	}
	return next
}

func nameForSyntheticUUID(servers map[string]any, uuid string) string {
	for name := range servers {
		if syntheticServerUUID(name) == uuid {
			return name
		}
	}
	return ""
}

func (s *Server) localMCPSyncTargets() ([]map[string]any, error) {
	clients := []string{"claude-desktop", "cursor", "vscode"}
	results := make([]map[string]any, 0, len(clients))
	for _, client := range clients {
		path, candidates, exists, err := s.resolveLocalClientConfigTarget(client, "")
		if err != nil {
			return nil, err
		}
		results = append(results, map[string]any{
			"client":     client,
			"path":       path,
			"candidates": candidates,
			"exists":     exists,
		})
	}
	return results, nil
}

func (s *Server) localMCPExportClientConfig(client string, overridePath string) (map[string]any, error) {
	servers, err := s.localConfiguredMCPServers()
	if err != nil {
		return nil, err
	}
	targetPath, _, exists, err := s.resolveLocalClientConfigTarget(client, overridePath)
	if err != nil {
		return nil, err
	}
	document := map[string]any{
		"mcpServers": buildClientConfigMCPServers(servers),
	}
	jsonText := prettyJSON(document) + "\n"
	return map[string]any{
		"client":      client,
		"targetPath":  targetPath,
		"existed":     exists,
		"serverCount": len(document["mcpServers"].(map[string]any)),
		"document":    document,
		"json":        jsonText,
	}, nil
}

func (s *Server) resolveLocalClientConfigTarget(client string, overridePath string) (string, []string, bool, error) {
	candidates := []string{}
	if strings.TrimSpace(overridePath) != "" {
		candidates = []string{overridePath}
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", nil, false, err
		}
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(homeDir, "AppData", "Roaming")
		}
		switch client {
		case "claude-desktop":
			candidates = []string{filepath.Join(appData, "Claude", "claude_desktop_config.json")}
		case "cursor":
			candidates = []string{
				filepath.Join(appData, "Cursor", "User", "globalStorage", "mcp-servers.json"),
				filepath.Join(appData, "Cursor", "User", "mcp.json"),
			}
		case "vscode":
			candidates = []string{
				filepath.Join(appData, "Code", "User", "globalStorage", "mcp-servers.json"),
				filepath.Join(appData, "Code", "User", "settings.json"),
				filepath.Join(s.cfg.WorkspaceRoot, ".vscode", "mcp.json"),
			}
		default:
			return "", nil, false, errors.New("unsupported client")
		}
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, candidates, true, nil
		}
	}
	if len(candidates) == 0 {
		return "", nil, false, errors.New("no client config candidates")
	}
	return candidates[0], candidates, false, nil
}

func buildClientConfigMCPServers(servers []map[string]any) map[string]any {
	result := make(map[string]any)
	for _, server := range servers {
		name, _ := server["name"].(string)
		if strings.TrimSpace(name) == "" {
			continue
		}
		serverType, _ := server["type"].(string)
		command, _ := server["command"].(string)
		url, _ := server["url"].(string)
		args := stringSlice(server["args"])
		env := stringMap(server["env"])
		headers := stringMap(server["headers"])
		bearerToken, _ := server["bearerToken"].(string)

		if serverType == "STDIO" {
			if strings.TrimSpace(command) == "" {
				continue
			}
			definition := map[string]any{
				"command": command,
			}
			if len(args) > 0 {
				definition["args"] = args
			}
			if len(env) > 0 {
				definition["env"] = env
			}
			result[name] = definition
			continue
		}
		if strings.TrimSpace(url) == "" {
			continue
		}
		definition := map[string]any{
			"url": url,
		}
		if bearerToken != "" {
			headers["Authorization"] = "Bearer " + bearerToken
		}
		if len(headers) > 0 {
			definition["headers"] = headers
		}
		result[name] = definition
	}
	return result
}

func (s *Server) localToolPreferences() (map[string]any, error) {
	editor, err := s.localMCPJsoncEditor()
	if err != nil {
		return nil, err
	}
	content, _ := editor["content"].(string)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(stripJSONCLineComments(content)), &parsed); err != nil {
		return nil, err
	}
	settings, _ := parsed["settings"].(map[string]any)
	toolSelection, _ := settings["toolSelection"].(map[string]any)
	return normalizeToolPreferences(toolSelection), nil
}

func (s *Server) saveLocalToolPreferences(payload map[string]any) (map[string]any, error) {
	editor, err := s.localMCPJsoncEditor()
	if err != nil {
		return nil, err
	}
	content, _ := editor["content"].(string)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(stripJSONCLineComments(content)), &parsed); err != nil {
		return nil, err
	}
	settings, _ := parsed["settings"].(map[string]any)
	if settings == nil {
		settings = map[string]any{}
	}
	currentToolSelection, _ := settings["toolSelection"].(map[string]any)
	current := normalizeToolPreferences(currentToolSelection)
	next := applyToolPreferencePatch(current, payload)
	settings["toolSelection"] = next
	parsed["settings"] = settings
	if _, ok := parsed["mcpServers"]; !ok {
		parsed["mcpServers"] = map[string]any{}
	}
	if err := s.saveLocalMCPJsonc(prettyJSON(parsed)); err != nil {
		return nil, err
	}
	result := make(map[string]any, len(next)+1)
	for key, value := range next {
		result[key] = value
	}
	result["ok"] = true
	return result, nil
}

func normalizeToolPreferences(raw map[string]any) map[string]any {
	return map[string]any{
		"importantTools":          normalizeToolNameList(raw["importantTools"]),
		"alwaysLoadedTools":       normalizeAlwaysLoadedTools(raw["alwaysLoadedTools"]),
		"autoLoadMinConfidence":   clampFloat(raw["autoLoadMinConfidence"], 0.85, 0.5, 0.99),
		"maxLoadedTools":          clampInt(raw["maxLoadedTools"], 16, 4, 64),
		"maxHydratedSchemas":      clampInt(raw["maxHydratedSchemas"], 8, 2, 32),
		"idleEvictionThresholdMs": clampInt(raw["idleEvictionThresholdMs"], 5*60*1000, 10_000, 24*60*60*1000),
	}
}

func applyToolPreferencePatch(current map[string]any, patch map[string]any) map[string]any {
	next := map[string]any{
		"importantTools":          current["importantTools"],
		"alwaysLoadedTools":       current["alwaysLoadedTools"],
		"autoLoadMinConfidence":   current["autoLoadMinConfidence"],
		"maxLoadedTools":          current["maxLoadedTools"],
		"maxHydratedSchemas":      current["maxHydratedSchemas"],
		"idleEvictionThresholdMs": current["idleEvictionThresholdMs"],
	}
	for _, key := range []string{"importantTools", "alwaysLoadedTools", "autoLoadMinConfidence", "maxLoadedTools", "maxHydratedSchemas", "idleEvictionThresholdMs"} {
		if value, ok := patch[key]; ok {
			next[key] = value
		}
	}
	normalizedPatch := map[string]any{
		"importantTools":          next["importantTools"],
		"alwaysLoadedTools":       next["alwaysLoadedTools"],
		"autoLoadMinConfidence":   next["autoLoadMinConfidence"],
		"maxLoadedTools":          next["maxLoadedTools"],
		"maxHydratedSchemas":      next["maxHydratedSchemas"],
		"idleEvictionThresholdMs": next["idleEvictionThresholdMs"],
	}
	return normalizeToolPreferences(normalizedPatch)
}

func normalizeToolNameList(value any) []string {
	result := []string{}
	seen := map[string]struct{}{}
	switch typed := value.(type) {
	case []any:
		for _, raw := range typed {
			name, _ := raw.(string)
			trimmed := strings.TrimSpace(name)
			if trimmed == "" {
				continue
			}
			if _, ok := seen[trimmed]; ok {
				continue
			}
			seen[trimmed] = struct{}{}
			result = append(result, trimmed)
		}
	case []string:
		for _, name := range typed {
			trimmed := strings.TrimSpace(name)
			if trimmed == "" {
				continue
			}
			if _, ok := seen[trimmed]; ok {
				continue
			}
			seen[trimmed] = struct{}{}
			result = append(result, trimmed)
		}
	}
	return result
}

func localSessionExportFormatDetection(raw string) map[string]any {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return map[string]any{
			"format": "invalid",
			"valid":  false,
		}
	}

	var parsed any
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		return map[string]any{
			"format": "invalid",
			"valid":  false,
		}
	}

	if record, ok := parsed.(map[string]any); ok {
		if version, _ := record["version"].(string); version == "1.0" {
			if sessions, ok := record["sessions"].([]any); ok {
				_ = sessions
				return map[string]any{
					"format": "borg-export",
					"valid":  true,
				}
			}
		}
		if recordType, _ := record["type"].(string); recordType == "conversation" {
			if _, ok := record["messages"]; ok {
				return map[string]any{
					"format": "claude-code",
					"valid":  true,
				}
			}
		}
	}

	if records, ok := parsed.([]any); ok && len(records) > 0 {
		if first, ok := records[0].(map[string]any); ok {
			if _, hasID := first["id"]; hasID {
				return map[string]any{
					"format": "generic-sessions",
					"valid":  true,
				}
			}
		}
	}

	return map[string]any{
		"format": "unknown",
		"valid":  true,
	}
}

func normalizeAlwaysLoadedTools(value any) []string {
	if value == nil {
		return []string{"search_tools", "read_file", "write_file", "grep_search", "execute_command", "browser__open"}
	}
	return normalizeToolNameList(value)
}

func clampFloat(value any, fallback float64, min float64, max float64) float64 {
	number, ok := value.(float64)
	if !ok {
		if intValue, ok := value.(int); ok {
			number = float64(intValue)
		} else {
			return fallback
		}
	}
	if number < min {
		return min
	}
	if number > max {
		return max
	}
	return number
}

func clampInt(value any, fallback int, min int, max int) int {
	number := fallback
	switch typed := value.(type) {
	case float64:
		number = int(typed)
	case int:
		number = typed
	default:
		return fallback
	}
	if number < min {
		return min
	}
	if number > max {
		return max
	}
	return number
}

func stripServerMeta(value any) any {
	servers, ok := value.(map[string]any)
	if !ok {
		return value
	}
	stripped := make(map[string]any, len(servers))
	for name, rawServer := range servers {
		serverMap, ok := rawServer.(map[string]any)
		if !ok {
			stripped[name] = rawServer
			continue
		}
		copyMap := make(map[string]any, len(serverMap))
		for key, fieldValue := range serverMap {
			if key == "_meta" {
				continue
			}
			copyMap[key] = fieldValue
		}
		stripped[name] = copyMap
	}
	return stripped
}

func prettyJSON(value any) string {
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(encoded)
}

func syntheticServerUUID(name string) string {
	hash := stableHash("mcp-server\n" + name)
	return hash[:8] + "-" + hash[8:12] + "-" + hash[12:16] + "-" + hash[16:20] + "-" + hash[20:32]
}

func nullableString(value any) any {
	text, _ := value.(string)
	if strings.TrimSpace(text) == "" {
		return nil
	}
	return text
}

func emptyStringToNilAny(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func stringSlice(value any) []string {
	values, ok := value.([]any)
	if !ok {
		if stringsValue, ok := value.([]string); ok {
			return append([]string(nil), stringsValue...)
		}
		return []string{}
	}
	result := make([]string, 0, len(values))
	for _, raw := range values {
		text, _ := raw.(string)
		if strings.TrimSpace(text) == "" {
			continue
		}
		result = append(result, text)
	}
	return result
}

func stringMap(value any) map[string]string {
	record, ok := value.(map[string]any)
	if !ok {
		if typedRecord, ok := value.(map[string]string); ok {
			copyMap := make(map[string]string, len(typedRecord))
			for key, entry := range typedRecord {
				copyMap[key] = entry
			}
			return copyMap
		}
		return map[string]string{}
	}
	result := make(map[string]string, len(record))
	for key, raw := range record {
		text, _ := raw.(string)
		if key == "" || text == "" {
			continue
		}
		result[key] = text
	}
	return result
}

func stripJSONCLineComments(content string) string {
	lines := strings.Split(content, "\n")
	for index, line := range lines {
		inString := false
		escaped := false
		for i := 0; i < len(line)-1; i++ {
			ch := line[i]
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = !inString
				continue
			}
			if !inString && ch == '/' && line[i+1] == '/' {
				line = line[:i]
				break
			}
		}
		lines[index] = line
	}
	return strings.Join(lines, "\n")
}

func (s *Server) handleRuntimeLocks(w http.ResponseWriter, _ *http.Request) {
	statuses := interop.DiscoverControlPlanes(s.cfg.MainLockPath(), s.cfg.LockPath())
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    statuses,
	})
}

func (s *Server) handleImportedInstructions(w http.ResponseWriter, _ *http.Request) {
	document := interop.ReadImportedInstructions(s.cfg.ImportedInstructionsPath())
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    document,
	})
}

func (s *Server) scanImportSources() ([]sessionimport.Candidate, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = s.cfg.MainConfigDir
	}

	scanner := sessionimport.NewScanner(s.cfg.WorkspaceRoot, homeDir, 50)
	return scanner.Scan()
}

func (s *Server) scanValidatedImportSources() ([]sessionimport.ValidationResult, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = s.cfg.MainConfigDir
	}

	scanner := sessionimport.NewScanner(s.cfg.WorkspaceRoot, homeDir, 50)
	return scanner.ScanValidated()
}

func (s *Server) importedSessionsArchiveRoot() string {
	return filepath.Join(s.cfg.WorkspaceRoot, ".borg", "imported_sessions", "archive")
}

func readGzipJSON(filePath string, target any) error {
	payload, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	reader, err := gzip.NewReader(bytes.NewReader(payload))
	if err != nil {
		return err
	}
	defer reader.Close()

	return json.NewDecoder(reader).Decode(target)
}

func readGzipText(filePath string) (string, error) {
	payload, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	reader, err := gzip.NewReader(bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	defer reader.Close()

	decoded, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func archivedImportedSessionRecord(sidecar importedSessionArchiveSidecar, metadataPath string) (ImportedSessionRecord, error) {
	transcriptPath := strings.TrimSuffix(metadataPath, ".meta.json.gz") + ".txt.gz"
	transcript, err := readGzipText(transcriptPath)
	if err != nil {
		return ImportedSessionRecord{}, err
	}

	sessionID := strings.TrimSpace(sidecar.SessionID)
	if sessionID == "" {
		sessionID = "import-" + stableHash(metadataPath)[:16]
	}
	transcriptHash := strings.TrimSpace(sidecar.TranscriptHash)
	if transcriptHash == "" {
		transcriptHash = stableHash(sidecar.SourceTool + "\n" + sidecar.SourcePath + "\n" + sidecar.SessionFormat)
	}

	title := sidecar.Title
	if title == nil || strings.TrimSpace(*title) == "" {
		titleText := filepath.Base(sidecar.SourcePath)
		if strings.TrimSpace(titleText) != "" {
			title = &titleText
		}
	}

	excerpt := sidecar.Excerpt
	if excerpt == nil && strings.TrimSpace(transcript) != "" {
		excerptText := transcript
		if len(excerptText) > 240 {
			excerptText = excerptText[:240]
		}
		excerpt = &excerptText
	}

	importedAt := sidecar.ArchivedAt
	if importedAt <= 0 {
		if stat, statErr := os.Stat(metadataPath); statErr == nil {
			importedAt = stat.ModTime().UTC().UnixMilli()
		} else {
			importedAt = time.Now().UTC().UnixMilli()
		}
	}

	normalized := map[string]any{
		"archiveFormat": "gzip-text-v1",
	}
	metadata := map[string]any{
		"archiveFormat":           "gzip-text-v1",
		"durableMemoryCount":      sidecar.DurableMemoryCount,
		"durableInstructionCount": sidecar.DurableInstructionCount,
		"memoryTags":              sidecar.MemoryTags,
	}
	if sidecar.RetentionSummary != nil {
		metadata["retentionSummary"] = sidecar.RetentionSummary
	}

	return ImportedSessionRecord{
		ID:                sessionID,
		SourceTool:        sidecar.SourceTool,
		SourcePath:        sidecar.SourcePath,
		ExternalSessionID: nil,
		Title:             title,
		SessionFormat:     sidecar.SessionFormat,
		Transcript:        transcript,
		Excerpt:           excerpt,
		WorkingDirectory:  sidecar.WorkingDirectory,
		TranscriptHash:    transcriptHash,
		NormalizedSession: normalized,
		Metadata:          metadata,
		DiscoveredAt:      importedAt,
		ImportedAt:        importedAt,
		LastModifiedAt:    &importedAt,
		CreatedAt:         importedAt,
		UpdatedAt:         importedAt,
		ParsedMemories:    []ImportedSessionMemory{},
	}, nil
}

func (s *Server) loadArchivedImportedSessionRecords() ([]ImportedSessionRecord, error) {
	archiveRoot := s.importedSessionsArchiveRoot()
	if _, err := os.Stat(archiveRoot); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	records := make([]ImportedSessionRecord, 0, 32)
	err := filepath.WalkDir(archiveRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".meta.json.gz") {
			return nil
		}

		var sidecar importedSessionArchiveSidecar
		if err := readGzipJSON(path, &sidecar); err != nil {
			return nil
		}
		record, err := archivedImportedSessionRecord(sidecar, path)
		if err != nil {
			return nil
		}
		records = append(records, record)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(records, func(i, j int) bool {
		if records[i].ImportedAt == records[j].ImportedAt {
			return records[i].ID < records[j].ID
		}
		return records[i].ImportedAt > records[j].ImportedAt
	})
	return records, nil
}

func archivedImportedSessionMaintenanceStats(records []ImportedSessionRecord) ImportedSessionMaintenanceStats {
	stats := ImportedSessionMaintenanceStats{
		TotalSessions:           len(records),
		ArchivedTranscriptCount: len(records),
	}
	for _, record := range records {
		if _, ok := record.Metadata["retentionSummary"]; !ok {
			stats.MissingRetentionSummaryCount++
		}
	}
	return stats
}

func (s *Server) importedInstructionDocPath() *string {
	document := interop.ReadImportedInstructions(s.cfg.ImportedInstructionsPath())
	if !document.Available {
		return nil
	}
	return &document.Path
}

func (s *Server) archivedImportedSessionScanSummary(records []ImportedSessionRecord) map[string]any {
	toolsSet := make(map[string]struct{})
	storedMemoryCount := 0
	for _, record := range records {
		if record.SourceTool != "" {
			toolsSet[record.SourceTool] = struct{}{}
		}
		storedMemoryCount += intNumber(record.Metadata["durableMemoryCount"])
	}

	tools := make([]string, 0, len(toolsSet))
	for tool := range toolsSet {
		tools = append(tools, tool)
	}
	sort.Strings(tools)

	return map[string]any{
		"discoveredCount":    len(records),
		"importedCount":      len(records),
		"skippedCount":       0,
		"storedMemoryCount":  storedMemoryCount,
		"instructionDocPath": s.importedInstructionDocPath(),
		"tools":              tools,
	}
}

func (s *Server) mergedImportedSessionScanSummary(records []ImportedSessionRecord, candidates []sessionimport.ValidationResult) map[string]any {
	seenSourcePaths := make(map[string]struct{}, len(records))
	toolsSet := make(map[string]struct{})
	storedMemoryCount := 0
	for _, record := range records {
		seenSourcePaths[record.SourcePath] = struct{}{}
		if record.SourceTool != "" {
			toolsSet[record.SourceTool] = struct{}{}
		}
		storedMemoryCount += intNumber(record.Metadata["durableMemoryCount"])
	}

	discoveredCount := len(records)
	skippedCount := 0
	for _, candidate := range candidates {
		if _, seen := seenSourcePaths[candidate.SourcePath]; !seen {
			discoveredCount++
		}
		if candidate.Valid {
			skippedCount++
		}
		if candidate.SourceTool != "" {
			toolsSet[candidate.SourceTool] = struct{}{}
		}
	}

	tools := make([]string, 0, len(toolsSet))
	for tool := range toolsSet {
		tools = append(tools, tool)
	}
	sort.Strings(tools)

	return map[string]any{
		"discoveredCount":    discoveredCount,
		"importedCount":      len(records),
		"skippedCount":       skippedCount,
		"storedMemoryCount":  storedMemoryCount,
		"instructionDocPath": s.importedInstructionDocPath(),
		"tools":              tools,
	}
}

func intNumber(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	default:
		return 0
	}
}

func (s *Server) importedSessionFallbackRecords(candidates []sessionimport.ValidationResult) []ImportedSessionRecord {
	records := make([]ImportedSessionRecord, 0, len(candidates))
	for _, candidate := range candidates {
		records = append(records, importedSessionFallbackRecord(candidate))
	}
	sort.Slice(records, func(i, j int) bool {
		return records[i].ID < records[j].ID
	})
	return records
}

func importedSessionFallbackRecord(candidate sessionimport.ValidationResult) ImportedSessionRecord {
	transcriptHash := stableHash(candidate.SourceTool + "\n" + candidate.SourcePath + "\n" + candidate.Format)
	id := "import-" + transcriptHash[:16]
	transcript := fallbackTranscript(candidate)
	excerptText := transcript
	if len(excerptText) > 240 {
		excerptText = excerptText[:240]
	}
	titleText := filepath.Base(candidate.SourcePath)
	lastModified := parseFallbackUnixMillis(candidate.LastModifiedAt)
	discoveredAt := time.Now().UTC().UnixMilli()
	if lastModified != nil {
		discoveredAt = *lastModified
	}

	var excerpt *string
	if strings.TrimSpace(excerptText) != "" {
		excerpt = &excerptText
	}
	var title *string
	if strings.TrimSpace(titleText) != "" {
		title = &titleText
	}

	normalized := map[string]any{
		"sourceType":     candidate.SourceType,
		"detectedModels": candidate.DetectedModels,
		"valid":          candidate.Valid,
	}
	if len(candidate.Errors) > 0 {
		normalized["errors"] = candidate.Errors
	}

	metadata := map[string]any{
		"fallback":       "go-sessionimport",
		"estimatedSize":  candidate.EstimatedSize,
		"sourceType":     candidate.SourceType,
		"detectedModels": candidate.DetectedModels,
		"valid":          candidate.Valid,
	}
	if len(candidate.Errors) > 0 {
		metadata["errors"] = candidate.Errors
	}

	return ImportedSessionRecord{
		ID:                id,
		SourceTool:        candidate.SourceTool,
		SourcePath:        candidate.SourcePath,
		ExternalSessionID: nil,
		Title:             title,
		SessionFormat:     candidate.Format,
		Transcript:        transcript,
		Excerpt:           excerpt,
		WorkingDirectory:  nil,
		TranscriptHash:    transcriptHash,
		NormalizedSession: normalized,
		Metadata:          metadata,
		DiscoveredAt:      discoveredAt,
		ImportedAt:        discoveredAt,
		LastModifiedAt:    lastModified,
		CreatedAt:         discoveredAt,
		UpdatedAt:         discoveredAt,
		ParsedMemories:    []ImportedSessionMemory{},
	}
}

func fallbackTranscript(candidate sessionimport.ValidationResult) string {
	lines := []string{
		"Go fallback imported session summary",
		"sourceTool: " + candidate.SourceTool,
		"sourcePath: " + candidate.SourcePath,
		"sourceType: " + candidate.SourceType,
		"sessionFormat: " + candidate.Format,
	}
	if len(candidate.DetectedModels) > 0 {
		lines = append(lines, "detectedModels: "+strings.Join(candidate.DetectedModels, ", "))
	}
	if len(candidate.Errors) > 0 {
		lines = append(lines, "errors: "+strings.Join(candidate.Errors, "; "))
	}
	return strings.Join(lines, "\n")
}

func parseFallbackUnixMillis(value string) *int64 {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, trimmed)
	if err != nil {
		return nil
	}
	millis := parsed.UTC().UnixMilli()
	return &millis
}

func stableHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func (s *Server) importRoots() []sessionimport.RootStatus {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = s.cfg.MainConfigDir
	}

	scanner := sessionimport.NewScanner(s.cfg.WorkspaceRoot, homeDir, 50)
	return scanner.Roots()
}

func (s *Server) discoveredSessions() ([]Session, error) {
	candidates, err := s.scanValidatedImportSources()
	if err != nil {
		return nil, err
	}

	sessions := make([]Session, 0, len(candidates))
	for index, candidate := range candidates {
		sessions = append(sessions, Session{
			ID:             "discovered_" + fmtInt(index+1),
			CLIType:        candidate.SourceTool,
			Status:         "discovered",
			Task:           candidate.SourceType,
			StartedAt:      candidate.LastModifiedAt,
			SourcePath:     candidate.SourcePath,
			SessionFormat:  candidate.Format,
			Valid:          candidate.Valid,
			DetectedModels: candidate.DetectedModels,
		})
	}
	return sessions, nil
}

func summarizeSessions(sessions []Session) SessionSummary {
	byCLIType := make(map[string]int)
	byFormat := make(map[string]int)
	byTask := make(map[string]int)
	byModelHint := make(map[string]int)
	validCount := 0

	for _, session := range sessions {
		byCLIType[session.CLIType]++
		byFormat[session.SessionFormat]++
		byTask[session.Task]++
		if session.Valid {
			validCount++
		}
		for _, model := range session.DetectedModels {
			byModelHint[model]++
		}
	}

	return SessionSummary{
		Count:       len(sessions),
		ValidCount:  validCount,
		ByCLIType:   summaryBucketsFromMap(byCLIType),
		ByFormat:    summaryBucketsFromMap(byFormat),
		ByTask:      summaryBucketsFromMap(byTask),
		ByModelHint: summaryBucketsFromMap(byModelHint),
	}
}

func summaryBucketsFromMap(values map[string]int) []SummaryBucket {
	buckets := make([]SummaryBucket, 0, len(values))
	for key, count := range values {
		if key == "" {
			key = "unknown"
		}
		buckets = append(buckets, SummaryBucket{Key: key, Count: count})
	}
	for i := 0; i < len(buckets); i++ {
		for j := i + 1; j < len(buckets); j++ {
			if buckets[j].Count > buckets[i].Count || (buckets[j].Count == buckets[i].Count && buckets[j].Key < buckets[i].Key) {
				buckets[i], buckets[j] = buckets[j], buckets[i]
			}
		}
	}
	return buckets
}

func toHTTPBuckets(values []providers.SummaryBucket) []SummaryBucket {
	buckets := make([]SummaryBucket, 0, len(values))
	for _, value := range values {
		buckets = append(buckets, SummaryBucket{
			Key:   value.Key,
			Count: value.Count,
		})
	}
	return buckets
}

func toImportBuckets(values []sessionimport.SummaryBucket) []SummaryBucket {
	buckets := make([]SummaryBucket, 0, len(values))
	for _, value := range values {
		buckets = append(buckets, SummaryBucket{
			Key:   value.Key,
			Count: value.Count,
		})
	}
	return buckets
}

func detectImportSourceTool(targetPath string) string {
	lowerPath := strings.ToLower(targetPath)
	switch {
	case strings.Contains(lowerPath, ".claude"):
		return "claude-code"
	case strings.Contains(lowerPath, ".copilot"):
		return "copilot-cli"
	case strings.Contains(lowerPath, "chatgpt"), strings.Contains(lowerPath, ".openai"), strings.Contains(lowerPath, "openai"):
		return "openai"
	default:
		return "unknown"
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func jsonNumber(value int) string {
	return string(rune('0' + 0))[:0] + fmtInt(value)
}

func fmtInt(value int) string {
	if value == 0 {
		return "0"
	}

	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}

	buf := [20]byte{}
	index := len(buf)
	for value > 0 {
		index--
		buf[index] = byte('0' + (value % 10))
		value /= 10
	}

	return sign + string(buf[index:])
}

// --- Repograph Handlers ---

func (s *Server) handleRepoGraphBuild(w http.ResponseWriter, r *http.Request) {
	graph, err := s.repoGraph.Build(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": graph})
}

func (s *Server) handleRepoGraphGet(w http.ResponseWriter, _ *http.Request) {
	graph := s.repoGraph.GetGraph()
	if graph == nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": nil, "message": "graph not built yet"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": graph})
}

func (s *Server) handleRepoGraphReferences(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing symbol parameter"})
		return
	}
	refs := s.repoGraph.FindReferences(symbol)
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": refs})
}

func (s *Server) handleRepoGraphDependents(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "missing path parameter"})
		return
	}
	deps := s.repoGraph.FindDependents(path)
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": deps})
}

func (s *Server) handleRepoGraphSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	results := s.repoGraph.SearchSymbols(query, limit)
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": results})
}
