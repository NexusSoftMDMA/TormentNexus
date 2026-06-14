package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetSchema(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to get schema: " + e.Error())
}

	defer resp.Body.Close()
	var data interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode error: " + e.Error())
}

	return ok(fmt.Sprintf("Schema: %v", data))
}

func HandleGetRoutes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to get routes: " + e.Error())
}

	defer resp.Body.Close()
	var data interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode error: " + e.Error())
}

	return ok(fmt.Sprintf("Routes: %v", data))
}