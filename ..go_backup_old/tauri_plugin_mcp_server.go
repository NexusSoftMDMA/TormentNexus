package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleInvokeCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd, _ :=getString(args, "command")
	payload, found := args["payload"]
	if !found {
		payload = map[string]interface{}{}
	}
	body, e := json.Marshal(map[string]interface{}{"command": cmd, "payload": payload})
	if e != nil {
		return err("failed to marshal request")
	}
	req, e := http.NewRequestWithContext(ctx, "POST", "http://localhost:3000/invoke", bytes.NewBuffer(body))
	if e != nil {
		return err("failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("status %d", resp.StatusCode))
	}
	return success("command executed")
}

func HandleGetState(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	req, e := http.NewRequestWithContext(ctx, "GET", "http://localhost:3000/state", nil)
	if e != nil {
		return err("failed to create request")
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("status %d", resp.StatusCode))
	}
	return success("state retrieved")
}// touch 1781132142
