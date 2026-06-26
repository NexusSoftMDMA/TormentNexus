package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

func HandleCreateThread(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	body, _ :=getString(args, "body")
	apiBase := os.Getenv("UNTHREAD_API_BASE")
	if apiBase == "" {
		apiBase = "https://api.unthread.io/v1"
	}
	reqBody, e := json.Marshal(map[string]string{"title": title, "body": body})
	if e != nil {
		return err("failed to marshal request body")
}

	resp, e := http.DefaultClient.Post(apiBase+"/threads", "application/json", strings.NewReader(string(reqBody)))
	if e != nil {
		return err("failed to send request")
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return err("unexpected status code " + resp.Status)
}

	return success("thread created")
}