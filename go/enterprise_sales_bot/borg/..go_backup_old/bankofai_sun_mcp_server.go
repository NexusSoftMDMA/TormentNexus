package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleGetSUNPrice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	_ = args
	resp, e := http.DefaultClient.Get("https://api.sun.io/v1/price")
	if e != nil {
		return err("failed to fetch price: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	price, found := data["price"].(float64)
	if !found {
		return err("price not found in response")
}

	result, e := json.Marshal(map[string]float64{"price": price})
	if e != nil {
		return err("failed to marshal result: " + e.Error())
}

	return ok(string(result))
}