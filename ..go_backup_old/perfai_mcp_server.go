package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleAnalyze(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("missing code")
	}
	reqBody, e := json.Marshal(map[string]string{"content": code, "type": "security"})
	if e != nil {
		return err("failed to marshal request")
	}
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.perfai.io/analyze", bytes.NewReader(reqBody))
	if e != nil {
		return err("failed to create request")
	}
	req.Header.Set("Authorization", "Bearer "+getString(args, "token"))
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("analysis failed with status " + string(resp.StatusCode))
	}
	return success("Analysis complete")
}

func HandleAuthCheck(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	if token == "" {
		return err("missing auth token")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.perfai.io/verify", nil)
	if e != nil {
		return err("failed to create verify request")
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("verification failed: " + e.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("invalid token")
	}
	return success("authenticated")
}