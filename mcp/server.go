package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// MCPHost represents the internal Model Context Protocol engine of TormentNexus
type MCPHost struct {
	server *server.MCPServer
}

// NewMCPHost creates a native Go MCP server matching the TS implementation
func NewMCPHost() *MCPHost {
	mcpServer := server.NewMCPServer(
		"TormentNexus TormentNexus Core",
		"0.2.0",
	)

	return &MCPHost{
		server: mcpServer,
	}
}

// RegisterNativeTools hooks our purely native Go implementations
// directly into the MCP Server so they bypass IPC routing overhead.
func (h *MCPHost) RegisterNativeTools() {
	// Example stub of native Go tool registration (equivalent to @tormentnexus/tools)

	systemStatusTool := mcp.NewTool("system_status",
		mcp.WithDescription("Get the health and status of the TormentNexus native core."),
	)

	h.server.AddTool(systemStatusTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("TormentNexus Native Core is running smoothly with 100% capacity."), nil
	})

	fmt.Println("[MCP] Native Tools registered directly to host.")
}

// StartStdio launches the server via the standard Stdio transport (CLI mode)
func (h *MCPHost) StartStdio() error {
	return server.ServeStdio(h.server)
}

// We will also implement StartWebSocket for the Fiber bridging later!
