package mcpimpl

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleGetSecret_equivault_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path is required")
}

	url := fmt.Sprintf("https://equivault.example.com/secrets/%s", path)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to execute request: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("received status " + resp.Status + ": " + string(body))
}

	return ok("Secret: " + string(body))
}