package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleGetCollections(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	if apiKey == "" {
		return err("apiKey required")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.getpostman.com/collections", nil)
	if e != nil {
		return err(e.Error())
	}
	req.Header.Set("X-Api-Key", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	var data map[string]interface{}
	if e = json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(e.Error())
	}
	return success("collections retrieved")
}

func HandleGetEnvironments(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	if apiKey == "" {
		return err("apiKey required")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.getpostman.com/environments", nil)
	if e != nil {
		return err(e.Error())
	}
	req.Header.Set("X-Api-Key", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	var data map[string]interface{}
	if e = json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(e.Error())
	}
	return success("environments retrieved")
}// touch 1781132138
