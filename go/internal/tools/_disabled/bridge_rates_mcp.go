package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetRate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base")
	if base == "" {
		base = "USD"
	}
	target, _ :=getString(args, "target")
	if target == "" {
		target = "EUR"
	}
	url := fmt.Sprintf("https://open.er-api.com/v6/latest/%s", base)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch rates: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	e = json.NewDecoder(resp.Body).Decode(&data)
	if e != nil {
		return err("failed to parse response: " + e.Error())
}

	rates, found := data["rates"].(map[string]interface{})
	if !found {
		return err("rates not found in response")
}

	rate, found := rates[target].(float64)
	if !found {
		return err(fmt.Sprintf("rate for %s not found", target))
}

	return ok(fmt.Sprintf("%.4f", rate))
}