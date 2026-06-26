package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleListApps(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.shinkai.app/apps")
	if e != nil {
		return err("failed to fetch apps: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}

func HandleGetAppInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	appID, _ :=getString(args, "app_id")
	if appID == "" {
		return err("app_id is required")
}

	resp, e := http.DefaultClient.Get("https://api.shinkai.app/apps/" + appID)
	if e != nil {
		return err("failed to fetch app info: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}