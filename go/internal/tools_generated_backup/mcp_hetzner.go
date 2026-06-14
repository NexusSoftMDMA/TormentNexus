package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleHetznerListServers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("API key is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.hetzner.cloud/v1/servers", nil)
	if e != nil {
		return err(e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("Hetzner API returned HTTP %d: %s", resp.StatusCode, string(body)))
}

	return success(string(body))
}