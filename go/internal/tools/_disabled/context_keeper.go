package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleFetchContext(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	url := "https://api.contextkeeper.example.com/context?key=" + key
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch context: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok("context fetched")
}

func HandleStoreContext(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	payload, e := json.Marshal(map[string]string{"key": key, "value": value})
	if e != nil {
		return err("failed to marshal payload: " + e.Error())
}

	url := "https://api.contextkeeper.example.com/context"
	resp, e := http.DefaultClient.Post(url, "application/json", nil)
	if e != nil {
		return err("failed to store context: " + e.Error())
}

	defer resp.Body.Close()
	return success("context stored")
}