package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSeekchat(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	url := fmt.Sprintf("https://api.seekchat.dev/chat?q=%s", message)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to call Seekchat API: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid JSON: " + e.Error())
}

	reply, found := result["reply"].(string)
	if !found {
		return err("unexpected response format")
}

	return ok(reply)
}