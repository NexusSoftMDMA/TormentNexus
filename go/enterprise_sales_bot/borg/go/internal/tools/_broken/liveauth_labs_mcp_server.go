package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleLiveAuthPow(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	challenge, _ :=getString(args, "challenge")
	if challenge == "" {
		return err("missing challenge")
	}
	reqBody, e := json.Marshal(map[string]string{"challenge": challenge})
	if e != nil {
		return err("failed to marshal request")
	}
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.liveauth.io/pow", bytes.NewReader(reqBody))
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
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed")
	}
	return success("PoW solved")
}

func HandleLiveAuthLightning(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	invoice, _ :=getString(args, "invoice")
	if invoice == "" {
		return err("missing invoice")
	}
	reqBody, e := json.Marshal(map[string]string{"invoice": invoice})
	if e != nil {
		return err("failed to marshal request")
	}
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.liveauth.io/lightning", bytes.NewReader(reqBody))
	if e != nil {
		return err("failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	return success("Lightning payment verified")
}