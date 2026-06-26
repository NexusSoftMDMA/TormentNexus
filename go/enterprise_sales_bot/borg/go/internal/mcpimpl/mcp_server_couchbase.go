package mcpimpl

import (
	"context"
	"net/http"
)

func HandlePing_mcp_server_couchbase(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	if host == "" {
		host = "localhost:8091"
	}
	url := "http://" + host + "/"
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to ping couchbase: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("couchbase ping failed: " + resp.Status)
}

	return ok("couchbase server at " + host + " is healthy")
}