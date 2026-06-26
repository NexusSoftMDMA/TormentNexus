package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleSmgStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	endpoint, _ :=getString(args, "endpoint")
	if endpoint == "" {
		endpoint = "http://localhost:8000/health"
	}
	req, e := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if e != nil {
		return err("request creation failed: " + e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	return success("Smg gateway status: " + resp.Status)
}

func HandleSmgRoute(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	model, _ :=getString(args, "model")
	prompt, _ :=getString(args, "prompt")
	if model == "" || prompt == "" {
		return err("missing model or prompt")
	}
	payload := map[string]interface{}{"model": model, "prompt": prompt}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("marshal failed: " + e.Error())
	}
	req, e := http.NewRequestWithContext(ctx, "POST", "http://localhost:8000/v1/completions", nil)
	if e != nil {
		return err("request creation failed: " + e.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("routing failed: " + e.Error())
	}
	defer resp.Body.Close()
	return success("routed to " + model)
}