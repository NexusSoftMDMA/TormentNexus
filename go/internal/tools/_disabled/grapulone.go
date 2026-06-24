package tools

import (
	"context"
	"net/http"
	"encoding/json"
	"fmt"
)

func HandleGetCallers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	repo, _ :=getString(args, "repo")
	funcName, _ :=getString(args, "function")
	url := fmt.Sprintf("https://api.grapulone.example/callers?repo=%s&function=%s", repo, funcName)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch callers: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	callers, found := result["callers"].([]interface{})
	if !found {
		return success("no callers found")
}

	return ok(fmt.Sprintf("callers: %v", callers))
}

func HandleGetCallees(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	repo, _ :=getString(args, "repo")
	funcName, _ :=getString(args, "function")
	url := fmt.Sprintf("https://api.grapulone.example/callees?repo=%s&function=%s", repo, funcName)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch callees: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	callees, found := result["callees"].([]interface{})
	if !found {
		return success("no callees found")
}

	return ok(fmt.Sprintf("callees: %v", callees))
}