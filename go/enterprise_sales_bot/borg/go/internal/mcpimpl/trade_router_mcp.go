package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func HandleGetTradeRoutes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	from, _ :=getString(args, "from")
	to, _ :=getString(args, "to")
	url := "https://api.example.com/trades?from=" + from + "&to=" + to
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch routes: " + e.Error())
}

	defer resp.Body.Close()
	var result interface{}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err("failed to decode response")
}

	return success(result)
}

func HandleExecuteTrade_trade_router_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	from, _ :=getString(args, "from")
	to, _ :=getString(args, "to")
	amount, _ :=getInt(args, "amount")
	body := strings.NewReader(`{"from":"` + from + `","to":"` + to + `","amount":` + string(amount) + `}`)
	resp, e := http.DefaultClient.Post("https://api.example.com/trade", "application/json", body)
	if e != nil {
		return err("failed to execute trade: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return err("trade execution failed with status " + resp.Status)
}

	var result interface{}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err("failed to decode response")
}

	return success(result)
}