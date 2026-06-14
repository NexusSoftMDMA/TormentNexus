package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleGetFood(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	url := "https://foodblock.example.com/food/" + id
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch food: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return success(result)
}

func HandleSearchFood(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	url := "https://foodblock.example.com/search?q=" + query
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("search failed: " + e.Error())
}

	defer resp.Body.Close()
	var results []map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&results); e != nil {
		return err("decode error: " + e.Error())
}

	return success(results)
}