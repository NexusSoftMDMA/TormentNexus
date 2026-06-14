package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleListHITLItems(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		return err("base_url is required")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/hitl-items", nil)
	if e != nil {
		return err(e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
	}
	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(e.Error())
	}
	return ok("HITL items retrieved successfully")
}

func HandleProcessHITLItem(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	itemID, _ :=getString(args, "item_id")
	action, _ :=getString(args, "action")
	if baseURL == "" || itemID == "" || action == "" {
		return err("base_url, item_id, and action are required")
	}
	payload, _ := json.Marshal(map[string]string{"action": action})
	req, e := http.NewRequestWithContext(ctx, "POST", baseURL+"/hitl-items/"+itemID+"/process", payload)
	if e != nil {
		return err(e.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
	}
	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(e.Error())
	}
	return ok("HITL item processed successfully")
}