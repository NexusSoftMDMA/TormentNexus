package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleListItems_jkawwa_paperclipai_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "https://api.paperclip.ai/v1/items"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var items []interface{}
	if e := json.Unmarshal(body, &items); e != nil {
		return err(fmt.Sprintf("parse failed: %v", e))
}

	return ok(fmt.Sprintf("Found %d items", len(items)))
}

func HandleCreateItem(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	payload := map[string]string{"name": name}
	data, e := json.Marshal(payload)
	if e != nil {
		return err(fmt.Sprintf("marshal failed: %v", e))
}

	url := "https://api.paperclip.ai/v1/items"
	body := strings.NewReader(string(data))
	req, e := http.NewRequestWithContext(ctx, "POST", url, body)
	if e != nil {
		return err(fmt.Sprintf("create request failed: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("post failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	return success("item created")
}