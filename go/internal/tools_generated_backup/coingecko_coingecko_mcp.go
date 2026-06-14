package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetSimplePrice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ids, _ :=getString(args, "ids")
	vsCurrencies, _ :=getString(args, "vs_currencies")
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=%s", ids, vsCurrencies)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}

func HandleGetCoinList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "https://api.coingecko.com/api/v3/coins/list"
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result []map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}