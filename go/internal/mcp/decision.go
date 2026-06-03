package mcp

/**
	"strings"
 * @module go/internal/mcp
 *
 * WHAT: The TormentNexus MCP Decision System — a ranked discovery, auto-load,
 *       and unified tool search-and-call engine.
 *
 * WHY: Models fail not because tools are missing, but because they face too
 *      many choices or are forced through a ceremonial search→load→call dance.
 *      This system exposes only 5-6 permanent meta-tools, uses ranked discovery
 *      with silent high-confidence auto-load, deferred binary startup,
 *      and LRU eviction to keep the loaded set tiny.
 *
 * DESIGN RULES (from aggregator research):
 *   - The model should almost never face more than a handful of visible choices.
 *   - The model should almost never be forced to manually perform the full
 *     discovery workflow when the system already knows the likely best capability.
 *   - Tiny permanent meta-tool surface (search_tools, load_tool, call_tool,
 *     list_loaded_tools, unload_tool)
 *   - Ranked discovery, not raw search
 *   - Silent auto-load when confidence is high
 *   - Deferred binary startup (index metadata without spawning)
 *   - Small active loaded set with LRU eviction
 *   - Profiles for common workflows
 *   - Code mode for multi-step execution
 *   - Strong observability for routing improvement
 *
 * ADDED: v1.0.0-alpha.32
*/

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// ---------- Configuration ----------

type DecisionConfig struct {
	// SoftCap is the soft limit on loaded tool metadata entries.
	SoftCap int `json:"softCap"`
	// HardCap is the hard limit; triggers aggressive LRU eviction.
	HardCap int `json:"hardCap"`
	// ActiveBinaryCap is the max concurrently-running MCP server binaries.
	ActiveBinaryCap int `json:"activeBinaryCap"`
	// HighConfidenceThreshold: scores above this trigger silent auto-load.
	HighConfidenceThreshold float64 `json:"highConfidenceThreshold"`
	// SearchResultLimit: max results returned from search.
	SearchResultLimit int `json:"searchResultLimit"`
	// IdleTimeout: tools idle longer than this are candidates for eviction.
	IdleTimeout time.Duration `json:"idleTimeout"`
	// CatalogDBPath: path to the catalog SQLite database for persisted metadata.
	CatalogDBPath string `json:"catalogDBPath"`
	// Profile: active profile name (e.g., "web-research", "repo-coding").
	Profile string `json:"profile"`
}

func DefaultDecisionConfig() DecisionConfig {
	return DecisionConfig{
		SoftCap:                 16,
		HardCap:                 24,
		ActiveBinaryCap:         4,
		HighConfidenceThreshold: 50.0,
		SearchResultLimit:       8,
		IdleTimeout:             10 * time.Minute,
		Profile:                 "general",
	}
}

// ---------- Loaded Tool Tracking ----------

type LoadedTool struct {
	ToolEntry
	LoadedAt   time.Time `json:"loadedAt"`
	LastUsedAt time.Time `json:"lastUsedAt"`
	UseCount   int       `json:"useCount"`
	BinaryLive bool      `json:"binaryLive"` // true if server binary is actually running
	AutoLoaded bool      `json:"autoLoaded"` // true if loaded by auto-load, not explicit request
}

// ---------- Search Result ----------

type UnifiedSearchResult struct {
	Rank           int                `json:"rank"`
	Name           string             `json:"name"`
	OriginalName   string             `json:"originalName"`
	Server         string             `json:"server"`
	Description    string             `json:"description"`
	Score          float64            `json:"score"`
	ScoreBreakdown map[string]float64 `json:"scoreBreakdown,omitempty"`
	MatchReason    string             `json:"matchReason"`
	IsLoaded       bool               `json:"isLoaded"`
	IsAlwaysOn     bool               `json:"isAlwaysOn"`
	RequiresBinary bool               `json:"requiresBinary"`
	TypicalLatency string             `json:"typicalLatency,omitempty"`
	ShortExample   string             `json:"shortExample,omitempty"`
	InputSchema    interface{}        `json:"inputSchema,omitempty"`
}

// ---------- Observability Event ----------

type DecisionEvent struct {
	Timestamp  time.Time              `json:"timestamp"`
	Type       string                 `json:"type"` // "search", "load", "call", "evict", "autoload", "unload"
	Query      string                 `json:"query,omitempty"`
	ToolName   string                 `json:"toolName,omitempty"`
	ServerName string                 `json:"serverName,omitempty"`
	Score      float64                `json:"score,omitempty"`
	Reason     string                 `json:"reason,omitempty"`
	Latency    time.Duration          `json:"latency,omitempty"`
	Success    bool                   `json:"success"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ---------- Decision System ----------

type DecisionSystem struct {
	cfg       DecisionConfig
	mu        sync.RWMutex
	loaded    map[string]*LoadedTool // keyed by advertised name
	known     []ToolEntry            // full catalog of known tools
	events    []DecisionEvent        // circular buffer of observability events
	agg       *Aggregator            // live MCP connections
	catalog   []ToolEntry            // persisted catalog loaded from disk
	eventIdx  int                    // circular buffer write position
	maxEvents int
}

func NewDecisionSystem(cfg DecisionConfig, agg *Aggregator) *DecisionSystem {
	if cfg.SoftCap <= 0 {
		cfg.SoftCap = 16
	}
	if cfg.HardCap <= 0 {
		cfg.HardCap = 24
	}
	if cfg.ActiveBinaryCap <= 0 {
		cfg.ActiveBinaryCap = 4
	}
	if cfg.HighConfidenceThreshold <= 0 {
		cfg.HighConfidenceThreshold = 50.0
	}
	if cfg.SearchResultLimit <= 0 {
		cfg.SearchResultLimit = 8
	}
	if cfg.IdleTimeout <= 0 {
		cfg.IdleTimeout = 10 * time.Minute
	}

	ds := &DecisionSystem{
		cfg:       cfg,
		loaded:    make(map[string]*LoadedTool),
		known:     []ToolEntry{},
		events:    make([]DecisionEvent, 200),
		agg:       agg,
		catalog:   []ToolEntry{},
		maxEvents: 200,
	}

	return ds
}

// ---------- Meta-Tool: search_tools ----------

// SearchTools performs ranked discovery across all known tools (catalog + loaded + live).
// Returns compact, model-friendly results.
func (ds *DecisionSystem) SearchTools(ctx context.Context, query string) ([]UnifiedSearchResult, error) {
	start := time.Now()

	allTools := ds.getAllKnownTools()
	ranked := RankTools(query, allTools, ds.cfg.SearchResultLimit)

	var results []UnifiedSearchResult
	for _, r := range ranked {
		_, isLoaded := ds.loaded[r.AdvertisedName]
		results = append(results, UnifiedSearchResult{
			Rank:           r.Rank,
			Name:           r.AdvertisedName,
			OriginalName:   r.OriginalName,
			Server:         r.Server,
			Description:    ds.truncateDescription(r.Description, 120),
			Score:          r.Score,
			ScoreBreakdown: r.ScoreBreakdown,
			MatchReason:    r.MatchReason,
			IsLoaded:       isLoaded,
			IsAlwaysOn:     r.AlwaysOn,
			RequiresBinary: ds.requiresBinary(r.Server),
			TypicalLatency: ds.estimateLatency(r.Server),
			ShortExample:   ds.generateExample(r),
		})
	}

	ds.recordEvent(DecisionEvent{
		Type:    "search",
		Query:   query,
		Success: true,
		Latency: time.Since(start),
		Metadata: map[string]interface{}{
			"resultCount": len(results),
		},
	})

	return results, nil
}

// ---------- Meta-Tool: search_and_call (one-shot) ----------

// SearchAndCall performs the full discover→select→execute pipeline in one shot.
// If high-confidence match found, auto-loads and calls immediately.
// If ambiguous, returns ranked options without calling.
func (ds *DecisionSystem) SearchAndCall(ctx context.Context, query string, arguments map[string]interface{}) (*CallResult, error) {
	start := time.Now()

	allTools := ds.getAllKnownTools()
	ranked := RankTools(query, allTools, 3)

	if len(ranked) == 0 {
		ds.recordEvent(DecisionEvent{
			Type:    "call",
			Query:   query,
			Success: false,
			Latency: time.Since(start),
			Reason:  "no matching tools found",
		})
		return nil, fmt.Errorf("no tools found matching %q", query)
	}

	// Pick top result if high confidence, or if only one result
	best := ranked[0]
	if best.Score >= ds.cfg.HighConfidenceThreshold || len(ranked) == 1 {
		// Auto-load and call
		result, err := ds.CallTool(ctx, best.AdvertisedName, arguments)
		if err != nil {
			ds.recordEvent(DecisionEvent{
				Type:       "call",
				Query:      query,
				ToolName:   best.AdvertisedName,
				ServerName: best.Server,
				Score:      best.Score,
				Success:    false,
				Latency:    time.Since(start),
				Reason:     err.Error(),
			})
			return nil, err
		}

		ds.recordEvent(DecisionEvent{
			Type:       "call",
			Query:      query,
			ToolName:   best.AdvertisedName,
			ServerName: best.Server,
			Score:      best.Score,
			Success:    true,
			Latency:    time.Since(start),
			Reason:     "auto-selected (high confidence)",
		})

		return result, nil
	}

	// Ambiguous — return options
	ds.recordEvent(DecisionEvent{
		Type:    "call",
		Query:   query,
		Success: false,
		Reason:  fmt.Sprintf("ambiguous (%d candidates, top score %.1f below threshold %.1f)", len(ranked), best.Score, ds.cfg.HighConfidenceThreshold),
		Latency: time.Since(start),
	})

	return nil, fmt.Errorf("ambiguous query %q — %d candidates found (top: %s, score: %.1f). Use search_tools first to select.",
		query, len(ranked), best.AdvertisedName, best.Score)
}

// ---------- Meta-Tool: load_tool ----------

func (ds *DecisionSystem) LoadTool(ctx context.Context, advertisedName string) (*LoadedTool, error) {
	start := time.Now()
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if lt, ok := ds.loaded[advertisedName]; ok {
		lt.LastUsedAt = time.Now()
		lt.UseCount++
		ds.recordEventLocked(DecisionEvent{
			Type:     "load",
			ToolName: advertisedName,
			Success:  true,
			Latency:  time.Since(start),
			Reason:   "already loaded",
		})
		return lt, nil
	}

	// Find in known catalog
	var found *ToolEntry
	for i := range ds.known {
		if ds.known[i].AdvertisedName == advertisedName || ds.known[i].Name == advertisedName {
			found = &ds.known[i]
			break
		}
	}
	if found == nil {
		// Try to find in all catalog
		allTools := ds.getAllKnownToolsLocked()
		for i := range allTools {
			if allTools[i].AdvertisedName == advertisedName || allTools[i].Name == advertisedName {
				found = &allTools[i]
				break
			}
		}
	}
	if found == nil {
		ds.recordEventLocked(DecisionEvent{
			Type:     "load",
			ToolName: advertisedName,
			Success:  false,
			Latency:  time.Since(start),
			Reason:   "tool not found in catalog",
		})
		return nil, fmt.Errorf("tool %q not found in catalog", advertisedName)
	}

	// Enforce loaded set caps
	ds.evictIfNeededLocked()

	lt := &LoadedTool{
		ToolEntry:  *found,
		LoadedAt:   time.Now(),
		LastUsedAt: time.Now(),
		UseCount:   1,
		BinaryLive: false,
		AutoLoaded: false,
	}
	ds.loaded[advertisedName] = lt

	ds.recordEventLocked(DecisionEvent{
		Type:       "load",
		ToolName:   advertisedName,
		ServerName: found.Server,
		Success:    true,
		Latency:    time.Since(start),
	})

	return lt, nil
}

// ---------- Meta-Tool: call_tool ----------

type CallResult struct {
	ToolName   string      `json:"toolName"`
	Server     string      `json:"server"`
	Result     interface{} `json:"result"`
	IsError    bool        `json:"isError,omitempty"`
	Duration   string      `json:"duration"`
	AutoLoaded bool        `json:"autoLoaded"`
}

func (ds *DecisionSystem) CallTool(ctx context.Context, advertisedName string, arguments map[string]interface{}) (*CallResult, error) {
	start := time.Now()

	// Auto-load if needed
	ds.mu.Lock()
	lt, isLoaded := ds.loaded[advertisedName]
	if !isLoaded {
		ds.mu.Unlock()
		loaded, err := ds.LoadTool(ctx, advertisedName)
		if err != nil {
			return nil, err
		}
		loaded.AutoLoaded = true
		lt = loaded
	} else {
		lt.LastUsedAt = time.Now()
		lt.UseCount++
		ds.mu.Unlock()
	}

	// Call through aggregator if available
	if ds.agg != nil {
		resp, err := ds.agg.CallTool(ctx, lt.Server, lt.OriginalName, arguments)
		duration := time.Since(start)
		if err != nil {
			ds.recordEvent(DecisionEvent{
				Type:       "call",
				ToolName:   advertisedName,
				ServerName: lt.Server,
				Success:    false,
				Latency:    duration,
				Reason:     err.Error(),
			})
			return nil, err
		}

		result := &CallResult{
			ToolName:   advertisedName,
			Server:     lt.Server,
			Result:     resp.Result,
			Duration:   duration.String(),
			AutoLoaded: lt.AutoLoaded,
		}

		ds.recordEvent(DecisionEvent{
			Type:       "call",
			ToolName:   advertisedName,
			ServerName: lt.Server,
			Success:    true,
			Latency:    duration,
		})

		return result, nil
	}

	// No aggregator — return tool metadata as result (catalog-only mode)
	duration := time.Since(start)
	return &CallResult{
		ToolName:   advertisedName,
		Server:     lt.Server,
		Result:     map[string]string{"status": "catalog_only", "message": "tool metadata available but no live MCP connection"},
		Duration:   duration.String(),
		AutoLoaded: lt.AutoLoaded,
	}, nil
}

// ---------- Meta-Tool: list_loaded_tools ----------

type LoadedToolSummary struct {
	Name        string    `json:"name"`
	Server      string    `json:"server"`
	Description string    `json:"description"`
	LoadedAt    time.Time `json:"loadedAt"`
	LastUsedAt  time.Time `json:"lastUsedAt"`
	UseCount    int       `json:"useCount"`
	IsAlwaysOn  bool      `json:"isAlwaysOn"`
	AutoLoaded  bool      `json:"autoLoaded"`
}

func (ds *DecisionSystem) ListLoadedTools() []LoadedToolSummary {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	var summaries []LoadedToolSummary
	for _, lt := range ds.loaded {
		summaries = append(summaries, LoadedToolSummary{
			Name:        lt.AdvertisedName,
			Server:      lt.Server,
			Description: ds.truncateDescription(lt.Description, 80),
			LoadedAt:    lt.LoadedAt,
			LastUsedAt:  lt.LastUsedAt,
			UseCount:    lt.UseCount,
			IsAlwaysOn:  lt.AlwaysOn,
			AutoLoaded:  lt.AutoLoaded,
		})
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Name < summaries[j].Name
	})

	return summaries
}

// ---------- Meta-Tool: unload_tool ----------

func (ds *DecisionSystem) UnloadTool(advertisedName string) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if _, ok := ds.loaded[advertisedName]; !ok {
		return fmt.Errorf("tool %q not loaded", advertisedName)
	}

	delete(ds.loaded, advertisedName)

	ds.recordEventLocked(DecisionEvent{
		Type:     "unload",
		ToolName: advertisedName,
		Success:  true,
		Reason:   "explicit unload",
	})

	return nil
}

// ---------- Meta-Tool: list_all_tools (always-on + builtin overview) ----------

type ToolOverview struct {
	AlwaysOnTools  []ToolSummary `json:"alwaysOnTools"`
	BuiltinTools   []ToolSummary `json:"builtinTools"`
	LoadedTools    []string      `json:"loadedTools"`
	TotalKnown     int           `json:"totalKnown"`
	TotalLoaded    int           `json:"totalLoaded"`
	ActiveBinaries int           `json:"activeBinaries"`
}

type ToolSummary struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Server      string `json:"server"`
}

func (ds *DecisionSystem) ListAllTools() *ToolOverview {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	allTools := ds.getAllKnownToolsLocked()

	overview := &ToolOverview{
		TotalKnown:  len(allTools),
		TotalLoaded: len(ds.loaded),
	}

	var alwaysOn, builtin []ToolSummary
	for _, t := range allTools {
		summary := ToolSummary{
			Name:        t.AdvertisedName,
			Description: ds.truncateDescription(t.Description, 80),
			Server:      t.Server,
		}
		if t.AlwaysOn {
			alwaysOn = append(alwaysOn, summary)
		}
		if t.Server == "builtin" || t.Server == "tormentnexus" {
			builtin = append(builtin, summary)
		}
	}

	for name := range ds.loaded {
		overview.LoadedTools = append(overview.LoadedTools, name)
	}

	sort.Slice(alwaysOn, func(i, j int) bool { return alwaysOn[i].Name < alwaysOn[j].Name })
	sort.Slice(builtin, func(i, j int) bool { return builtin[i].Name < builtin[j].Name })

	overview.AlwaysOnTools = alwaysOn
	overview.BuiltinTools = builtin
	return overview
}

// ---------- Catalog Management ----------

// LoadCatalog loads a persisted tool catalog from a JSON file.
func (ds *DecisionSystem) LoadCatalog(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var catalog []ToolEntry
	if err := json.Unmarshal(data, &catalog); err != nil {
		return err
	}

	ds.mu.Lock()
	ds.catalog = catalog
	ds.rebuildKnownLocked()
	ds.mu.Unlock()

	return nil
}

// AddCatalogEntries adds tool entries to the catalog.
func (ds *DecisionSystem) AddCatalogEntries(entries []ToolEntry) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	// Deduplicate by advertised name
	existing := make(map[string]bool)
	for _, t := range ds.catalog {
		existing[t.AdvertisedName] = true
	}

	for _, entry := range entries {
		if !existing[entry.AdvertisedName] {
			ds.catalog = append(ds.catalog, entry)
			existing[entry.AdvertisedName] = true
		}
	}

	ds.rebuildKnownLocked()
}

// SaveCatalog persists the current catalog to a JSON file.
func (ds *DecisionSystem) SaveCatalog(path string) error {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(ds.catalog, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

// RefreshFromInventory loads tools from the live inventory system.
func (ds *DecisionSystem) RefreshFromInventory(inv *Inventory) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	// Mark all existing tools from inventory as needing refresh
	existing := make(map[string]bool)
	for _, t := range ds.catalog {
		existing[t.AdvertisedName] = true
	}

	for _, tool := range inv.Tools {
		if !existing[tool.AdvertisedName] {
			ds.catalog = append(ds.catalog, tool)
			existing[tool.AdvertisedName] = true
		}
	}

	ds.rebuildKnownLocked()
}

// ---------- Observability ----------

func (ds *DecisionSystem) GetEvents(limit int) []DecisionEvent {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	var events []DecisionEvent
	count := 0
	for i := ds.eventIdx - 1; i >= 0 && count < limit; i-- {
		if i < 0 {
			i += ds.maxEvents
		}
		events = append(events, ds.events[i])
		count++
		if count >= ds.maxEvents {
			break
		}
	}
	return events
}

// ---------- LRU Eviction ----------

func (ds *DecisionSystem) evictIfNeededLocked() {
	// Evict if at hard cap
	for len(ds.loaded) >= ds.cfg.HardCap {
		ds.evictLRULocked()
	}

	// Evict idle tools if at soft cap
	if len(ds.loaded) >= ds.cfg.SoftCap {
		now := time.Now()
		var idleKeys []string
		for k, lt := range ds.loaded {
			if !lt.AlwaysOn && now.Sub(lt.LastUsedAt) > ds.cfg.IdleTimeout {
				idleKeys = append(idleKeys, k)
			}
		}
		sort.Slice(idleKeys, func(i, j int) bool {
			return ds.loaded[idleKeys[i]].LastUsedAt.Before(ds.loaded[idleKeys[j]].LastUsedAt)
		})
		for _, k := range idleKeys {
			if len(ds.loaded) < ds.cfg.SoftCap {
				break
			}
			ds.recordEventLocked(DecisionEvent{
				Type:     "evict",
				ToolName: k,
				Reason:   "idle eviction (LRU)",
			})
			delete(ds.loaded, k)
		}
	}
}

func (ds *DecisionSystem) evictLRULocked() {
	var oldest string
	var oldestTime time.Time
	for k, lt := range ds.loaded {
		if lt.AlwaysOn {
			continue
		}
		if oldest == "" || lt.LastUsedAt.Before(oldestTime) {
			oldest = k
			oldestTime = lt.LastUsedAt
		}
	}
	if oldest != "" {
		ds.recordEventLocked(DecisionEvent{
			Type:     "evict",
			ToolName: oldest,
			Reason:   "hard cap eviction (LRU)",
		})
		delete(ds.loaded, oldest)
	}
}

// ---------- Internal Helpers ----------

func (ds *DecisionSystem) getAllKnownTools() []ToolEntry {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.getAllKnownToolsLocked()
}

func (ds *DecisionSystem) getAllKnownToolsLocked() []ToolEntry {
	seen := make(map[string]bool)
	var all []ToolEntry

	// Always-on tools first
	for _, t := range ds.known {
		if t.AlwaysOn {
			all = append(all, t)
			seen[t.AdvertisedName] = true
		}
	}

	// Loaded tools
	for _, lt := range ds.loaded {
		if !seen[lt.AdvertisedName] {
			all = append(all, lt.ToolEntry)
			seen[lt.AdvertisedName] = true
		}
	}

	// Catalog
	for _, t := range ds.catalog {
		if !seen[t.AdvertisedName] {
			all = append(all, t)
			seen[t.AdvertisedName] = true
		}
	}

	return all
}

func (ds *DecisionSystem) rebuildKnownLocked() {
	ds.known = append([]ToolEntry{}, ds.catalog...)

	// Ensure always-on tools are loaded
	for i := range ds.known {
		if ds.known[i].AlwaysOn {
			name := ds.known[i].AdvertisedName
			if _, ok := ds.loaded[name]; !ok {
				ds.loaded[name] = &LoadedTool{
					ToolEntry:  ds.known[i],
					LoadedAt:   time.Now(),
					LastUsedAt: time.Now(),
					UseCount:   0,
					BinaryLive: false,
					AutoLoaded: true,
				}
			}
		}
	}
}

func (ds *DecisionSystem) recordEvent(event DecisionEvent) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.recordEventLocked(event)
}

func (ds *DecisionSystem) recordEventLocked(event DecisionEvent) {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	ds.events[ds.eventIdx%ds.maxEvents] = event
	ds.eventIdx++
}

func (ds *DecisionSystem) requiresBinary(server string) bool {
	return server != "builtin" && server != "tormentnexus"
}

func (ds *DecisionSystem) estimateLatency(server string) string {
	switch server {
	case "builtin", "tormentnexus":
		return "<1ms"
	default:
		return "50-500ms (spawn on first call)"
	}
}

func (ds *DecisionSystem) truncateDescription(desc string, maxLen int) string {
	if len(desc) <= maxLen {
		return desc
	}
	return desc[:maxLen-3] + "..."
}

func (ds *DecisionSystem) generateExample(r RankedTool) string {
	if r.OriginalName == "" {
		return ""
	}
	switch {
	case strings.Contains(r.OriginalName, "search"):
		return fmt.Sprintf(`search_and_call("%s", {"query": "example"})`, r.OriginalName)
	case strings.Contains(r.OriginalName, "read") || strings.Contains(r.OriginalName, "file"):
		return fmt.Sprintf(`call_tool("%s", {"path": "/path/to/file"})`, r.AdvertisedName)
	default:
		return fmt.Sprintf(`call_tool("%s", {})`, r.AdvertisedName)
	}
}

// ---------- Builtin Tool Definitions ----------

// BuiltinTools returns the set of always-available filesystem and shell tools
// that every model expects, modeled after Claude Code, Codex, Gemini CLI, etc.
func BuiltinTools() []ToolEntry {
	return []ToolEntry{
		{Name: "bash", OriginalName: "bash", Server: "tormentnexus", AdvertisedName: "tormentnexus__bash",
			Description: "Execute a bash command in the working directory. Returns stdout, stderr, and exit code.",
			AlwaysOn:    true},
		{Name: "read_file", OriginalName: "read_file", Server: "tormentnexus", AdvertisedName: "tormentnexus__read_file",
			Description: "Read the contents of a file. Supports text files with offset/limit for large files.",
			AlwaysOn:    true},
		{Name: "write_file", OriginalName: "write_file", Server: "tormentnexus", AdvertisedName: "tormentnexus__write_file",
			Description: "Write content to a file. Creates the file and parent directories if needed.",
			AlwaysOn:    true},
		{Name: "edit_file", OriginalName: "edit_file", Server: "tormentnexus", AdvertisedName: "tormentnexus__edit_file",
			Description: "Make precise edits to a file using exact text replacement. Multiple disjoint edits in one call.",
			AlwaysOn:    true},
		{Name: "search_files", OriginalName: "search_files", Server: "tormentnexus", AdvertisedName: "tormentnexus__search_files",
			Description: "Search file contents for a regex pattern. Returns matching lines with file paths and line numbers.",
			AlwaysOn:    true},
		{Name: "list_directory", OriginalName: "list_directory", Server: "tormentnexus", AdvertisedName: "tormentnexus__list_directory",
			Description: "List directory contents sorted alphabetically. Returns entries with type indicators.",
			AlwaysOn:    true},
		{Name: "find_files", OriginalName: "find_files", Server: "tormentnexus", AdvertisedName: "tormentnexus__find_files",
			Description: "Search for files by glob pattern. Returns matching file paths relative to the search directory.",
			AlwaysOn:    true},

		// Codex-compatible aliases
		{Name: "shell", OriginalName: "shell", Server: "tormentnexus", AdvertisedName: "tormentnexus__shell",
			Description: "Execute a shell command. Codex-compatible alias for bash.",
			AlwaysOn:    true},
		{Name: "codex_apply_patch", OriginalName: "apply_patch", Server: "tormentnexus", AdvertisedName: "tormentnexus__apply_patch",
			Description: "Apply a unified diff patch to a file. Codex-compatible.",
			AlwaysOn:    true},

		// Claude Code-compatible aliases
		{Name: "cat", OriginalName: "cat", Server: "tormentnexus", AdvertisedName: "tormentnexus__cat",
			Description: "Display file contents. Claude Code-compatible alias for read_file.",
			AlwaysOn:    true},
		{Name: "sed", OriginalName: "sed", Server: "tormentnexus", AdvertisedName: "tormentnexus__sed",
			Description: "Stream editor for file transformations. Claude Code-compatible.",
			AlwaysOn:    true},
		{Name: "grep", OriginalName: "grep", Server: "tormentnexus", AdvertisedName: "tormentnexus__grep",
			Description: "Search file contents for patterns. Claude Code-compatible alias for search_files.",
			AlwaysOn:    true},
		{Name: "ls", OriginalName: "ls", Server: "tormentnexus", AdvertisedName: "tormentnexus__ls",
			Description: "List directory contents. Claude Code-compatible alias for list_directory.",
			AlwaysOn:    true},
		{Name: "find", OriginalName: "find", Server: "tormentnexus", AdvertisedName: "tormentnexus__find",
			Description: "Find files by pattern. Claude Code-compatible alias for find_files.",
			AlwaysOn:    true},

		// Gemini CLI-compatible
		{Name: "run_command", OriginalName: "run_command", Server: "tormentnexus", AdvertisedName: "tormentnexus__run_command",
			Description: "Execute a command and return output. Gemini CLI-compatible.",
			AlwaysOn:    true},
		{Name: "read_many_files", OriginalName: "read_many_files", Server: "tormentnexus", AdvertisedName: "tormentnexus__read_many_files",
			Description: "Read multiple files at once. Gemini CLI-compatible.",
			AlwaysOn:    true},

		// Copilot CLI-compatible
		{Name: "execute_command", OriginalName: "execute_command", Server: "tormentnexus", AdvertisedName: "tormentnexus__execute_command",
			Description: "Execute a shell command with optional timeout. Copilot CLI-compatible.",
			AlwaysOn:    true},
		{Name: "get_file_content", OriginalName: "get_file_content", Server: "tormentnexus", AdvertisedName: "tormentnexus__get_file_content",
			Description: "Get the content of a file. Copilot CLI-compatible.",
			AlwaysOn:    true},

		// Cursor-compatible
		{Name: "codebase_search", OriginalName: "codebase_search", Server: "tormentnexus", AdvertisedName: "tormentnexus__codebase_search",
			Description: "Semantic code search across the codebase. Cursor-compatible.",
			AlwaysOn:    true},
		{Name: "read_file_block", OriginalName: "read_file_block", Server: "tormentnexus", AdvertisedName: "tormentnexus__read_file_block",
			Description: "Read a specific line range from a file. Cursor-compatible.",
			AlwaysOn:    true},

		// Meta-tools (the permanent decision system surface)
		{Name: "search_tools", OriginalName: "search_tools", Server: "tormentnexus", AdvertisedName: "tormentnexus__search_tools",
			Description: "Search the tool catalog by keyword. Returns ranked, compact results with match reasons.",
			AlwaysOn:    true},
		{Name: "call_tool", OriginalName: "call_tool", Server: "tormentnexus", AdvertisedName: "tormentnexus__call_tool",
			Description: "Call any tool by name. Auto-loads if not yet loaded. Combined search-and-call.",
			AlwaysOn:    true},
		{Name: "load_tool", OriginalName: "load_tool", Server: "tormentnexus", AdvertisedName: "tormentnexus__load_tool",
			Description: "Load a tool into the active working set. Required before calling tools not yet loaded.",
			AlwaysOn:    true},
		{Name: "list_loaded_tools", OriginalName: "list_loaded_tools", Server: "tormentnexus", AdvertisedName: "tormentnexus__list_loaded_tools",
			Description: "List all currently loaded tools with usage stats.",
			AlwaysOn:    true},
		{Name: "unload_tool", OriginalName: "unload_tool", Server: "tormentnexus", AdvertisedName: "tormentnexus__unload_tool",
			Description: "Unload a tool from the active working set to free context.",
			AlwaysOn:    true},

		// Repograph-native tools
		{Name: "repograph_build", OriginalName: "repograph_build", Server: "tormentnexus", AdvertisedName: "tormentnexus__repograph_build",
			Description: "Build or rebuild the repository dependency graph. Triggers a full scan of source files.",
			AlwaysOn:    true},
		{Name: "repograph_get", OriginalName: "repograph_get", Server: "tormentnexus", AdvertisedName: "tormentnexus__repograph_get",
			Description: "Get the current repository graph structure and statistics.",
			AlwaysOn:    true},
		{Name: "repograph_find_references", OriginalName: "repograph_find_references", Server: "tormentnexus", AdvertisedName: "tormentnexus__repograph_find_references",
			Description: "Find all references to a specific code symbol (function, type, interface) in the repository.",
			AlwaysOn:    true},
		{Name: "repograph_find_dependents", OriginalName: "repograph_find_dependents", Server: "tormentnexus", AdvertisedName: "tormentnexus__repograph_find_dependents",
			Description: "Find all files that depend on or import a given source file.",
			AlwaysOn:    true},
		{Name: "repograph_search", OriginalName: "repograph_search", Server: "tormentnexus", AdvertisedName: "tormentnexus__repograph_search",
			Description: "Search for symbols across the repository using the native Go repograph engine.",
			AlwaysOn:    true},
	}
}
