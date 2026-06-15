package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandlePoolsList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	req, e := http.NewRequestWithContext(ctx, "GET", "https://truenas.local/api/v2.0/pool", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

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

	return ok(fmt.Sprintf("Pools: %v", result))
}