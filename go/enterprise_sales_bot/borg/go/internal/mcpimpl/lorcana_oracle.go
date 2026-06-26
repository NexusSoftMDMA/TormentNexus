package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSearchCards_lorcana_oracle(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	url := fmt.Sprintf("https://api.lorcanaapi.com/cards/search?q=%s", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body: %v", e))
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("parse json: %v", e))
}

	return ok(fmt.Sprintf("Search results: %+v", data))
}

func HandleGetCard_lorcana_oracle(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cardID, _ :=getString(args, "card_id")
	if cardID == "" {
		return err("card_id parameter is required")
}

	url := fmt.Sprintf("https://api.lorcanaapi.com/cards/%s", cardID)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body: %v", e))
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("parse json: %v", e))
}

	return ok(fmt.Sprintf("Card details: %+v", data))
}