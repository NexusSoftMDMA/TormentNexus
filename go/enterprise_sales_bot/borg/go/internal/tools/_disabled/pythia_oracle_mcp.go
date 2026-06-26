package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetPrice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("missing symbol argument")
}

	url := fmt.Sprintf("https://api.pyth.network/v1/price/%s/latest", symbol)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var data map[string]interface{}
	if e = json.Unmarshal(body, &data); e != nil {
		return err("json parse: " + e.Error())
}

	price, found := data["price"].(float64)
	if !found {
		return err("price not found in response")
}

	return ok(fmt.Sprintf("Current price of %s: %.2f", symbol, price))
}

func HandleCheckStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("Pythia Oracle MCP is running")
}