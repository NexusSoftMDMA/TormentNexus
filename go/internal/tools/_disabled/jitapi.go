package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleJitApi(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	url, _ :=getString(args, "url")
	if url == "" {
		url = "https://api.jit.dev/v1/projects"
	}
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result interface{}
	if e = json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	raw, _ := json.Marshal(result)
	return ok(string(raw))
}
}