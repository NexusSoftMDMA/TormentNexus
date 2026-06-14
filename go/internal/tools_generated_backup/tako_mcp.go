package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func HandleGetTime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	now := time.Now().Format(time.RFC3339)
	return ok(fmt.Sprintf(`{"time":"%s"}`, now))
}

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message parameter required")
}

	resp, e := http.DefaultClient.Get("https://httpbin.org/anything")
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	jsonStr, _ := json.Marshal(result)
	return ok(string(jsonStr))
}