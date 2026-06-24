package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleMemtraceStore(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("missing code argument")
	}
	payload := map[string]string{"content": code}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal")
	}
	req, e := http.NewRequestWithContext(ctx, "POST", "http://localhost:8080/store", strings.NewReader(string(body)))
	if e != nil {
		return err("failed to create request")
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
	}
	defer resp.Body.Close()
	return success("code stored in graph")
}

func HandleMemtraceQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("missing query argument")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/query?q="+query, nil)
	if e != nil {
		return err("failed to create request")
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
	}
	defer resp.Body.Close()
	return success("graph query executed")
}