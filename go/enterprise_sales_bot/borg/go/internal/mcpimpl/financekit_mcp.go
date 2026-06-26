package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleStockQuote_financekit_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol is required")
}

	url := fmt.Sprintf("https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol=%s&apikey=demo", symbol)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch quote: " + e.Error())
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON")
}

	globalQuote, found := result["Global Quote"].(map[string]interface{})
	if !found {
		return err("no quote data found")
}

	price, found := globalQuote["05. price"].(string)
	if !found {
		return err("price not found")
}

	return ok("Current price: " + price)
}