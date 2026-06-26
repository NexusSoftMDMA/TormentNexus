package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleListFlux(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	repo, _ :=getString(args, "repo")
	url := "https://api.example.com/flux/list?repo=" + repo
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch list")
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid response")
}

	return ok("list received")
}

func HandleRunFlux(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	action, _ :=getString(args, "action")
	if action == "" {
		return err("action required")
}

	payload := map[string]string{"action": action}
	body, _ := json.Marshal(payload)
	resp, e := http.DefaultClient.Post("https://api.example.com/flux/run", "application/json", bytes.NewReader(body))
	if e != nil {
		return err("failed to run action")
}

	defer resp.Body.Close()
	return ok("action executed")
}