package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleOpenDeal(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	value, _ :=getInt(args, "value")
	if name == "" {
		return err("name is required")
}

	payload := fmt.Sprintf(`{"name":"%s","value":%d}`, name, value)
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.opensalesstack.com/deals", strings.NewReader(payload))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	return success("Deal opened successfully")
}

func HandleGetLeads(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	page, _ :=getInt(args, "page")
	if page < 1 {
		page = 1
	}
	url := fmt.Sprintf("https://api.opensalesstack.com/leads?page=%d", page)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return ok(fmt.Sprintf("Got %d leads", int(result["count"].(float64))))
}