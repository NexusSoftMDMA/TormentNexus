//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetStockPrice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return success("Please provide a symbol")
}

	url := fmt.Sprintf("https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol=%s&apikey=demo", symbol)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("Failed to fetch stock price: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("Failed to read response: " + e.Error())
}

	var result map[string]interface{	if e := json.Unmarshal(body, &result); e != nil {
		return err("Failed to parse JSON: " + e.Error())
	if note, found := result["Note"]; found {
		return err(note.(string))
}

	quote, found := result["Global Quote"]
	if !found {
		return err("No quote data")
}


-reasoner (deepseek)*
}
}
}