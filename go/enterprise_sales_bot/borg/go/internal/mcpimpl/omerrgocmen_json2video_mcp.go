package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleCreateVideo_omerrgocmen_json2video_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	jsonStr, _ :=getString(args, "json")
	if jsonStr == "" {
		return err("missing 'json' argument")
}

	bodyBytes, e := json.Marshal(map[string]interface{}{"json": jsonStr})
	if e != nil {
		return err("marshal error: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.json2video.com/v1/create", bytes.NewReader(bodyBytes))
	if e != nil {
		return err("request error: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API status %d: %s", resp.StatusCode, string(respBody)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(respBody, &result); e != nil {
		return err("parse error: " + e.Error())
}

	return ok(fmt.Sprintf("Video created: %v", result))
}