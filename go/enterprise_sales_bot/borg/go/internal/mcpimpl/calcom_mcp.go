package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetEvents_calcom_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key is required")
}

	url := "https://api.cal.com/v1/events?apiKey=" + apiKey
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch events: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	events, found := result["events"].([]interface{})
	if !found {
		return err("no events in response")
}

	return ok(fmt.Sprintf("Found %d events", len(events)))
}