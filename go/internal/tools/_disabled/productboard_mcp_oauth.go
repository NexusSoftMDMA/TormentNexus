package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListFeatures(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	limit, _ :=getInt(args, "limit")
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.productboard.com/features?limit=%d", limit), nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %d", resp.StatusCode))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response")
}

	return success(fmt.Sprintf("Features: %v", result))
}

func HandleCreateFeature(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	body := map[string]string{"name": name}
	if desc := getString(args, "description"); desc != "" {
		body["description"] = desc
	}
	jsonData, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.productboard.com/features", nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(bytes.NewReader(jsonData))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		return err(fmt.Sprintf("API error: %d", resp.StatusCode))
}

	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return success(fmt.Sprintf("Feature created: %s", string(respBody)))
}