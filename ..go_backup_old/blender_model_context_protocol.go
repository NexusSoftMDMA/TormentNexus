package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleRunScript(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	script, _ :=getString(args, "script")
	if url == "" || script == "" {
		return err("missing url or script")
}

	payload := map[string]string{"script": script}
	body, _ := json.Marshal(payload)
	req, e := http.NewRequestWithContext(ctx, "POST", url+"/run", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("server returned %d: %s", resp.StatusCode, string(data)))
}

	return success(string(data))
}

func HandleGetSceneInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("missing url")
}

	req, e := http.NewRequestWithContext(ctx, "GET", url+"/scene", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("server returned %d: %s", resp.StatusCode, string(data)))
}

	return ok(string(data))
}