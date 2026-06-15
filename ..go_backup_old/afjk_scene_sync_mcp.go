package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetScene(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sceneID, _ :=getString(args, "sceneId")
	if sceneID == "" {
		return err("sceneId is required")
}

	req, e := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://afjk.jp/api/scenes/%s", sceneID), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

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
		return err(fmt.Sprintf("unexpected status %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("Scene %s retrieved", sceneID))
}

func HandleSyncScene(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sceneID, _ :=getString(args, "sceneId")
	if sceneID == "" {
		return err("sceneId is required")
}

	req, e := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("https://afjk.jp/api/scenes/%s/sync", sceneID), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

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
		return err(fmt.Sprintf("unexpected status %d: %s", resp.StatusCode, string(body)))
}

	return ok("Scene synced successfully")
}