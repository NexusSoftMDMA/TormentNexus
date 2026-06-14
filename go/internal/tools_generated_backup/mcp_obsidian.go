package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListNotes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	vault, _ :=getString(args, "vault")
	url := fmt.Sprintf("http://127.0.0.1:27123/vault/%s/notes", vault)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to reach Obsidian API: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON")
}

	return ok(fmt.Sprintf("Notes: %v", result))
}

func HandleGetNote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	vault, _ :=getString(args, "vault")
	path, _ :=getString(args, "path")
	url := fmt.Sprintf("http://127.0.0.1:27123/vault/%s/%s", vault, path)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to reach Obsidian API: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return ok(fmt.Sprintf("Note content: %s", string(body)))
}