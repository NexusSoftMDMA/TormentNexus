package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListItems(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "https://api.example.com/items"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to get items: %v", e))
}

	defer resp.Body.Close()
	var items []map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&items); e != nil {
		return err(fmt.Sprintf("failed to decode items: %v", e))
}

	return ok(fmt.Sprintf("found %d items", len(items)))
}

func HandleGetItem(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	url := fmt.Sprintf("https://api.example.com/items/%s", id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to get item: %v", e))
}

	defer resp.Body.Close()
	var item map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&item); e != nil {
		return err(fmt.Sprintf("failed to decode item: %v", e))
}

	return success(fmt.Sprintf("item: %v", item))
}