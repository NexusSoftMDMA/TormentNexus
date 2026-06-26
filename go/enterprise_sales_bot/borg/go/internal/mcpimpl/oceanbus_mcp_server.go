package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleRegister_oceanbus_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
	}
	payload := map[string]string{"agent": name, "action": "register"}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal")
	}
	req, e := http.NewRequestWithContext(ctx, "POST", "https://oceanbus.network/api/register", bytes.NewReader(body))
	if e != nil {
		return err("request creation failed")
	}
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("network error")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("registration failed")
	}
	return success("agent registered on OceanBus")
}

func HandleMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	target, _ :=getString(args, "target")
	msg, _ :=getString(args, "message")
	if target == "" || msg == "" {
		return err("target and message required")
	}
	payload := map[string]string{"to": target, "body": msg}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("marshal error")
	}
	req, e := http.NewRequestWithContext(ctx, "POST", "https://oceanbus.network/api/send", bytes.NewReader(body))
	if e != nil {
		return err("request error")
	}
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("send failed")
	}
	defer resp.Body.Close()
	return success("message dispatched to " + target)
}// touch 1781132136
