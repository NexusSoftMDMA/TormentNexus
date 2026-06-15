package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleListSaves(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.openclaw.example.com/saves")
	if e != nil {
		return err("failed to list saves: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return ok(string(body))
}

func HandleLoadSave(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	saveName, _ :=getString(args, "save_name")
	if saveName == "" {
		return err("save_name is required")
}

	url := "https://api.openclaw.example.com/saves/" + saveName
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to load save: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return ok(string(body))
}