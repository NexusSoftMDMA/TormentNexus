package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetServers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://mcp-connector.example.com/servers")
	if e != nil {
		return err("failed to fetch servers: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var servers []string
	if e := json.Unmarshal(body, &servers); e != nil {
		return err("failed to parse servers: " + e.Error())
}

	return ok("servers: " + fmt.Sprint(servers))
}

func HandleConnect(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serverName, _ :=getString(args, "server_name")
	if serverName == "" {
		return err("server_name is required")
}

	resp, e := http.DefaultClient.Post("https://mcp-connector.example.com/connect?server="+serverName, "application/json", nil)
	if e != nil {
		return err("failed to connect: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("connection failed with status " + resp.Status)
}

	return success("connected to " + serverName)
}