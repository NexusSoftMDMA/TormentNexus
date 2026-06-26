package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListDashboards_mcp_grafana(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base_url")
	if base == "" {
		base = "http://localhost:3000"
	}
	api := base + "/api/search"
	resp, e := http.DefaultClient.Get(api)
	if e != nil {
		return err("failed to fetch dashboards: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var dashboards []map[string]interface{}
	if e := json.Unmarshal(body, &dashboards); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("Found %d dashboards", len(dashboards)))
}

func HandleGetDashboard_mcp_grafana(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	uid, _ :=getString(args, "uid")
	if uid == "" {
		return err("uid is required")
}

	base, _ :=getString(args, "base_url")
	if base == "" {
		base = "http://localhost:3000"
	}
	api := base + "/api/dashboards/uid/" + uid
	resp, e := http.DefaultClient.Get(api)
	if e != nil {
		return err("failed to fetch dashboard: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("Grafana returned status %d", resp.StatusCode))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var dashboard map[string]interface{}
	if e := json.Unmarshal(body, &dashboard); e != nil {
		return err("failed to parse dashboard: " + e.Error())
}

	title, _ := dashboard["dashboard"].(map[string]interface{})["title"].(string)
	return ok(fmt.Sprintf("Dashboard title: %s", title))
}