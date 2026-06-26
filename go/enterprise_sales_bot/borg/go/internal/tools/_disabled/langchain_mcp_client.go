package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleConnect(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serverURL, _ :=getString(args, "server_url")
	provider, _ :=getString(args, "provider")
	if serverURL == "" {
		return err("server_url is required")
}

	payload := map[string]string{"url": serverURL, "provider": provider}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal request")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "http://localhost:8080/connect", nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("connection failed: " + e.Error())
}

	defer resp.Body.Close()
	return success("connected to " + serverURL + " via " + provider)
}

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	model, _ :=getString(args, "model")
	if query == "" {
		return err("query is required")
}

	payload := map[string]string{"query": query, "model": model}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal query")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "http://localhost:8080/query", nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("query failed: " + e.Error())
}

	defer resp.Body.Close()
	return success("query executed with model " + model)
}