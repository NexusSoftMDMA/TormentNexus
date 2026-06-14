package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleAlphaVantage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("missing symbol")
}

	apiKey := os.Getenv("ALPHA_VANTAGE_API_KEY")
	if apiKey == "" {
		return err("missing ALPHA_VANTAGE_API_KEY environment variable")
}

	url := fmt.Sprintf("https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol=%s&apikey=%s", symbol, apiKey)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to query Alpha Vantage: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	quote, found := result["Global Quote"].(map[string]interface{})
	if !found {
		return err("unexpected response structure")
}

	price, _ := quote["05. price"].(string)
	return ok(fmt.Sprintf("Stock %s price: %s", symbol, price))
}