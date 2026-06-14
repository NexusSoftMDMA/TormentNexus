package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message is required")
}

	return ok(fmt.Sprintf("You said: %s", message))
}

func HandleCurrentTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("http://worldtimeapi.org/api/timezone/Etc/UTC")
	if e != nil {
		return err(fmt.Sprintf("failed to fetch time: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("invalid JSON response")
}

	datetime, found := data["datetime"].(string)
	if !found {
		return err("datetime field not found in response")
}

	return ok(fmt.Sprintf("Current UTC time: %s", datetime))
}