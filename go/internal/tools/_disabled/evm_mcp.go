package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleEvmMcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	rpcURL, _ :=getString(args, "rpc_url")
	if rpcURL == "" {
		return err("rpc_url required")
}

	method, _ :=getString(args, "method")
	if method == "" {
		method = "eth_blockNumber"
	}
	paramsStr, _ :=getString(args, "params")
	reqBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  []interface{}{},
		"id":      1,
	}
	if paramsStr != "" {
		var parsed []interface{}
		if e := json.Unmarshal([]byte(paramsStr), &parsed); e != nil {
			return err("invalid params json")
}

		reqBody["params"] = parsed
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req, e := http.NewRequestWithContext(ctx, http.MethodPost, rpcURL, bytes.NewReader(bodyBytes))
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	respBytes, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var rpcResp map[string]interface{}
	if e := json.Unmarshal(respBytes, &rpcResp); e != nil {
		return err("invalid json response")
}

	if rpcResp["error"] != nil {
		return err("rpc error")
}

	return ok(fmt.Sprintf("%v", rpcResp["result"]))
}