//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListApps(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "apiKey")
	if token == "" {
		return err("apiKey is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.appstoreconnect.apple.com/v1/apps", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read error: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %s", string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("json error: %v", e))
}

	apps, found := result["data"].([]interface{})
	if !found {
		return err("no apps found")
}

	out := fmt.Sprintf("Found %d apps", len(apps))
	return success(out)
}

func HandleGetApp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "apiKey")
	if token == "" {
		return err("apiKey is required")
}

	appID, _ :=getString(args, "appId")
	if appID == "" {
		return err("appId is required")
}

	url := fmt.Sprintf("https://api.appstoreconnect.apple.com/v1/apps/%s", appID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read error: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %s", string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("json error: %v", e))
}

	data, found := result["data"].(map[string]interface{})
	if !found {
		return err("app data not found")
}

	attr, found := data["attributes"].(map[string]interface{})
	if !found {
		return err("no attributes")
}

	name, _ :=getString(attr, "name")
	bundleID, _ :=getString(attr, "bundleId")
	out := fmt.Sprintf("App: %s (%s)", name, bundleID)
	return success(out)
}
