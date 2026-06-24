package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetBlocks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit == 0 {
		limit = 10
	}
	u := fmt.Sprintf("https://api.b10cks.com/v1/blocks?limit=%d", limit)
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	var data interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode error")
}

	return ok(fmt.Sprintf("blocks: %v", data))
}

func HandleCreateBlock(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	payload := map[string]string{"name": name}
	b, e := json.Marshal(payload)
	if e != nil {
		return err("marshal error")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.b10cks.com/v1/blocks", nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	return success(fmt.Sprintf("block %s created", name))
}