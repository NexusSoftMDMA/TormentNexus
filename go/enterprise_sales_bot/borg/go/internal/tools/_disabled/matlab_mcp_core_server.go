package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

func HandleEvalMatlab(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("code is required")
}

	baseURL := os.Getenv("MATLAB_MCP_URL")
	if baseURL == "" {
		baseURL = "http://localhost:31415"
	}
	body := strings.NewReader(`{"code":"` + code + `"}`)
	resp, e := http.DefaultClient.Post(baseURL+"/eval", "application/json", body)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Success bool   `json:"success"`
		Output  string `json:"output"`
		Error   string `json:"error"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	if !result.Success {
		return err(result.Error)
}

	return ok(result.Output)
}

func HandleListVars(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL := os.Getenv("MATLAB_MCP_URL")
	if baseURL == "" {
		baseURL = "http://localhost:31415"
	}
	resp, e := http.DefaultClient.Get(baseURL + "/vars")
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Variables []string `json:"variables"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	out, _ := json.Marshal(result.Variables)
	return ok(string(out))
}