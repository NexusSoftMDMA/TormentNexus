package tools

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func HandleAnalyze(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	file, _ :=getString(args, "file")
	if file == "" {
		return err("file parameter is required")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "http://localhost:8080/analyze", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
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
		return err("invalid JSON: " + e.Error())
}

	msg, found := result["message"].(string)
	if !found || msg == "" {
		return err("no message in response")
}

	return ok(msg)
}

func HandleDecompile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	function, _ :=getString(args, "function")
	if function == "" {
		return err("function parameter is required")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "http://localhost:8080/decompile", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
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
		return err("invalid JSON: " + e.Error())
}

	code, found := result["decompiled"].(string)
	if !found || code == "" {
		return err("no decompiled code in response")
}

	return success(code)
}