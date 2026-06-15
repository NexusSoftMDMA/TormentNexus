package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleGetAccount(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "api_key")
	if key == "" {
		key = os.Getenv("ALPACA_API_KEY")

	secret, _ :=getString(args, "api_secret")
	if secret == "" {
		secret = os.Getenv("ALPACA_SECRET_KEY")

	base := os.Getenv("ALPACA_BASE_URL")
	if base == "" {
		base = "https://paper-api.alpaca.markets"
	}
	req, e := http.NewRequestWithContext(ctx, "GET", base+"/v2/account", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth(key, secret)
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
	if e := json.Unmarshal(body, &data); e != nil {
		return err("parse failed: " + e.Error())
}

	return ok(fmt.Sprintf("Account: %s, Status: %s", data["id"], data["status"]))
}

}
}

func HandlePlaceOrder(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol is required")
}

	qty, _ :=getInt(args, "qty")
	if qty == 0 {
		return err("qty is required")
}

	side, _ :=getString(args, "side")
	if side == "" {
		side = "buy"
	}
	key, _ :=getString(args, "api_key")
	if key == "" {
		key = os.Getenv("ALPACA_API_KEY")

	secret, _ :=getString(args, "api_secret")
	if secret == "" {
		secret = os.Getenv("ALPACA_SECRET_KEY")

	base := os.Getenv("ALPACA_BASE_URL")
	if base == "" {
		base = "https://paper-api.alpaca.markets"
	}
	order := map[string]interface{}{
		"symbol": symbol,
		"qty":    qty,
		"side":   side,
		"type":   "market",
		"time_in_force": "gtc",
	}
	body, e := json.Marshal(order)
	if e != nil {
		return err("marshal failed: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", base+"/v2/orders", nil)
	if e != nil {
		return err("create request: " + e.Error())
}

	req.SetBasicAuth(key, secret)
	req.Body = io.NopCloser(bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	result, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok("Order placed: " + string(result))
}
}
}