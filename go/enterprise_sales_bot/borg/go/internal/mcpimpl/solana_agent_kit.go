package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func GetSolBalance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	if address == "" {
		return err("address is required")
}

	payload := strings.NewReader(fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"method":"getBalance","params":["%s"]}`, address))
	resp, e := http.DefaultClient.Post("https://api.mainnet-beta.solana.com", "application/json", payload)
	if e != nil {
		return err(fmt.Sprintf("RPC call failed: %v", e))
}

	defer resp.Body.Close()
	var result struct {
		Result struct {
			Value int64 `json:"value"`
		} `json:"result"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	return success(fmt.Sprintf("Balance: %d lamports", result.Result.Value))
}

func GetSolPrice(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	currency, _ :=getString(args, "currency")
	if currency == "" {
		currency = "usd"
	}
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=solana&vs_currencies=%s", currency)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("price request failed: %v", e))
}

	defer resp.Body.Close()
	var data map[string]map[string]float64
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	solData, found := data["solana"]
	if !found {
		return err("solana price not found")
}

	price, found := solData[currency]
	if !found {
		return err(fmt.Sprintf("price in %s not found", currency))
}

	return success(fmt.Sprintf("SOL price: %.2f %s", price, strings.ToUpper(currency)))
}