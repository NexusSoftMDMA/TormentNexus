package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetStock(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("missing symbol")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.fugle.tw/realtime/v1.0/stock/"+symbol, nil)
	if e != nil {
		return err(e.Error())
}

	req.Header.Set("Authorization", "Bearer "+getString(args, "token"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API error: %s", string(body)))
}

	return ok(string(body))
}

func HandleGetMarket(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	market, _ :=getString(args, "market")
	if market == "" {
		market = "TWSE"
	}
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.fugle.tw/realtime/v1.0/market/"+market, nil)
	if e != nil {
		return err(e.Error())
}

	req.Header.Set("Authorization", "Bearer "+getString(args, "token"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid json response")
}

	return success("market data retrieved")
}