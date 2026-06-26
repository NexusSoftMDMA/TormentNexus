package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

type Aggregator struct {
	clients map[string]*StdioClient
	mu      sync.RWMutex
}

func NewAggregator() *Aggregator {
	return &Aggregator{
		clients: make(map[string]*StdioClient),
	}
}

func (a *Aggregator) AddServer(name string, command string, args []string, env map[string]string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, ok := a.clients[name]; ok {
		return nil // Already exists
	}

	client := NewStdioClient(name, command, args, env)
	if err := client.Start(); err != nil {
		return err
	}

	a.clients[name] = client
	return nil
}

// ListTools broadcasts tools/list to all connected servers and aggregates the results.
func (a *Aggregator) ListTools(ctx context.Context) ([]ToolEntry, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var allTools []ToolEntry
	for name, client := range a.clients {
		resp, err := client.Call(ctx, "tools/list", nil)
		if err != nil {
			fmt.Printf("[MCP Aggregator] Error listing tools for %s: %v\n", name, err)
			continue
		}

		if resp.Result != nil {
			// MCP protocol returns { "tools": [ { "name": "...", "description": "...", "inputSchema": {...} } ] }
			resultBytes, err := json.Marshal(resp.Result)
			if err != nil {
				continue
			}

			var listResult struct {
				Tools []struct {
					Name        string      `json:"name"`
					Description string      `json:"description"`
					InputSchema interface{} `json:"inputSchema"`
				} `json:"tools"`
			}

			if err := json.Unmarshal(resultBytes, &listResult); err == nil {
				for _, t := range listResult.Tools {
					allTools = append(allTools, ToolEntry{
						Name:              fmt.Sprintf("%s__%s", name, t.Name),
						OriginalName:      t.Name,
						Description:       t.Description,
						Server:            name,
						ServerDisplayName: name,
						AdvertisedName:    fmt.Sprintf("%s__%s", name, t.Name),
						InputSchema:       t.InputSchema,
					})
				}
			}
		}
	}

	return allTools, nil
}

// CallTool routes a tool execution to the appropriate connected server.
func (a *Aggregator) CallTool(ctx context.Context, serverName, toolName string, arguments map[string]interface{}) (*JsonRpcResponse, error) {
	a.mu.RLock()
	client, exists := a.clients[serverName]
	a.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("MCP server '%s' is not connected", serverName)
	}

	params := map[string]interface{}{
		"name":      toolName,
		"arguments": arguments,
	}

	return client.Call(ctx, "tools/call", params)
}

func (a *Aggregator) Shutdown() {
	a.mu.Lock()
	defer a.mu.Unlock()
	for name, client := range a.clients {
		client.Stop()
		delete(a.clients, name)
	}
}
