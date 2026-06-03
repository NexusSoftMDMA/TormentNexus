package httpapi

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/harnesses"
	"github.com/tormentnexushq/tormentnexus-go/internal/mcp"
)

type localMCPInventoryView struct {
	Inventory                   *mcp.Inventory
	CachePath                   string
	CachePresent                bool
	InventorySource             string
	CachedAt                    string
	PersistedOverlayCheckedAt   string
	LiveOverlayCheckedAt        string
	PersistedOverlayServerCount int
	PersistedOverlayToolCount   int
	RuntimeOverlayServerCount   int
	RuntimeOverlayToolCount     int
	ServerSources               map[string]string
	PersistedOverlayRecords     map[string]mcp.RuntimeOverlayServer
	LiveOverlayRecords          map[string]mcp.RuntimeOverlayServer
}

func (s *Server) localMCPInventoryCachePath() string {
	return filepath.Join(s.cfg.ConfigDir, "mcp_inventory_cache.json")
}

func (s *Server) localMCPInventory() (*mcp.Inventory, error) {
	return mcp.LoadInventoryWithCache(s.cfg.WorkspaceRoot, s.cfg.MainConfigDir, s.localMCPInventoryCachePath())
}

func (s *Server) syncRuntimeOverlayCache() error {
	return mcp.SyncRuntimeOverlayCache(s.localMCPInventoryCachePath(), runtimeOverlayServersFromRecords(s.runtimeServers.list()))
}

func (s *Server) localMCPInventoryView() (*localMCPInventoryView, error) {
	inventory, err := s.localMCPInventory()
	if err != nil {
		return nil, err
	}
	cachePath := s.localMCPInventoryCachePath()
	_, statErr := os.Stat(cachePath)
	view := &localMCPInventoryView{
		Inventory:               cloneInventory(inventory),
		CachePath:               cachePath,
		CachePresent:            statErr == nil,
		InventorySource:         inventory.Source,
		CachedAt:                inventory.CachedAt,
		ServerSources:           map[string]string{},
		PersistedOverlayRecords: map[string]mcp.RuntimeOverlayServer{},
		LiveOverlayRecords:      map[string]mcp.RuntimeOverlayServer{},
	}
	for _, server := range view.Inventory.Servers {
		view.ServerSources[server.Name] = inventory.Source
	}
	if snapshot, snapshotErr := mcp.LoadInventoryCacheSnapshot(cachePath); snapshotErr == nil && snapshot != nil {
		view.applyPersistedRuntimeOverlay(snapshot.RuntimeOverlay)
		if len(snapshot.RuntimeOverlay) > 0 {
			view.PersistedOverlayCheckedAt = freshestRuntimeOverlayCheck(snapshot.RuntimeOverlay, snapshot.CachedAt)
		}
		if view.CachedAt == "" {
			view.CachedAt = snapshot.CachedAt
		}
	}
	liveOverlay := runtimeOverlayServersFromRecords(s.runtimeServers.list())
	view.LiveOverlayCheckedAt = freshestRuntimeOverlayCheck(liveOverlay, "")
	view.applyRuntimeOverlay(liveOverlay)
	return view, nil
}

func (v *localMCPInventoryView) applyPersistedRuntimeOverlay(records []mcp.RuntimeOverlayServer) {
	v.applyOverlay(records, true)
}

func (v *localMCPInventoryView) applyRuntimeOverlay(records []mcp.RuntimeOverlayServer) {
	v.applyOverlay(records, false)
}

func (v *localMCPInventoryView) applyOverlay(records []mcp.RuntimeOverlayServer, persisted bool) {
	if v == nil || v.Inventory == nil || len(records) == 0 {
		return
	}
	serverIndex := make(map[string]int, len(v.Inventory.Servers))
	toolIndex := make(map[string]int, len(v.Inventory.Tools))
	for i, server := range v.Inventory.Servers {
		serverIndex[server.Name] = i
	}
	for i, tool := range v.Inventory.Tools {
		toolKey := tool.Server + "::" + normalizedInventoryToolName(tool)
		toolIndex[toolKey] = i
	}

	for _, record := range records {
		if strings.TrimSpace(record.Name) == "" || len(record.Tools) == 0 {
			continue
		}
		if index, ok := serverIndex[record.Name]; ok {
			server := v.Inventory.Servers[index]
			server.Command = record.Command
			server.Args = append([]string(nil), record.Args...)
			server.Env = copyStringMap(record.Env)
			server.Enabled = true
			v.Inventory.Servers[index] = server
		} else {
			v.Inventory.Servers = append(v.Inventory.Servers, mcp.ServerEntry{
				UUID:        "runtime:" + record.Name,
				Name:        record.Name,
				DisplayName: record.Name,
				Type:        "STDIO",
				Command:     record.Command,
				Args:        append([]string(nil), record.Args...),
				Env:         copyStringMap(record.Env),
				Enabled:     true,
			})
			serverIndex[record.Name] = len(v.Inventory.Servers) - 1
			if persisted {
				v.PersistedOverlayServerCount++
			} else {
				v.RuntimeOverlayServerCount++
			}
		}
		source := record.Source
		if persisted {
			source = "go-persisted-runtime-overlay"
			v.PersistedOverlayRecords[record.Name] = record
		} else {
			v.LiveOverlayRecords[record.Name] = record
		}
		v.ServerSources[record.Name] = source
		for _, tool := range record.Tools {
			name := strings.TrimSpace(tool.Name)
			if name == "" {
				continue
			}
			entry := mcp.ToolEntryFromMetadata(record.Name, tool)
			toolKey := record.Name + "::" + name
			if index, ok := toolIndex[toolKey]; ok {
				v.Inventory.Tools[index] = entry
			} else {
				v.Inventory.Tools = append(v.Inventory.Tools, entry)
				toolIndex[toolKey] = len(v.Inventory.Tools) - 1
				if persisted {
					v.PersistedOverlayToolCount++
				} else {
					v.RuntimeOverlayToolCount++
				}
			}
		}
	}
}

func runtimeOverlayServersFromRecords(records []runtimeServerRecord) []mcp.RuntimeOverlayServer {
	if len(records) == 0 {
		return []mcp.RuntimeOverlayServer{}
	}
	overlay := make([]mcp.RuntimeOverlayServer, 0, len(records))
	for _, record := range records {
		if strings.TrimSpace(record.Name) == "" || len(record.Tools) == 0 {
			continue
		}
		overlay = append(overlay, mcp.RuntimeOverlayServer{
			Name:                record.Name,
			Command:             record.Command,
			Args:                append([]string(nil), record.Args...),
			Env:                 copyStringMap(record.Env),
			RuntimeConnected:    record.RuntimeConnected,
			ToolCount:           record.ToolCount,
			ToolInventoryStatus: record.ToolInventoryStatus,
			IntegrationLevel:    record.IntegrationLevel,
			Source:              record.Source,
			Tools:               mcp.MetadataToolsFromAny(genericAnySlice(record.Tools)),
			LastCheckedAt:       record.LastCheckedAt,
			LastError:           record.LastError,
		})
	}
	return overlay
}

func cloneInventory(inventory *mcp.Inventory) *mcp.Inventory {
	if inventory == nil {
		return &mcp.Inventory{Servers: []mcp.ServerEntry{}, Tools: []mcp.ToolEntry{}, Source: "empty"}
	}
	return &mcp.Inventory{
		Servers:  append([]mcp.ServerEntry(nil), inventory.Servers...),
		Tools:    append([]mcp.ToolEntry(nil), inventory.Tools...),
		Source:   inventory.Source,
		CachedAt: inventory.CachedAt,
	}
}

func normalizedInventoryToolName(tool mcp.ToolEntry) string {
	if strings.TrimSpace(tool.OriginalName) != "" {
		return tool.OriginalName
	}
	if strings.TrimSpace(tool.Name) != "" {
		return tool.Name
	}
	if strings.TrimSpace(tool.AdvertisedName) != "" {
		return tool.AdvertisedName
	}
	return "unknown"
}

func normalizedInventoryTools(inventory *mcp.Inventory) []mcp.ToolEntry {
	if inventory == nil || len(inventory.Tools) == 0 {
		return []mcp.ToolEntry{}
	}
	tools := make([]mcp.ToolEntry, 0, len(inventory.Tools))
	for _, tool := range inventory.Tools {
		copyTool := tool
		copyTool.Name = normalizedInventoryToolName(tool)
		if strings.TrimSpace(copyTool.AdvertisedName) == "" {
			copyTool.AdvertisedName = tool.Name
		}
		tools = append(tools, copyTool)
	}
	return tools
}

func inventoryBridgeMeta(view *localMCPInventoryView) map[string]any {
	if view == nil {
		return map[string]any{}
	}
	meta := map[string]any{
		"inventorySource":             view.InventorySource,
		"cachedAt":                    nullableString(view.CachedAt),
		"cachePath":                   view.CachePath,
		"cachePresent":                view.CachePresent,
		"cacheAuthority":              "go-local-live-sync",
		"metadataAuthority":           "mcp.jsonc",
		"persistedOverlayServerCount": view.PersistedOverlayServerCount,
		"persistedOverlayToolCount":   view.PersistedOverlayToolCount,
		"runtimeOverlayServerCount":   view.RuntimeOverlayServerCount,
		"runtimeOverlayToolCount":     view.RuntimeOverlayToolCount,
	}
	for key, value := range freshnessBridgeMeta("baseInventory", view.CachedAt, 24*time.Hour) {
		meta[key] = value
	}
	for key, value := range freshnessBridgeMeta("persistedOverlay", view.PersistedOverlayCheckedAt, 15*time.Minute) {
		meta[key] = value
	}
	for key, value := range freshnessBridgeMeta("liveOverlay", view.LiveOverlayCheckedAt, 15*time.Minute) {
		meta[key] = value
	}
	return meta
}

func freshestRuntimeOverlayCheck(records []mcp.RuntimeOverlayServer, fallback string) string {
	freshest := strings.TrimSpace(fallback)
	freshestTime := parseFreshnessTimestamp(freshest)
	for _, record := range records {
		candidate := strings.TrimSpace(record.LastCheckedAt)
		candidateTime := parseFreshnessTimestamp(candidate)
		if candidateTime.IsZero() {
			continue
		}
		if freshestTime.IsZero() || candidateTime.After(freshestTime) {
			freshest = candidate
			freshestTime = candidateTime
		}
	}
	return freshest
}

func freshnessBridgeMeta(prefix string, timestamp string, staleAfter time.Duration) map[string]any {
	meta := map[string]any{
		prefix + "CachedAt": nullableString(timestamp),
	}
	parsed := parseFreshnessTimestamp(timestamp)
	if parsed.IsZero() {
		meta[prefix+"AgeMs"] = nil
		meta[prefix+"StaleHeuristic"] = nil
		return meta
	}
	age := time.Since(parsed)
	if age < 0 {
		age = 0
	}
	meta[prefix+"AgeMs"] = age.Milliseconds()
	meta[prefix+"StaleHeuristic"] = age > staleAfter
	return meta
}

func parseFreshnessTimestamp(value string) time.Time {
	if strings.TrimSpace(value) == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return parsed.UTC()
}

func genericAnySlice(items []map[string]any) []any {
	if len(items) == 0 {
		return []any{}
	}
	result := make([]any, 0, len(items))
	for _, item := range items {
		result = append(result, item)
	}
	return result
}

func inventoryToolSource(view *localMCPInventoryView, serverName string) string {
	if view == nil {
		return "unknown"
	}
	if source, ok := view.ServerSources[serverName]; ok && strings.TrimSpace(source) != "" {
		return source
	}
	return view.InventorySource
}

func inventoryToolLayerMeta(view *localMCPInventoryView, serverName string) map[string]any {
	return inventoryLayerMeta(view, inventoryToolSource(view, serverName), "base-inventory")
}

func inventoryServerLayerMeta(view *localMCPInventoryView, source string) map[string]any {
	return inventoryLayerMeta(view, source, "source-backed-summary")
}

func inventoryLayerMeta(view *localMCPInventoryView, source string, defaultLayer string) map[string]any {
	layer := defaultLayer
	timestamp := view.CachedAt
	staleAfter := 24 * time.Hour
	switch source {
	case "go-persisted-runtime-overlay":
		layer = "persisted-runtime-overlay"
		timestamp = view.PersistedOverlayCheckedAt
		staleAfter = 15 * time.Minute
	case "go-runtime-registry":
		layer = "live-runtime-overlay"
		timestamp = view.LiveOverlayCheckedAt
		staleAfter = 15 * time.Minute
	case "source-backed":
		layer = defaultLayer
		timestamp = ""
	}
	mirrors := map[string]any{"originLayer": layer}
	for key, value := range freshnessBridgeMeta("layer", timestamp, staleAfter) {
		mirrors[key] = value
	}
	mirrors["provenance"] = newRecordProvenanceObject(layer, source, nullableString(timestamp), mirrors["layerAgeMs"], mirrors["layerStaleHeuristic"], []string{"originLayer", "layerCachedAt", "layerAgeMs", "layerStaleHeuristic"})
	return mirrors
}

func newRecordProvenanceObject(layer string, source string, cachedAt any, ageMs any, staleHeuristic any, legacyMirrorFields []string) map[string]any {
	return map[string]any{
		"schemaVersion":      1,
		"primary":            true,
		"compatibilityMode":  "legacy-top-level-mirrors-retained",
		"legacyMirrorFields": append([]string(nil), legacyMirrorFields...),
		"layer":              layer,
		"source":             source,
		"cachedAt":           cachedAt,
		"ageMs":              ageMs,
		"staleHeuristic":     staleHeuristic,
		"cacheAuthority":     "go-local-live-sync",
		"metadataAuthority":  "mcp.jsonc",
	}
}

func fallbackMCPInventoryTools(view *localMCPInventoryView) []map[string]any {
	if view == nil {
		return []map[string]any{}
	}
	tools := normalizedInventoryTools(view.Inventory)
	result := make([]map[string]any, 0, len(tools))
	for _, tool := range tools {
		item := map[string]any{
			"name":              tool.Name,
			"description":       tool.Description,
			"server":            tool.Server,
			"alwaysShow":        tool.AlwaysOn,
			"source":            inventoryToolSource(view, tool.Server),
			"availability":      "cache-backed",
			"inventoryCachedAt": nullableString(view.CachedAt),
		}
		for key, value := range inventoryToolLayerMeta(view, tool.Server) {
			item[key] = value
		}
		result = append(result, primaryProvenanceOnly(item))
	}
	return result
}

func fallbackSearchMCPInventoryTools(query string, view *localMCPInventoryView, limit int) []map[string]any {
	if view == nil {
		return []map[string]any{}
	}
	ranked := mcp.RankTools(query, normalizedInventoryTools(view.Inventory), limit)
	results := make([]map[string]any, 0, len(ranked))
	for _, item := range ranked {
		result := map[string]any{
			"name":              normalizedInventoryToolName(item.ToolEntry),
			"server":            item.Server,
			"alwaysShow":        item.AlwaysOn,
			"matchReason":       item.MatchReason,
			"score":             item.Score,
			"source":            inventoryToolSource(view, item.Server),
			"inventoryCachedAt": nullableString(view.CachedAt),
		}
		for key, value := range inventoryToolLayerMeta(view, item.Server) {
			result[key] = value
		}
		results = append(results, primaryProvenanceOnly(result))
	}
	return results
}

func primaryProvenanceOnly(record map[string]any) map[string]any {
	if record == nil {
		return map[string]any{}
	}
	result := cloneMap(record)
	delete(result, "originLayer")
	delete(result, "layerCachedAt")
	delete(result, "layerAgeMs")
	delete(result, "layerStaleHeuristic")
	if provenance, ok := result["provenance"].(map[string]any); ok {
		cloned := cloneMap(provenance)
		cloned["compatibilityMode"] = "legacy-top-level-mirrors-trimmed"
		cloned["legacyMirrorFields"] = []string{}
		result["provenance"] = cloned
	}
	return result
}

func primaryConfiguredServerProvenanceOnly(record map[string]any) map[string]any {
	if record == nil {
		return map[string]any{}
	}
	result := cloneMap(record)
	delete(result, "originLayer")
	delete(result, "metadataOrigin")
	delete(result, "metadataCachedAt")
	delete(result, "metadataAgeMs")
	delete(result, "metadataStaleHeuristic")
	if provenance, ok := result["provenance"].(map[string]any); ok {
		cloned := cloneMap(provenance)
		cloned["compatibilityMode"] = "legacy-top-level-mirrors-trimmed"
		cloned["legacyMirrorFields"] = []string{}
		result["provenance"] = cloned
	}
	return result
}

func fallbackControlToolsFromInventory(view *localMCPInventoryView) []map[string]any {
	if view == nil || view.Inventory == nil {
		return []map[string]any{}
	}
	serverUUIDs := make(map[string]string, len(view.Inventory.Servers))
	for _, server := range view.Inventory.Servers {
		if strings.TrimSpace(server.Name) != "" {
			serverUUIDs[server.Name] = server.UUID
		}
	}
	tools := normalizedInventoryTools(view.Inventory)
	result := make([]map[string]any, 0, len(tools))
	for _, tool := range tools {
		schemaParamCount := 0
		if inputSchemaMap, ok := tool.InputSchema.(map[string]any); ok {
			if properties, ok := inputSchemaMap["properties"].(map[string]any); ok {
				schemaParamCount = len(properties)
			}
		}
		item := map[string]any{
			"uuid":              tool.Name,
			"name":              tool.Name,
			"description":       nullableString(tool.Description),
			"server":            tool.Server,
			"inputSchema":       tool.InputSchema,
			"isDeferred":        false,
			"schemaParamCount":  schemaParamCount,
			"mcpServerUuid":     serverUUIDs[tool.Server],
			"always_on":         tool.AlwaysOn,
			"source":            inventoryToolSource(view, tool.Server),
			"inventoryCachedAt": nullableString(view.CachedAt),
		}
		for key, value := range inventoryToolLayerMeta(view, tool.Server) {
			item[key] = value
		}
		result = append(result, item)
	}
	return result
}

func fallbackControlToolListFromInventory(view *localMCPInventoryView) []map[string]any {
	items := fallbackControlToolsFromInventory(view)
	trimmed := make([]map[string]any, 0, len(items))
	for _, item := range items {
		trimmed = append(trimmed, primaryProvenanceOnly(item))
	}
	return trimmed
}

func fallbackControlToolFromInventory(view *localMCPInventoryView, uuid string) any {
	for _, tool := range fallbackControlToolsFromInventory(view) {
		if stringValue(tool["uuid"]) == uuid || stringValue(tool["name"]) == uuid {
			return tool
		}
	}
	return nil
}

func fallbackControlToolPrimaryProvenance(view *localMCPInventoryView, uuid string) any {
	tool := fallbackControlToolFromInventory(view, uuid)
	toolMap, _ := tool.(map[string]any)
	if toolMap == nil {
		return nil
	}
	return primaryProvenanceOnly(toolMap)
}

func fallbackRuntimeServerListWithPrimaryProvenance(definitions []harnesses.Definition, view *localMCPInventoryView) []map[string]any {
	items := fallbackRuntimeServersWithProvenance(definitions, view)
	trimmed := make([]map[string]any, 0, len(items))
	for _, item := range items {
		trimmed = append(trimmed, primaryProvenanceOnly(item))
	}
	return trimmed
}

func fallbackRuntimeServersWithProvenance(definitions []harnesses.Definition, view *localMCPInventoryView) []map[string]any {
	servers := make([]map[string]any, 0)
	seen := map[string]struct{}{}
	for _, definition := range sourceBackedInstalledHarnesses(definitions) {
		item := map[string]any{
			"name":                definition.ID,
			"runtimeConnected":    false,
			"toolCount":           definition.ToolCallCount,
			"toolInventoryStatus": definition.ToolInventoryStatus,
			"integrationLevel":    definition.IntegrationLevel,
			"source":              definition.ToolSource,
		}
		for key, value := range inventoryServerLayerMeta(view, "source-backed") {
			item[key] = value
		}
		servers = append(servers, item)
		seen[definition.ID] = struct{}{}
	}
	if view != nil {
		names := make([]string, 0, len(view.PersistedOverlayRecords)+len(view.LiveOverlayRecords))
		for name := range view.PersistedOverlayRecords {
			names = append(names, name)
		}
		for name := range view.LiveOverlayRecords {
			if _, ok := view.PersistedOverlayRecords[name]; !ok {
				names = append(names, name)
			}
		}
		sort.Strings(names)
		for _, name := range names {
			record, source := view.PersistedOverlayRecords[name], "go-persisted-runtime-overlay"
			if live, ok := view.LiveOverlayRecords[name]; ok {
				record = live
				source = "go-runtime-registry"
			}
			if strings.TrimSpace(record.Name) == "" {
				continue
			}
			if _, exists := seen[record.Name]; exists && source == "source-backed" {
				continue
			}
			item := map[string]any{
				"name":                record.Name,
				"runtimeConnected":    record.RuntimeConnected,
				"toolCount":           record.ToolCount,
				"toolInventoryStatus": record.ToolInventoryStatus,
				"integrationLevel":    record.IntegrationLevel,
				"source":              source,
				"command":             record.Command,
				"args":                record.Args,
				"lastCheckedAt":       nullableString(record.LastCheckedAt),
				"lastError":           nullableString(record.LastError),
			}
			for key, value := range inventoryServerLayerMeta(view, source) {
				item[key] = value
			}
			servers = append(servers, item)
			seen[record.Name] = struct{}{}
		}
	}
	return servers
}
