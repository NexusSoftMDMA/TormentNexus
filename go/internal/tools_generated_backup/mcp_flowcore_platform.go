package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListFlows(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "https://api.flowcore.io/v1/flows"
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
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

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Flows: %v", result))
}

func HandleGetFlow(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	flowID, _ :=getString(args, "flow_id")
	if flowID == "" {
		return err("flow_id is required")
}

	url := fmt.Sprintf("https://api.flowcore.io/v1/flows/%s", flowID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
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

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Flow: %v", result))
}