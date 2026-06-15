package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleCreateFlow(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	desc, _ :=getString(args, "description")
	body, e := json.Marshal(map[string]string{"name": name, "description": desc})
	if e != nil {
		return err("failed to marshal request body")
}

	req, e := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.flowlib.dev/flows", bytes.NewReader(body))
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status %d", resp.StatusCode))
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	return ok("flow created successfully")
}

func HandleExecuteFlow(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	flowID, _ :=getString(args, "flow_id")
	input, _ :=getString(args, "input")
	body, e := json.Marshal(map[string]string{"input": input})
	if e != nil {
		return err("failed to marshal request body")
}

	url := fmt.Sprintf("https://api.flowlib.dev/flows/%s/execute", flowID)
	req, e := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status %d", resp.StatusCode))
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	return ok("flow executed successfully")
}