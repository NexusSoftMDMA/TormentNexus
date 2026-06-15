//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func HandleSemgrepScan(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	language, _ :=getString(args, "language")
	if code == "" {
		return err("missing code argument")
	}
	if language == "" {
		language = "python"
	}
	payload := map[string]interface{}{"code": code, "language": language}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal request")
	}
	req, e := http.NewRequestWithContext(ctx, "POST", "https://semgrep.dev/api/scan", strings.NewReader(string(body)))
	if e != nil {
		return err("failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err("failed to decode response")
	}
	return success("scan completed")
}

func HandleSemgrepRule(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ruleID, _ :=getString(args, "rule_id")
	if ruleID == "" {
		return err("missing rule_id argument")
	}
	url := "https://semgrep.dev/api/rules/" + ruleID
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	var rule map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&rule)
	if e != nil {
		return err("failed to decode rule")
	}
	return success("rule fetched")
}// touch 1781132133
