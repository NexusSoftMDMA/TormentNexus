package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func ListApps(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("http://localhost:8080/api/apps")
	if e != nil {
		return err(fmt.Sprintf("failed to fetch apps: %v", e))
}

	defer resp.Body.Close()
	var apps []string
	if e := json.NewDecoder(resp.Body).Decode(&apps); e != nil {
		return err(fmt.Sprintf("failed to decode apps: %v", e))
}

	return ok(apps)
}

func GetAppInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	appName, _ :=getString(args, "appName")
	if appName == "" {
		return err("appName is required")
}

	url := fmt.Sprintf("http://localhost:8080/api/apps/%s", appName)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch app info: %v", e))
}

	defer resp.Body.Close()
	var info map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&info); e != nil {
		return err(fmt.Sprintf("failed to decode app info: %v", e))
}

	return ok(info)
}