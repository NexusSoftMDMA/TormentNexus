package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func CreateIncident(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	desc, _ :=getString(args, "description")
	severity, _ :=getString(args, "severity")

	body := map[string]string{
		"title":       title,
		"description": desc,
		"severity":    severity,
	}
	payload, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.runframe.io/incidents", bytes.NewReader(payload))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	return ok("incident created")
}

func ListIncidents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.runframe.io/incidents", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	var incidents []map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&incidents); e != nil {
		return err("failed to decode response: " + e.Error())
}

	output, e := json.Marshal(incidents)
	if e != nil {
		return err("failed to marshal output: " + e.Error())
}

	return ok("incidents: " + string(output))
}