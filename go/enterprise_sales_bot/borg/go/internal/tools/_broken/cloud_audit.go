package mcpimpl

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

// HandleGetAuditLogs fetches audit logs from a given URL.
func HandleGetAuditLogs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	limit, _ :=getInt(args, "limit")
	if limit < 1 {
		limit = 100
	}
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	q := req.URL.Query()
	q.Set("limit", itoa(limit))
	req.URL.RawQuery = q.Encode()
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err("unexpected status: " + resp.Status)
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid JSON: " + e.Error())
}

	return ok(formatJSON(result))
}

// HandleListProviders returns available cloud providers (mock).
func HandleListProviders(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok(`["aws","gcp","azure"]`)
}