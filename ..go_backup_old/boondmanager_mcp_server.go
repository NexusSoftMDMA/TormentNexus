package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func HandleGetResource(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	url := fmt.Sprintf("https://api.boondmanager.com/resource/%s", id)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %d", resp.StatusCode))
}

	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(e.Error())
}

	return ok(data)
}

func HandleListResources(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 20
	}
	page, _ :=getInt(args, "page")
	if page <= 0 {
		page = 1
	}
	url := fmt.Sprintf("https://api.boondmanager.com/resources?limit=%d&page=%d", limit, page)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %d", resp.StatusCode))
}

	var data []map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(e.Error())
}

	return ok(data)
}