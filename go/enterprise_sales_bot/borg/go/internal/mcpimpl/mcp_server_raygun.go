package mcpimpl

import (
	"context"
	"io"
	"net/http"
)

func HandleRaygunErrors(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apikey")
	if apiKey == "" {
		return err("apikey is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.raygun.io/errors?apikey="+apiKey, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}