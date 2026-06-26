package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetIncident(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	instance, _ :=getString(args, "instance_url")
	table, _ :=getString(args, "table")
	if table == "" {
		table = "incident"
	}
	sysID, _ :=getString(args, "sys_id")
	if instance == "" || sysID == "" {
		return err("instance_url and sys_id are required")
}

	url := fmt.Sprintf("%s/api/now/table/%s/%s", instance, table, sysID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth(getString(args, "username"), getString(args, "password"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("Incident: %v", result["result"]))
}

func HandleListIncidents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	instance, _ :=getString(args, "instance_url")
	table, _ :=getString(args, "table")
	if table == "" {
		table = "incident"
	}
	query, _ :=getString(args, "query")
	if instance == "" {
		return err("instance_url is required")
}

	url := fmt.Sprintf("%s/api/now/table/%s?sysparm_query=%s", instance, table, query)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth(getString(args, "username"), getString(args, "password"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("Incidents: %v", result["result"]))
}