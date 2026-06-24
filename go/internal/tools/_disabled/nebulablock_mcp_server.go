package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetBlock(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "blockId")
	url := fmt.Sprintf("https://api.nebulablock.io/blocks/%s", id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch block")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse block data")
}

	return ok(fmt.Sprintf("Block data: %v", result))
}

func HandleSearchBlocks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 10
	}
	url := fmt.Sprintf("https://api.nebulablock.io/blocks?q=%s&limit=%d", query, limit)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to search blocks")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read search results")
}

	var results []map[string]interface{}
	if e := json.Unmarshal(body, &results); e != nil {
		return err("failed to parse search results")
}

	return ok(fmt.Sprintf("Found %d blocks", len(results)))
}