package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func ListHostingPlans(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.hostsmith.com/plans", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	var plans []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&plans); e != nil {
		return err("decode failed: " + e.Error())
}

	return ok(fmt.Sprintf("Plans: %+v", plans))
}

func CreateHosting(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	plan, _ :=getString(args, "plan")
	body, _ := json.Marshal(map[string]string{"plan": plan})
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.hostsmith.com/hosting", bytes.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	return success("Hosting created successfully")
}