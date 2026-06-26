package mcp

/**
 * @file native_router.go
 * @module go/internal/mcp
 *
 * WHAT: Go-native MCP router implementing the four-layer progressive
 * tool routing strategy with sqlite-vec for semantic search.
 *
 * The four layers are:
 *   L1 — Semantic search: vector-embed the query, find closest tools
 *   L2 — Catalog ranking:  merge with static catalog metadata & scores
 *   L3 — Working set:      load/unload tools, LRU eviction, hydration
 *   L4 — Runtime proxy:    call the actual MCP tool through the transport
 *
 * WHY: Full Assimilation — the Go sidecar must be able to route MCP
 * tool calls independently of the TypeScript core. This router replaces
 * the thin progressive_router.go stub with a complete implementation.
 *
 * ADDED: v1.0.0-alpha.51
 */

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// NativeMCPRouter is the Go-native implementation of the four-layer
// progressive MCP tool routing strategy.
type NativeMCPRouter struct {
	mu sync.RWMutex

	// L1: Semantic index
	decision *DecisionSystem

	// L2: Catalog with rankings
	catalog *ToolCatalog

	// L3: Working set with LRU eviction
	workingSet *WorkingSet

	// L4: Transport layer
	transport MCPTransport

	// Configuration
	cfg RouterConfig

	// Observability
	events []RouterEvent
}

// RouterConfig controls the behavior of the native MCP router.
type RouterConfig struct {
	// MaxWorkingSetSize is the maximum number of tools in the active working set.
	MaxWorkingSetSize int

	// MaxHydratedSchemas is the maximum number of full schemas kept in memory.
	MaxHydratedSchemas int

	// AutoLoadMinConfidence is the minimum confidence score (0-1) for auto-loading.
	AutoLoadMinConfidence float64

	// AlwaysOnTools are tool names that should always be loaded.
	AlwaysOnTools []string

	// SemanticSearchTopK is how many tools to retrieve from semantic search.
	SemanticSearchTopK int
}

// DefaultRouterConfig returns sensible defaults.
func DefaultRouterConfig() RouterConfig {
	return RouterConfig{
		MaxWorkingSetSize:     16,
		MaxHydratedSchemas:    8,
		AutoLoadMinConfidence: 0.85,
		AlwaysOnTools:         []string{},
		SemanticSearchTopK:    10,
	}
}

// WorkingSetEntry is a tool in the active working set.
type WorkingSetEntry struct {
	Name          string    `json:"name"`
	OriginalName  string    `json:"originalName"`
	Server        string    `json:"server"`
	Description   string    `json:"description"`
	InputSchema   any       `json:"inputSchema"`
	AlwaysOn      bool      `json:"alwaysOn"`
	Hydrated      bool      `json:"hydrated"`
	LoadedAt      time.Time `json:"loadedAt"`
	LastUsedAt    time.Time `json:"lastUsedAt"`
	UseCount      int       `json:"useCount"`
	AutoLoaded    bool      `json:"autoLoaded"`
	Confidence    float64   `json:"confidence,omitempty"`
}

// WorkingSet manages the active set of loaded MCP tools with LRU eviction.
type WorkingSet struct {
	mu       sync.RWMutex
	entries  map[string]*WorkingSetEntry
	capacity int
	maxHydrated int
	alwaysOn map[string]bool
}

// NewWorkingSet creates a new working set with the given capacity.
func NewWorkingSet(capacity, maxHydrated int, alwaysOn []string) *WorkingSet {
	ws := &WorkingSet{
		entries:    make(map[string]*WorkingSetEntry),
		capacity:   capacity,
		maxHydrated: maxHydrated,
		alwaysOn:   make(map[string]bool),
	}
	for _, name := range alwaysOn {
		ws.alwaysOn[name] = true
	}
	return ws
}

// Load adds a tool to the working set.
func (ws *WorkingSet) Load(entry WorkingSetEntry) []string {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	entry.LoadedAt = time.Now()
	if entry.LastUsedAt.IsZero() {
		entry.LastUsedAt = time.Now()
	}
	ws.entries[entry.Name] = &entry

	// Evict if over capacity
	evicted := ws.evictIfNeeded()
	return evicted
}

// Unload removes a tool from the working set.
func (ws *WorkingSet) Unload(name string) bool {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if _, ok := ws.entries[name]; !ok {
		return false
	}

	// Never evict always-on tools
	if ws.alwaysOn[name] {
		return false
	}

	delete(ws.entries, name)
	return true
}

// Touch marks a tool as recently used (updates LRU timestamp).
func (ws *WorkingSet) Touch(name string) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if entry, ok := ws.entries[name]; ok {
		entry.LastUsedAt = time.Now()
		entry.UseCount++
	}
}

// Get retrieves a tool from the working set.
func (ws *WorkingSet) Get(name string) (*WorkingSetEntry, bool) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	entry, ok := ws.entries[name]
	if !ok {
		return nil, false
	}
	return entry, true
}

// List returns all entries in the working set.
func (ws *WorkingSet) List() []*WorkingSetEntry {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	result := make([]*WorkingSetEntry, 0, len(ws.entries))
	for _, entry := range ws.entries {
		result = append(result, entry)
	}

	// Sort: always-on first, then by last used
	sort.Slice(result, func(i, j int) bool {
		if result[i].AlwaysOn != result[j].AlwaysOn {
			return result[i].AlwaysOn
		}
		return result[i].LastUsedAt.After(result[j].LastUsedAt)
	})

	return result
}

// HydratedCount returns the number of entries with full schemas.
func (ws *WorkingSet) HydratedCount() int {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	count := 0
	for _, entry := range ws.entries {
		if entry.Hydrated {
			count++
		}
	}
	return count
}

// evictIfNeeded performs LRU eviction when the working set exceeds capacity.
// Returns the names of evicted tools.
func (ws *WorkingSet) evictIfNeeded() []string {
	var evicted []string

	// Evict from working set if over capacity
	for len(ws.entries) > ws.capacity {
		victim := ws.findLRUVictim()
		if victim == "" {
			break
		}
		delete(ws.entries, victim)
		evicted = append(evicted, victim)
	}

	// Dehydrate if too many hydrated schemas
	hydratedCount := 0
	for _, entry := range ws.entries {
		if entry.Hydrated {
			hydratedCount++
		}
	}

	if hydratedCount > ws.maxHydrated {
		// Find least recently used hydrated entries
		type hydratedEntry struct {
			name      string
			lastUsed  time.Time
		}
		var hydrated []hydratedEntry
		for name, entry := range ws.entries {
			if entry.Hydrated && !ws.alwaysOn[name] {
				hydrated = append(hydrated, hydratedEntry{name: name, lastUsed: entry.LastUsedAt})
			}
		}
		sort.Slice(hydrated, func(i, j int) bool {
			return hydrated[i].lastUsed.Before(hydrated[j].lastUsed)
		})

		for i := 0; i < hydratedCount-ws.maxHydrated && i < len(hydrated); i++ {
			if entry, ok := ws.entries[hydrated[i].name]; ok {
				entry.Hydrated = false
				entry.InputSchema = nil
			}
		}
	}

	return evicted
}

// findLRUVictim finds the least recently used tool that is not always-on.
func (ws *WorkingSet) findLRUVictim() string {
	var victim string
	var oldest time.Time

	for name, entry := range ws.entries {
		if ws.alwaysOn[name] {
			continue
		}
		if victim == "" || entry.LastUsedAt.Before(oldest) {
			victim = name
			oldest = entry.LastUsedAt
		}
	}

	return victim
}

// ToolCatalog holds the complete known tool catalog with rankings.
type ToolCatalog struct {
	mu     sync.RWMutex
	tools  map[string]CatalogEntry
}

// CatalogEntry is a tool in the catalog with ranking metadata.
type CatalogEntry struct {
	Name          string   `json:"name"`
	OriginalName  string   `json:"originalName"`
	Server        string   `json:"server"`
	Description   string   `json:"description"`
	Keywords      []string `json:"keywords"`
	AlwaysOn      bool     `json:"alwaysOn"`
	Popularity    float64  `json:"popularity"`
	Relevance     float64  `json:"relevance"`
	Source        string   `json:"source"`
}

// NewToolCatalog creates a new empty tool catalog.
func NewToolCatalog() *ToolCatalog {
	return &ToolCatalog{
		tools: make(map[string]CatalogEntry),
	}
}

// Add inserts or updates a tool in the catalog.
func (tc *ToolCatalog) Add(entry CatalogEntry) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.tools[entry.Name] = entry
}

// Get retrieves a tool from the catalog.
func (tc *ToolCatalog) Get(name string) (CatalogEntry, bool) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	entry, ok := tc.tools[name]
	return entry, ok
}

// List returns all catalog entries.
func (tc *ToolCatalog) List() []CatalogEntry {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	result := make([]CatalogEntry, 0, len(tc.tools))
	for _, entry := range tc.tools {
		result = append(result, entry)
	}
	return result
}

// Size returns the number of tools in the catalog.
func (tc *ToolCatalog) Size() int {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return len(tc.tools)
}

// MCPTransport is the interface for L4 — actually calling MCP tools.
type MCPTransport interface {
	// CallTool invokes an MCP tool on a downstream server.
	CallTool(ctx context.Context, serverName, toolName string, arguments map[string]any) (any, error)

	// ListTools retrieves the tool list from a specific server.
	ListTools(ctx context.Context, serverName string) ([]ToolEntry, error)
}

// RouterEvent records an observability event from the router.
type RouterEvent struct {
	Type      string    `json:"type"` // "search", "load", "unload", "call", "evict", "hydrate"
	Timestamp time.Time `json:"timestamp"`
	ToolName  string    `json:"toolName,omitempty"`
	Query     string    `json:"query,omitempty"`
	Result    string    `json:"result,omitempty"`
	Duration  string    `json:"duration,omitempty"`
	Error     string    `json:"error,omitempty"`
}

// NewNativeMCPRouter creates a new Go-native MCP router.
func NewNativeMCPRouter(decision *DecisionSystem, transport MCPTransport, cfg RouterConfig) *NativeMCPRouter {
	if decision == nil {
		dcfg := DefaultDecisionConfig()
		decision = NewDecisionSystem(dcfg, nil)
		decision.AddCatalogEntries(BuiltinTools())
	}

	return &NativeMCPRouter{
		decision:   decision,
		catalog:    NewToolCatalog(),
		workingSet: NewWorkingSet(cfg.MaxWorkingSetSize, cfg.MaxHydratedSchemas, cfg.AlwaysOnTools),
		transport:  transport,
		cfg:        cfg,
		events:     make([]RouterEvent, 0),
	}
}

// Search executes the L1+L2 routing: semantic search + catalog ranking.
func (nr *NativeMCPRouter) Search(ctx context.Context, query string, profile string) ([]NativeRankedTool, error) {
	start := time.Now()

	// L1: Semantic search via decision system
	results, err := nr.decision.SearchTools(ctx, query)
	if err != nil {
		nr.recordEvent(RouterEvent{
			Type: "search", Query: query, Error: err.Error(), Duration: time.Since(start).String(),
		})
		return nil, fmt.Errorf("L1 semantic search failed: %w", err)
	}

	// L2: Enhance with catalog metadata and re-rank
	ranked := nr.enhanceWithCatalog(query, results, profile)

	nr.recordEvent(RouterEvent{
		Type: "search", Query: query, Result: fmt.Sprintf("%d results", len(ranked)), Duration: time.Since(start).String(),
	})

	return ranked, nil
}

// LoadTool adds a tool to the working set (L3).
func (nr *NativeMCPRouter) LoadTool(ctx context.Context, toolName string) (*WorkingSetEntry, error) {
	start := time.Now()

	// Check if already loaded
	if entry, ok := nr.workingSet.Get(toolName); ok {
		nr.workingSet.Touch(toolName)
		return entry, nil
	}

	// Find the tool in the catalog
	catEntry, ok := nr.catalog.Get(toolName)
	if !ok {
		// Try the decision system
		loaded, err := nr.decision.LoadTool(ctx, toolName)
		if err != nil {
			nr.recordEvent(RouterEvent{Type: "load", ToolName: toolName, Error: err.Error()})
			return nil, fmt.Errorf("tool %q not found in catalog: %w", toolName, err)
		}
		catEntry = CatalogEntry{
			Name:         loaded.Name,
			OriginalName: loaded.OriginalName,
			Server:       loaded.Server,
			Description:  loaded.Description,
			AlwaysOn:     loaded.AlwaysOn,
		}
	}

	entry := WorkingSetEntry{
		Name:         catEntry.Name,
		OriginalName: catEntry.OriginalName,
		Server:       catEntry.Server,
		Description:  catEntry.Description,
		AlwaysOn:     catEntry.AlwaysOn,
		Hydrated:     false,
		LoadedAt:     time.Now(),
		LastUsedAt:   time.Now(),
	}

	evicted := nr.workingSet.Load(entry)
	for _, name := range evicted {
		nr.recordEvent(RouterEvent{Type: "evict", ToolName: name})
	}

	nr.recordEvent(RouterEvent{Type: "load", ToolName: toolName, Duration: time.Since(start).String()})
	return &entry, nil
}

// UnloadTool removes a tool from the working set.
func (nr *NativeMCPRouter) UnloadTool(toolName string) error {
	if !nr.workingSet.Unload(toolName) {
		return fmt.Errorf("tool %q not in working set or is always-on", toolName)
	}
	nr.recordEvent(RouterEvent{Type: "unload", ToolName: toolName})
	return nil
}

// CallTool executes an MCP tool call (L4 — the transport layer).
func (nr *NativeMCPRouter) CallTool(ctx context.Context, toolName string, arguments map[string]any) (any, error) {
	start := time.Now()

	entry, ok := nr.workingSet.Get(toolName)
	if !ok {
		return nil, fmt.Errorf("tool %q not in working set — load it first", toolName)
	}

	if nr.transport == nil {
		return nil, fmt.Errorf("no MCP transport configured for tool execution")
	}

	result, err := nr.transport.CallTool(ctx, entry.Server, entry.OriginalName, arguments)
	if err != nil {
		nr.recordEvent(RouterEvent{Type: "call", ToolName: toolName, Error: err.Error(), Duration: time.Since(start).String()})
		return nil, err
	}

	nr.workingSet.Touch(toolName)
	nr.recordEvent(RouterEvent{Type: "call", ToolName: toolName, Duration: time.Since(start).String()})
	return result, nil
}

// GetWorkingSet returns the current working set entries.
func (nr *NativeMCPRouter) GetWorkingSet() []*WorkingSetEntry {
	return nr.workingSet.List()
}

// GetEvents returns recent router events.
func (nr *NativeMCPRouter) GetEvents(limit int) []RouterEvent {
	nr.mu.RLock()
	defer nr.mu.RUnlock()

	if limit <= 0 || limit > len(nr.events) {
		limit = len(nr.events)
	}

	result := make([]RouterEvent, limit)
	copy(result, nr.events[len(nr.events)-limit:])
	return result
}

// RefreshCatalog refreshes the tool catalog from inventory.
func (nr *NativeMCPRouter) RefreshCatalog(inventory *Inventory) int {
	count := 0
	for _, tool := range inventory.Tools {
		nr.catalog.Add(CatalogEntry{
			Name:         tool.Name,
			OriginalName: tool.OriginalName,
			Server:       tool.Server,
			Description:  tool.Description,
			AlwaysOn:     tool.AlwaysOn,
			Source:       inventory.Source,
		})
		count++
	}

	// Also load always-on tools into the working set
	for _, tool := range inventory.Tools {
		if tool.AlwaysOn {
			nr.workingSet.Load(WorkingSetEntry{
				Name:         tool.Name,
				OriginalName: tool.OriginalName,
				Server:       tool.Server,
				Description:  tool.Description,
				AlwaysOn:     true,
				Hydrated:     false,
				LoadedAt:     time.Now(),
				LastUsedAt:   time.Now(),
			})
		}
	}

	return count
}

// RankedTool is a search result with ranking metadata.
type NativeRankedTool struct {
	Name          string   `json:"name"`
	OriginalName  string   `json:"originalName"`
	Server        string   `json:"server"`
	Description   string   `json:"description"`
	Score         float64  `json:"score"`
	MatchReason   string   `json:"matchReason"`
	Keywords      []string `json:"keywords"`
	AlwaysOn      bool     `json:"alwaysOn"`
	Loaded        bool     `json:"loaded"`
	Hydrated      bool     `json:"hydrated"`
	AutoLoadable  bool     `json:"autoLoadable"`
}

// enhanceWithCatalog merges L1 search results with L2 catalog metadata.
func (nr *NativeMCPRouter) enhanceWithCatalog(query string, results []UnifiedSearchResult, profile string) []NativeRankedTool {
	ranked := make([]NativeRankedTool, 0, len(results))
	loadedNames := nr.workingSet.List()
	loadedSet := make(map[string]bool)
	hydratedSet := make(map[string]bool)
	for _, e := range loadedNames {
		loadedSet[e.Name] = true
		if e.Hydrated {
			hydratedSet[e.Name] = true
		}
	}

	for _, result := range results {
		catEntry, _ := nr.catalog.Get(result.Name)
		keywords := catEntry.Keywords
		if keywords == nil {
			keywords = []string{}
		}

		autoLoadable := result.Score >= nr.cfg.AutoLoadMinConfidence && !loadedSet[result.Name]

		ranked = append(ranked, NativeRankedTool{
			Name:         result.Name,
			OriginalName: result.OriginalName,
			Server:       result.Server,
			Description:  result.Description,
			Score:        result.Score,
			MatchReason:  nr.inferMatchReason(query, result),
			Keywords:     keywords,
			AlwaysOn:     catEntry.AlwaysOn || result.IsAlwaysOn,
			Loaded:       loadedSet[result.Name],
			Hydrated:     hydratedSet[result.Name],
			AutoLoadable: autoLoadable,
		})
	}

	// Sort by score descending
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].Score > ranked[j].Score
	})

	return ranked
}

// inferMatchReason generates a human-readable explanation for why a tool matched.
func (nr *NativeMCPRouter) inferMatchReason(query string, result UnifiedSearchResult) string {
	lowerQuery := strings.ToLower(query)
	desc := strings.ToLower(result.Description)
	name := strings.ToLower(result.Name)

	if strings.Contains(name, lowerQuery) {
		return fmt.Sprintf("tool name matches '%s'", query)
	}
	if strings.Contains(desc, lowerQuery) {
		return fmt.Sprintf("description contains '%s'", query)
	}
	if result.Score >= 0.9 {
		return "high semantic similarity"
	}
	if result.Score >= 0.7 {
		return "moderate semantic similarity"
	}
	return fmt.Sprintf("semantic score %.2f", result.Score)
}

// AutoLoadTools automatically loads high-confidence tools from search results.
func (nr *NativeMCPRouter) AutoLoadTools(ctx context.Context, ranked []NativeRankedTool) []string {
	var loaded []string

	for _, tool := range ranked {
		if !tool.AutoLoadable {
			continue
		}

		entry := WorkingSetEntry{
			Name:         tool.Name,
			OriginalName: tool.OriginalName,
			Server:       tool.Server,
			Description:  tool.Description,
			AlwaysOn:     tool.AlwaysOn,
			AutoLoaded:   true,
			Confidence:   tool.Score,
			LoadedAt:     time.Now(),
			LastUsedAt:   time.Now(),
		}

		evicted := nr.workingSet.Load(entry)
		for _, name := range evicted {
			nr.recordEvent(RouterEvent{Type: "evict", ToolName: name})
		}

		loaded = append(loaded, tool.Name)
		nr.recordEvent(RouterEvent{
			Type:     "load",
			ToolName: tool.Name,
			Result:   fmt.Sprintf("auto-loaded (confidence: %.2f)", tool.Score),
		})
	}

	return loaded
}

// SearchAndAutoLoad combines search + auto-load in one operation.
func (nr *NativeMCPRouter) SearchAndAutoLoad(ctx context.Context, query string, profile string) ([]NativeRankedTool, []string, error) {
	ranked, err := nr.Search(ctx, query, profile)
	if err != nil {
		return nil, nil, err
	}

	loaded := nr.AutoLoadTools(ctx, ranked)

	// Update loaded status in results
	for i := range ranked {
		if _, ok := nr.workingSet.Get(ranked[i].Name); ok {
			ranked[i].Loaded = true
		}
	}

	return ranked, loaded, nil
}

// MarshalState returns the full router state as JSON for diagnostics.
func (nr *NativeMCPRouter) MarshalState() json.RawMessage {
	nr.mu.RLock()
	defer nr.mu.RUnlock()

	wsEntries := nr.workingSet.List()

	state := map[string]any{
		"workingSetSize":   len(wsEntries),
		"catalogSize":      nr.catalog.Size(),
		"hydratedCount":    nr.workingSet.HydratedCount(),
		"eventCount":       len(nr.events),
		"workingSet":       wsEntries,
		"config":           nr.cfg,
	}

	data, _ := json.Marshal(state)
	return data
}

func (nr *NativeMCPRouter) recordEvent(event RouterEvent) {
	nr.mu.Lock()
	defer nr.mu.Unlock()

	event.Timestamp = time.Now()
	nr.events = append(nr.events, event)

	// Keep only last 1000 events
	if len(nr.events) > 1000 {
		nr.events = nr.events[len(nr.events)-1000:]
	}
}
