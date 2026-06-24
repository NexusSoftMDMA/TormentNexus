package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetFeed(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	feedID, _ :=getString(args, "feedID")
	u := fmt.Sprintf("https://api.byte.com/feeds/%s", feedID)
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode error: " + e.Error())
}

	return ok(data)
}

func HandleGetAttestation(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	txHash, _ :=getString(args, "txHash")
	u := fmt.Sprintf("https://api.byte.com/attestations/%s", txHash)
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode error: " + e.Error())
}

	return ok(data)
}