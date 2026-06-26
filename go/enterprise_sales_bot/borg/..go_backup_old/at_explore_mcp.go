package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleResolveHandle(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	handle, _ :=getString(args, "handle")
	if handle == "" {
		return err("handle is required")
	}
	url := fmt.Sprintf("https://public.api.bsky.app/xrpc/com.atproto.identity.resolveHandle?handle=%s", handle)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %s", resp.Status))
	}
	var result struct {
		DID string `json:"did"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
	}
	return success(fmt.Sprintf("Resolved handle %s to DID: %s", handle, result.DID))
}

func HandleGetProfile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	actor, _ :=getString(args, "actor")
	if actor == "" {
		return err("actor is required")
	}
	url := fmt.Sprintf("https://public.api.bsky.app/xrpc/app.bsky.actor.getProfile?actor=%s", actor)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %s", resp.Status))
	}
	var profile struct {
		DisplayName string `json:"displayName"`
		Description string `json:"description"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&profile); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
	}
	return success(fmt.Sprintf("Profile: %s - %s", profile.DisplayName, profile.Description))
}