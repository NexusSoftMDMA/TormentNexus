package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetBelief(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	req, e := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://api.believe.com/v1/beliefs/%s", id), nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("failed to decode response: %v", e))
}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(data))
}