package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleNotify(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("missing url parameter")
}

	message, _ :=getString(args, "message")
	payload, e := json.Marshal(map[string]string{"message": message})
	if e != nil {
		return err("failed to marshal message")
}

	resp, e := http.DefaultClient.Post(url, "application/json", bytes.NewReader(payload))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("server returned status " + resp.Status)
}

	return ok("notification sent")
}