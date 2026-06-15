package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleListServers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key required")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.volthq.com/v1/servers", nil)
	if e != nil {
		return err("request error: "+e.Error())
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request error: "+e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: "+e.Error())
	}
	return ok(string(body))
}