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
		return err("failed to fetch data")
}

	defer response.Body.Close()

	var data interface{}
	e = json.NewDecoder(response.Body).Decode(&data)
	if e != nil {
		return err("failed to decode response")
}

	return success("data retrieved successfully")
}

func HandleY(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getInt(args, "id")
	if id <= 0 {
		return err("invalid id")
}

	return success("valid id")
}