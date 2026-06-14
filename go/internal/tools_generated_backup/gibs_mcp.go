package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleGetGib(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.quotable.io/random")
	if e != nil {
		return err("failed to fetch gib: " + e.Error())
}

	defer resp.Body.Close()
	var data struct {
		Content string `json:"content"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode error: " + e.Error())
}

	return ok(data.Content)
}

func HandlePing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("pong")
}