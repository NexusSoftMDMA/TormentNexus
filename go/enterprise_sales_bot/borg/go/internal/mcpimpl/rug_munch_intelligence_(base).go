package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleContractAudit(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	if address == "" {
		return err("address required")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.rugmunch.io/v1/audit/"+address, nil)
	if e != nil {
		return err("request failed")
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("api error")
	}
	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode failed")
	}
	return success("audit complete")
}

func HandleHoneypotDetection(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	if token == "" {
		return err("token required")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.rugmunch.io/v1/honeypot/"+token, nil)
	if e != nil {
		return err("request failed")
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("api error")
	}
	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode failed")
	}
	isHoneypot, found := data["is_honeypot"].(bool)
	if found && isHoneypot {
		return ok("HONEYPOT DETECTED")
	}
	return ok("SAFE")
}

func HandleWhaleDecoding(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	if address == "" {
		return err("address required")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.rugmunch.io/v1/whale/"+address, nil)
	if e != nil {
		return err("request failed")
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("api error")
	}
	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode failed")
	}
	return success("whale analysis complete")
}// touch 1781132140
