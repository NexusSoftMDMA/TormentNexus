package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetBlock_cerebrochain_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	number, _ :=getString(args, "number")
	if number == "" {
		return err("block number is required")
}

	url := fmt.Sprintf("https://cerebrochain.io/api/block/%s", number)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch block: " + e.Error())
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return success(string(body))
}

func HandleGetTransaction_cerebrochain_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	hash, _ :=getString(args, "hash")
	if hash == "" {
		return err("transaction hash is required")
}

	url := fmt.Sprintf("https://cerebrochain.io/api/tx/%s", hash)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch transaction: " + e.Error())
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var found map[string]interface{}
	json.Unmarshal(body, &found)
	return ok("transaction retrieved")
}