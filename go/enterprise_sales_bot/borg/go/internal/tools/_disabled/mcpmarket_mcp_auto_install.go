package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleInstallMCP(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	body, e := json.Marshal(map[string]string{"name": name})
	if e != nil {
		return err("failed to marshal request")
}

	resp, e := http.DefaultClient.Post("http://localhost:3000/install", "application/json", bytes.NewReader(body))
	if e != nil {
		return err("install request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("install failed with status " + resp.Status)
}

	return ok("MCP server " + name + " installed successfully")
}