package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// HandleCreateDispatch sends a dispatch request to the capsule factory
func HandleCreateDispatch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	capsuleID, _ :=getString(args, "capsuleID")
	endpoint, _ :=getString(args, "endpoint")
	if capsuleID == "" || endpoint == "" {
		return err("capsuleID and endpoint are required")
}

	body := fmt.Sprintf(`{"capsuleID":"%s","endpoint":"%s"}`, capsuleID, endpoint)
	req, e := http.NewRequestWithContext(ctx, "POST", "https://czap.example.com/dispatch", strings.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("dispatch failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 202 {
		return err(fmt.Sprintf("dispatch returned status %d", resp.StatusCode))
}

	return ok("dispatch created")
}

// HandleGetDispatchStatus retrieves the status of a dispatch
func HandleGetDispatchStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dispatchID, _ :=getString(args, "dispatchID")
	if dispatchID == "" {
		return err("dispatchID is required")
}

	url := fmt.Sprintf("https://czap.example.com/dispatch/%s", dispatchID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("status check failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("status check returned %d", resp.StatusCode))
}

	data, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(data, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	status, found := result["status"].(string)
	if !found {
		return err("status field missing in response")
}

	return ok(fmt.Sprintf("dispatch status: %s", status))
}