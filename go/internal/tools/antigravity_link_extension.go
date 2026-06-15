//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func HandleSendMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message is required")
}

	body, e := json.Marshal(map[string]interface{}{"message": message})
	if e != nil {
		return err("failed to marshal: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://antigravity.link/api/send-message", strings.NewReader(string(body)))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("server returned " + resp.Status)
}

	return ok("message sent")
}
