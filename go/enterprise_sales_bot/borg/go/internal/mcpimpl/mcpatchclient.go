package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleGetVersion_mcpatchclient(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	version, _ :=getString(args, "version")
	if version == "" {
		version = "1.0"
	}
	return ok("McPatch client version: " + version)
}

func HandleCheckUpdate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	current, _ :=getString(args, "current")
	req, e := http.NewRequestWithContext(ctx, "GET", "http://example.com/check?ver="+current, nil)
	if e != nil {
		return err("request failed: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("http error: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode error: " + e.Error())
}

	data, found := result["update_available"].(bool)
	if !found {
		data = false
	}
	if data {
		return success("Update available")
}

	return ok("No updates found")
}