package mcpimpl

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
)

func HandleListNotes_obsidian_mcp_pro(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base := os.Getenv("OBSIDIAN_URI")
	if base == "" {
		base = "http://localhost:27123"
	}
	resp, e := http.DefaultClient.Get(base + "/notes")
	if e != nil {
		return err("failed to fetch notes: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}

func HandleReadNote_obsidian_mcp_pro(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	noteName, _ :=getString(args, "name")
	if noteName == "" {
		return err("note name is required")
}

	base := os.Getenv("OBSIDIAN_URI")
	if base == "" {
		base = "http://localhost:27123"
	}
	url := base + "/notes/" + noteName
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to read note: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}