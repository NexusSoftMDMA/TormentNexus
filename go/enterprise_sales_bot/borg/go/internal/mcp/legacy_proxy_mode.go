package mcp

import (
	"encoding/json"
	"fmt"
	"strings"
)

// LegacyProxyMode handles the backward-compatible proxy mode for MCP tool calls
// that were made through the older TS bridge pattern.
type LegacyProxyMode struct {
	inventory *CachedInventory
}

// NewLegacyProxyMode creates a new legacy proxy mode handler.
func NewLegacyProxyMode(inventory *CachedInventory) *LegacyProxyMode {
	return &LegacyProxyMode{
		inventory: inventory,
	}
}

// LegacyToolRequest represents a tool request in legacy format.
type LegacyToolRequest struct {
	Tool   string                 `json:"tool"`
	Server string                 `json:"server,omitempty"`
	Args   map[string]interface{} `json:"args"`
}

// IsLegacyFormat checks whether a JSON payload uses the legacy tool format.
func IsLegacyFormat(payload json.RawMessage) bool {
	var req LegacyToolRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		return false
	}
	return req.Tool != "" && !strings.Contains(req.Tool, "__")
}

// ConvertLegacyRequest converts a legacy tool request to the current format.
func (lpm *LegacyProxyMode) ConvertLegacyRequest(req LegacyToolRequest) (*ToolCallRequest, error) {
	if req.Server != "" {
		return &ToolCallRequest{
			Name:      NamespaceToolName(req.Server, req.Tool),
			Arguments: req.Args,
		}, nil
	}

	// Try to resolve bare tool name
	tools := lpm.inventory.FindTools("", req.Tool)
	if len(tools) == 0 {
		return nil, fmt.Errorf("could not resolve tool '%s' to any server", req.Tool)
	}

	return &ToolCallRequest{
		Name:      tools[0].Name,
		Arguments: req.Args,
	}, nil
}

// ToolCallRequest is the standard format for MCP tool calls.
type ToolCallRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolCallResult is the standard format for MCP tool call results.
type ToolCallResult struct {
	Content []ToolResultContent `json:"content"`
	IsError bool                `json:"isError"`
}

// ToolResultContent holds a single content item in a tool result.
type ToolResultContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// WrapLegacyResponse wraps a result in a legacy-compatible response format.
func WrapLegacyResponse(result *ToolCallResult) map[string]interface{} {
	return map[string]interface{}{
		"content": result.Content,
		"isError": result.IsError,
	}
}
