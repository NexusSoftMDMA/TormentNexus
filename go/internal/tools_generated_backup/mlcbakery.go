package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleMlcbakeryStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://mlcbakery.com/status")
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()

	var data map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&data)
	if e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	status, found := data["status"].(string)
	if !found {
		return err("status field missing")
}

	return ok(fmt.Sprintf("Mlcbakery status: %s", status))
}

func HandleMlcbakeryMenu(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://mlcbakery.com/menu")
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()

	var data map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&data)
	if e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	items, found := data["items"].([]interface{})
	if !found {
		return err("items field missing")
}

	return ok(fmt.Sprintf("Mlcbakery menu has %d items", len(items)))
}// touch 1781132135
