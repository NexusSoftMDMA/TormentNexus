//go:build ignore
// +build ignore

package tools

import (
	"context"
)

func HandleList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	content := `Awesome MCP 资源精选 (中文版)

## MCP Servers
- Filesystem (modelcontextprotocol)
- GitHub (modelcontextprotocol)
- PostgreSQL (modelcontextprotocol)
- Playwright (microsoft)

## MCP Clients
- Claude Desktop
- Continue.dev
- Goose (block)

## MCP 指南
- MCP 概览 (modelcontextprotocol)
- MCP 设计文档
- MCP 规范

更多资源请访问: https://github.com/topics/mcp`
	return ok(content)
}
