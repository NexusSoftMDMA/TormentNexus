package tools

import (
	"context"
	"net/http"
)

func HandleChatterboxio(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message is required")
	}
	return success("Chatterboxio received: " + message)
}

func HandleChatterboxioStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		url = "https://chatterbox.io/status"
	}
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to reach chatterbox: " + e.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("chatterbox returned status " + http.StatusText(resp.StatusCode))
	}
	return ok("chatterbox is reachable")
}