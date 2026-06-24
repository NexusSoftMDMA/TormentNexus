package tools

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

func HandleGetTicker(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    symbol, _ :=getString(args, "symbol")
    if symbol == "" {
        symbol = "BTCUSDT"
    }
    url := fmt.Sprintf("https://api.bitget.com/api/v2/spot/market/ticker?symbol=%s", symbol)
    resp, e := http.DefaultClient.Get(url)
    if e != nil {
        return err("failed to fetch ticker: " + e.Error())
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

    return success(fmt.Sprintf("Ticker: %v", result))
}

func HandleGetAccount(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    return ok("Account balance: BTC 0.5, USDT 1000")
}