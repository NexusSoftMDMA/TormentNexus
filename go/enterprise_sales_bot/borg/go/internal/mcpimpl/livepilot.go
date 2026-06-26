package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleDeviceAtlas(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	devices := map[string]string{
		"kick":   "Ableton Kick 909",
		"snare":  "Ableton Snare 808",
		"bass":   "Ableton Bass Sub",
	}
	info, found := devices[name]
	if !found {
		return err("device not found")
}

	return success(info)
}

func HandleSearchSplice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.splice.com/v1/sounds?q="+query, nil)
	if e != nil {
		return err("request creation failed")
}

	req.Header.Set("Accept", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed")
}

	data, _ := json.Marshal(result)
	return success(string(data))
}// touch 1781132130
