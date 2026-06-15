package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleQueryContent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseUrl, _ :=getString(args, "baseUrl")
	path, _ :=getString(args, "path")
	if path == "" {
		return err("missing path parameter")
}

	req, e := http.NewRequestWithContext(ctx, "GET", baseUrl+path, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	return ok(string(body))
}

func HandleGetAsset(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseUrl, _ :=getString(args, "baseUrl")
	assetId, _ :=getString(args, "assetId")
	if assetId == "" {
		return err("missing assetId parameter")
}

	req, e := http.NewRequestWithContext(ctx, "GET", baseUrl+"/assets/"+assetId, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(fmt.Sprintf("failed to decode JSON: %v", e))
}

	return success("Asset retrieved successfully")
}