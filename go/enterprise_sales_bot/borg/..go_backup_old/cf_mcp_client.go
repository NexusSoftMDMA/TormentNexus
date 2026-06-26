package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleChat(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return err("message argument is required")
}

	body, e := json.Marshal(map[string]string{"input": msg})
	if e != nil {
		return err("failed to marshal request")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.tanzu.vmware.com/chat", nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	return success("Chat request sent successfully")
}

func HandleStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.tanzu.vmware.com/status", nil)
	if e != nil {
		return err("failed to create status request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("status check failed: " + e.Error())
}

	defer resp.Body.Close()
	return ok("Tanzu Platform is operational")
}