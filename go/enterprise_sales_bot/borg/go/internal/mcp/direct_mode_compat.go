package mcp

import (
	"context"
	"fmt"
)

// DirectModeCompatibility handles the compatibility layer for direct-mode
// MCP tool calls, where tools are called directly without the progressive router.
type DirectModeCompatibility struct {
	inventory *CachedInventory
}

// DirectToolCall represents a direct MCP tool call.
type DirectToolCall struct {
	ToolName   string                 `json:"toolName"`
	ServerName string                 `json:"serverName"`
	Arguments  map[string]interface{} `json:"arguments"`
}

// DirectToolResult represents the result of a direct MCP tool call.
type DirectToolResult struct {
	Content []DirectContent `json:"content"`
	IsError bool            `json:"isError"`
}

// DirectContent represents content in a direct tool result.
type DirectContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// NewDirectModeCompatibility creates a new direct mode compatibility handler.
func NewDirectModeCompatibility(inventory *CachedInventory) *DirectModeCompatibility {
	return &DirectModeCompatibility{
		inventory: inventory,
	}
}

// ResolveTool resolves a tool name to find its server. Supports both
// namespaced (server__tool) and bare tool names.
func (dmc *DirectModeCompatibility) ResolveTool(namespacedName string) (*DirectToolCall, error) {
	serverName, toolName, ok := ParseNamespacedName(namespacedName)
	if !ok {
		// Try finding the tool across all servers
		tools := dmc.inventory.FindTools("", namespacedName)
		if len(tools) == 0 {
			return nil, fmt.Errorf("tool not found: %s", namespacedName)
		}
		return &DirectToolCall{
			ToolName:   tools[0].OriginalName,
			ServerName: tools[0].Server,
		}, nil
	}
	return &DirectToolCall{
		ToolName:   toolName,
		ServerName: serverName,
	}, nil
}

// ValidateCall checks whether a direct tool call is valid given the current inventory.
func (dmc *DirectModeCompatibility) ValidateCall(ctx context.Context, call *DirectToolCall) error {
	server := dmc.inventory.FindServer(call.ServerName)
	if server == nil {
		return fmt.Errorf("server not found: %s", call.ServerName)
	}
	if !server.Enabled {
		return fmt.Errorf("server is disabled: %s", call.ServerName)
	}
	return nil
}
