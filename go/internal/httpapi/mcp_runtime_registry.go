package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/tormentnexushq/tormentnexus-go/internal/mcp"
)

type runtimeServerRecord struct {
	Name                string            `json:"name"`
	Command             string            `json:"command,omitempty"`
	Args                []string          `json:"args,omitempty"`
	Env                 map[string]string `json:"env,omitempty"`
	RuntimeConnected    bool              `json:"runtimeConnected"`
	ToolCount           int               `json:"toolCount"`
	ToolInventoryStatus string            `json:"toolInventoryStatus"`
	IntegrationLevel    string            `json:"integrationLevel"`
	Source              string            `json:"source"`
	Tools               []map[string]any  `json:"tools,omitempty"`
	LastCheckedAt       string            `json:"lastCheckedAt,omitempty"`
	LastError           string            `json:"lastError,omitempty"`
}

type runtimeServerRegistry struct {
	mu      sync.RWMutex
	records map[string]runtimeServerRecord
}

func newRuntimeServerRegistry() *runtimeServerRegistry {
	return &runtimeServerRegistry{records: map[string]runtimeServerRecord{}}
}

func (r *runtimeServerRegistry) upsert(record runtimeServerRecord) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records[record.Name] = record
}

func (r *runtimeServerRegistry) remove(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.records, name)
}

func (r *runtimeServerRegistry) list() []runtimeServerRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()
	results := make([]runtimeServerRecord, 0, len(r.records))
	for _, record := range r.records {
		results = append(results, record)
	}
	return results
}

func probeRuntimeServer(ctx context.Context, name, command string, args []string, env map[string]string) runtimeServerRecord {
	record := runtimeServerRecord{
		Name:                name,
		Command:             command,
		Args:                append([]string(nil), args...),
		Env:                 copyStringMap(env),
		RuntimeConnected:    false,
		ToolCount:           0,
		ToolInventoryStatus: "unverified",
		IntegrationLevel:    "runtime-added",
		Source:              "go-runtime-registry",
		LastCheckedAt:       time.Now().UTC().Format(time.RFC3339),
	}
	if command == "" {
		record.LastError = "missing command"
		return record
	}

	probeCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	client := mcp.NewStdioClient(name, command, args, env)
	if err := client.Start(); err != nil {
		record.LastError = fmt.Sprintf("failed to start runtime server: %v", err)
		return record
	}
	defer client.Stop()

	resp, err := client.Call(probeCtx, "tools/list", nil)
	if err != nil {
		record.LastError = fmt.Sprintf("tools/list probe failed: %v", err)
		return record
	}
	if resp.Error != nil {
		record.LastError = "tools/list probe returned an MCP error"
		return record
	}

	resultBytes, err := json.Marshal(resp.Result)
	if err != nil {
		record.LastError = fmt.Sprintf("failed to normalize tools/list result: %v", err)
		return record
	}
	var listResult struct {
		Tools []struct {
			Name        string      `json:"name"`
			Description string      `json:"description"`
			InputSchema interface{} `json:"inputSchema"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(resultBytes, &listResult); err != nil {
		record.LastError = fmt.Sprintf("failed to parse tools/list result: %v", err)
		return record
	}

	record.RuntimeConnected = true
	record.ToolInventoryStatus = "live-probed"
	record.ToolCount = len(listResult.Tools)
	record.Tools = make([]map[string]any, 0, len(listResult.Tools))
	for _, tool := range listResult.Tools {
		record.Tools = append(record.Tools, map[string]any{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": tool.InputSchema,
		})
	}
	return record
}

func copyStringMap(input map[string]string) map[string]string {
	if input == nil {
		return map[string]string{}
	}
	copyMap := make(map[string]string, len(input))
	for key, value := range input {
		copyMap[key] = value
	}
	return copyMap
}
