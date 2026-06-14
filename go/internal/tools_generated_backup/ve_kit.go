package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		url = "https://api.vekit.example.com/status"
	}
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("json decode: %v", e))
}

	return ok(fmt.Sprintf("status: %v", data))
}

func HandlePing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("pong")
}