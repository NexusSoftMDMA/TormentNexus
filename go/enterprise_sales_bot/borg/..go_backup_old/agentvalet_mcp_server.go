package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleCallPlatform(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	platform, _ :=getString(args, "platform")
	action, _ :=getString(args, "action")
	payload, _ :=getString(args, "payload")
	reqData := map[string]string{"platform": platform, "action": action, "payload": payload}
	body, e := json.Marshal(reqData)
	if e != nil {
		return err("marshal error: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.agentvalet.ai/proxy", bytes.NewReader(body))
	if e != nil {
		return err("request error: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("do error: " + e.Error())
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	return ok(string(respBody))
}

func HandleListPlatforms(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.agentvalet.ai/list-platforms", nil)
	if e != nil {
		return err("request error: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("do error: " + e.Error())
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	return ok(string(respBody))
}