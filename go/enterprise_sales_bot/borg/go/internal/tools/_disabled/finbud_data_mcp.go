package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleFinbudData(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	url := "https://api.example.com/price?symbol=" + symbol
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch data")
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to parse response")
}

	return success(data)
}