package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleListIntegrations(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	service, _ :=getString(args, "service")
	url := "https://api.benchclaw.com/integrations?service=" + service
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode: " + e.Error())
}

	data, found := result["data"]
	if !found {
		return err("no data in response")
}

	return ok("Integrations: " + toString(data))
}

func toString(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}