package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListThreats(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	const baseURL = "http://localhost:9090"
	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/threats", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(string(body))
}

func HandleBlockThreat(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	threatID, _ :=getString(args, "threat_id")
	if threatID == "" {
		return err("missing threat_id")
}

	const baseURL = "http://localhost:9090"
	req, e := http.NewRequestWithContext(ctx, "POST", baseURL+"/threats/"+threatID+"/block", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("block failed with status " + resp.Status)
}

	return success("threat " + threatID + " blocked")
}