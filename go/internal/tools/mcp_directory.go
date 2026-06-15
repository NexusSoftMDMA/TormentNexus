//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListDirectory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	u := "https://mcp-dir.com/api/servers"
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch directory: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var list []map[string]interface{}
	if e = json.Unmarshal(body, &list); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	return ok(fmt.Sprintf("found %d servers", len(list)))
}

func HandleGetDirectoryEntry(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("missing required argument 'name'")
}

	u := fmt.Sprintf("https://mcp-dir.com/api/servers/%s", name)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch entry: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var entry map[string]interface{}
	if e = json.Unmarshal(body, &entry); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	return ok(fmt.Sprintf("entry: %s", entry["name"]))
}
