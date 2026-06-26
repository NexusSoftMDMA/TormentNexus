package mcpimpl

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

// HandleGetKoreanQuote retrieves a Korean quote based on category.
func HandleGetKoreanQuote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	category, _ :=getString(args, "category")
	url := "https://api.korean-quotes.com/random?category=" + category
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch quote: " + e.Error())
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

	quote, found := result["quote"].(string)
	if !found {
		return err("quote not found in response")
}

	return success(quote)
}