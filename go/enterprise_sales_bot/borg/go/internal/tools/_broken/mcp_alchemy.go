package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleGetTokenBalances(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	if address == "" {
		return err("address is required")
}

	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		apiKey = os.Getenv("ALCHEMY_API_KEY")

	if apiKey == "" {
		return err("api_key arg or ALCHEMY_API_KEY environment variable is required")
}

	chain, _ :=getString(args, "chain")
	if chain == "" {
		chain = "eth-mainnet"
	}
	url := fmt.Sprintf("https://%s.g.alchemy.com/v2/%s/getTokenBalances?owner=%s", chain, apiKey, address)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse failed: %v", e))
}

	return success(string(body))
}
}