package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetBalance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	if address == "" {
		return err("address is required")
}

	reqBody, e := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getBalance",
		"params":  []interface{}{address, "latest"},
		"id":      1,
	})
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	resp, e := http.DefaultClient.Post("https://bsc-dataseed.binance.org/", "application/json", bytes.NewReader(reqBody))
	if e != nil {
		return err("RPC call failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	if e, found := result["error"]; found {
		errObj := e.(map[string]interface{})
		return err(fmt.Sprintf("RPC error: %v", errObj["message"]))
}

	balance, found := result["result"]
	if !found {
		return err("no result in response")
}

	return ok(fmt.Sprintf("Balance: %v", balance))
}