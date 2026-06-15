package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleGetChainInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "nodeUrl")
	if url == "" {
		url = "http://localhost:9933"
	}
	body := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "system_chain",
		"params":  []interface{}{},
		"id":      1,
	}
	b, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request")
}

	resp, e := http.DefaultClient.Post(url, "application/json", bytes.NewReader(b))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Result string `json:"result"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("parse error: " + e.Error())
}

	return success(result.Result)
}