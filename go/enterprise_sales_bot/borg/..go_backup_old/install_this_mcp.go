package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleInstallMcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "mcp_name")
	if name == "" {
		return err("mcp_name is required")
}

	url := fmt.Sprintf("https://install-this-mcp.example.com/install?name=%s", name)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %s", e.Error()))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("install failed: %s", string(body)))
}

	return success(fmt.Sprintf("installed %s", name))
}