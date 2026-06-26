package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetTelemetry(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "event_id")
	url := fmt.Sprintf("https://api.businesscentral.dynamics.com/telemetry/%s", id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch telemetry")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse telemetry")
}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(data))
}

func HandleListTelemetry(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit == 0 {
		limit = 10
	}
	url := fmt.Sprintf("https://api.businesscentral.dynamics.com/telemetry?$top=%d", limit)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to list telemetry")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var list []map[string]interface{}
	if e := json.Unmarshal(body, &list); e != nil {
		return err("failed to parse list")
}

	data, _ := json.MarshalIndent(list, "", "  ")
	return ok(string(data))
}