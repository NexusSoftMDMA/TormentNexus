package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to make request")
}

	defer resp.Body.Close()

	var result map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err("failed to decode response")
}

	return success("request successful")
}

func HandleY(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getInt(args, "value")
	if value > 0 {
		return success("value is positive")
}

	return ok("key: " + key)
}