package mcp

import (
	"fmt"
	"os"
	"strings"

	"github.com/NexusSoftMDMA/TormentNexus/foundation/adapters"
)

// Client represents a Model Context Protocol client.
type Client struct {
	ServerURL string
	adapter   *adapters.MCPAdapter
}

func NewClient(url string) *Client {
	cwd, _ := os.Getwd()
	return &Client{ServerURL: url, adapter: adapters.NewMCPAdapter(cwd)}
}

func (c *Client) Connect() error {
	status := c.adapter.Status()
	if len(status.Servers) == 0 {
		return fmt.Errorf("no configured MCP servers")
	}
	if strings.TrimSpace(c.ServerURL) == "" {
		c.ServerURL = status.ConfigPath
	}
	fmt.Printf("Connecting to MCP adapter context at %s\n", c.ServerURL)
	return nil
}

func (c *Client) ListTools() ([]string, error) {
	return c.adapter.ListTools()
}

func (c *Client) CallTool(serverName, toolName string, args map[string]interface{}) (string, error) {
	result, err := c.adapter.CallTool(adapters.MCPCallRequest{
		ServerName: serverName,
		ToolName:   toolName,
		Arguments:  args,
	})
	if err != nil {
		return "", err
	}
	return result.Route, nil
}
