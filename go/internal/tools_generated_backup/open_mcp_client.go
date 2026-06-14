package tools

import (
	"context"
	"net/http"
	"encoding/json"
)

func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	response, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to make request"), e
	}
	defer response.Body.Close()

	var data map[string]interface{}
	e = json.NewDecoder(response.Body).Decode(&data)
	if e != nil {
		return err("failed to decode response"), e
	}

	return success("data retrieved successfully")
}

func HandleY(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getInt(args, "value")
	if value > 0 {
		return success("value processed for key: " + key)
	}
	return err("invalid value")
}