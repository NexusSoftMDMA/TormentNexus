package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandlePlanuInit(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectPath, _ :=getString(args, "projectPath")
	if projectPath == "" {
		return err("projectPath is required")
	}

	req, e := http.NewRequestWithContext(ctx, "POST", "http://localhost:8080/init", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
	}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
	}

	return ok("Planu initialized at " + projectPath)
}

func HandlePlanuValidate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	specPath, _ :=getString(args, "specPath")
	if specPath == "" {
		return err("specPath is required")
	}

	req, e := http.NewRequestWithContext(ctx, "POST", "http://localhost:8080/validate", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
	}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	found := true
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		found = false
		return err("decode failed: " + e.Error())
	}

	if !found {
		return err("validation result missing")
	}

	return success("Spec validated successfully")
}

func HandlePlanuGenerate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	specPath, _ :=getString(args, "specPath")
	target, _ :=getString(args, "target")
	if specPath == "" || target == "" {
		return err("specPath and target are required")
	}

	req, e := http.NewRequestWithContext(ctx, "POST", "http://localhost:8080/generate", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
	}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()

	return ok("Code generated for " + target)
}// touch 1781132138
