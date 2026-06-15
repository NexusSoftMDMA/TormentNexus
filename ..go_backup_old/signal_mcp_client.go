package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func HandleSendSignal(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	to, _ :=getString(args, "to")
	msg, _ :=getString(args, "message")
	if to == "" {
		return err("missing 'to' recipient")
	}
	payload := map[string]string{"to": to, "message": msg}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal json")
	}
	req, e := http.NewRequestWithContext(ctx, "POST", "http://localhost:8080/send", strings.NewReader(string(body)))
	if e != nil {
		return err("failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	return success("signal sent to " + to)
}

func HandleReceiveSignal(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	from, _ :=getString(args, "from")
	req, e := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/receive?from="+from, nil)
	if e != nil {
		return err("failed to create request")
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	return success("messages retrieved for " + from)
}