package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
)

func HandleListWorkflows(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host := os.Getenv("N8N_HOST")
	if host == "" {
		host = "http://localhost:5678"
	}
	req, e := http.NewRequestWithContext(ctx, "GET", host+"/rest/workflows", nil)
	if e != nil {
		return err("create request: " + e.Error())
}

	req.Header.Set("X-N8N-API-KEY", os.Getenv("N8N_API_KEY"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var r map[string]interface{}
	if e = json.NewDecoder(resp.Body).Decode(&r); e != nil {
		return err("decode: " + e.Error())
}

	b, _ := json.MarshalIndent(r, "", "  ")
	return ok(string(b))
}

func HandleCreateWorkflow(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name required")
}

	payload := map[string]interface{}{
		"name":        name,
		"nodes":       getString(args, "nodes"),
		"connections": getString(args, "connections"),
		"active":      false,
	}
	body, _ := json.Marshal(payload)
	host := os.Getenv("N8N_HOST")
	if host == "" {
		host = "http://localhost:5678"
	}
	req, e := http.NewRequestWithContext(ctx, "POST", host+"/rest/workflows", bytes.NewReader(body))
	if e != nil {
		return err("create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-N8N-API-KEY", os.Getenv("N8N_API_KEY"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var r map[string]interface{}
	if e = json.NewDecoder(resp.Body).Decode(&r); e != nil {
		return err("decode: " + e.Error())
}

	b, _ := json.MarshalIndent(r, "", "  ")
	return ok(string(b))
}