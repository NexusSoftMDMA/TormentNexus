package tools

import (
	"context"
	"encoding/json"
)

func HandleListDevopsMcpServers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	servers := []string{
		"github.com/modelcontextprotocol/servers",
		"github.com/docker/mcp-servers",
		"github.com/kubernetes/mcp-servers",
	}
	data, e := json.Marshal(servers)
	if e != nil {
		return err("failed to marshal server list")
}

	return success(string(data))
}