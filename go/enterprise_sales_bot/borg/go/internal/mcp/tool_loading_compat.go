package mcp

import (
	"fmt"
	"strings"
)

// ToolLoadingCompatibility handles compatibility checks for tool loading
// across different MCP server implementations.
type ToolLoadingCompatibility struct{}

// NewToolLoadingCompatibility creates a new tool loading compatibility handler.
func NewToolLoadingCompatibility() *ToolLoadingCompatibility {
	return &ToolLoadingCompatibility{}
}

// CompatibilityIssue describes a compatibility issue with a tool.
type CompatibilityIssue struct {
	ToolName   string `json:"toolName"`
	Severity   string `json:"severity"` // "error", "warning", "info"
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}

// CheckTool verifies that a tool can be loaded from its server.
func (tlc *ToolLoadingCompatibility) CheckTool(toolName, serverType string) []CompatibilityIssue {
	var issues []CompatibilityIssue

	switch serverType {
	case "STDIO":
		if strings.HasPrefix(toolName, "_") {
			issues = append(issues, CompatibilityIssue{
				ToolName:   toolName,
				Severity:   "warning",
				Message:    "Tool name starts with underscore, may be internal",
				Suggestion: "Verify this tool is intended for external use",
			})
		}
	case "SSE":
		if len(toolName) > 100 {
			issues = append(issues, CompatibilityIssue{
				ToolName:   toolName,
				Severity:   "warning",
				Message:    "Tool name exceeds 100 characters",
				Suggestion: "Shorten tool name for better SSE compatibility",
			})
		}
	case "STREAMABLE_HTTP":
		// Most compatible, few issues
	default:
		issues = append(issues, CompatibilityIssue{
			ToolName:   toolName,
			Severity:   "info",
			Message:    fmt.Sprintf("Unknown server type: %s", serverType),
			Suggestion: "Defaulting to STDIO compatibility checks",
		})
	}

	return issues
}

// CheckSchema verifies that a tool's input schema is compatible.
func (tlc *ToolLoadingCompatibility) CheckSchema(toolName string, inputSchema interface{}) []CompatibilityIssue {
	var issues []CompatibilityIssue

	if inputSchema == nil {
		issues = append(issues, CompatibilityIssue{
			ToolName:   toolName,
			Severity:   "warning",
			Message:    "Tool has no input schema",
			Suggestion: "Add an input schema to enable parameter validation",
		})
	}

	return issues
}

// CheckAll runs all compatibility checks for a tool.
func (tlc *ToolLoadingCompatibility) CheckAll(toolName, serverType string, inputSchema interface{}) []CompatibilityIssue {
	var issues []CompatibilityIssue
	issues = append(issues, tlc.CheckTool(toolName, serverType)...)
	issues = append(issues, tlc.CheckSchema(toolName, inputSchema)...)
	return issues
}
