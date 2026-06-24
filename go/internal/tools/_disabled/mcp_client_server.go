package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	response, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to make request")
}

	defer response.Body.Close()

	var data map[string]interface{}
	e = json.NewDecoder(response.Body).Decode(&data)
	if e != nil {
		return err("failed to decode response")
}

	return success("request successful")
}

func HandleY(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	message, _ :=getString(args, "message")
	if message == "" {
		return err("message is required")
}

	return success("received message: " + message)
}