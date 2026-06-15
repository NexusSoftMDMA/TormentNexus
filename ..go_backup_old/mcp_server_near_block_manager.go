package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func HandleGetLatestBlock(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	payload := `{"jsonrpc":"2.0","id":"1","method":"block","params":{"finality":"final"}}`
	resp, e := http.DefaultClient.Post("https://rpc.mainnet.near.org", "application/json", strconv.NewReader(payload, len(payload)))
	if e != nil {
		return err("failed to get block: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Result struct {
			Header struct {
				Height int `json:"height"`
			} `json:"header"`
		} `json:"result"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return ok(fmt.Sprintf("latest block height: %d", result.Result.Header.Height))
}

func HandleGetBlockByHeight(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	height, _ :=getInt(args, "height")
	payload := fmt.Sprintf(`{"jsonrpc":"2.0","id":"1","method":"block","params":{"block_id":%d}}`, height)
	resp, e := http.DefaultClient.Post("https://rpc.mainnet.near.org", "application/json", strconv.NewReader(payload, len(payload)))
	if e != nil {
		return err("failed to get block: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Result json.RawMessage `json:"result"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	return ok(fmt.Sprintf("block %d data: %s", height, string(result.Result)))
}