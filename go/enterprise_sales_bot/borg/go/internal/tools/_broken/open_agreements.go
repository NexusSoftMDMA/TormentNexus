package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleAgreements(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filter, _ :=getString(args, "status")
	url := "https://api.openagreements.com/agreements"
	if filter != "" {
		url += "?status=" + filter
	}
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("network error")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error")
}

	return success(string(body))
}