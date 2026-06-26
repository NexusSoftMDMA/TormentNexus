package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetChains(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	chainId, _ :=getString(args, "chainId")
	url := "https://api.debridge.finance/api/SupportedChainList"
	if chainId != "" {
		url = fmt.Sprintf("https://api.debridge.finance/api/SupportedChainInfo?chainId=%s", chainId)

	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch data")
}

	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	return ok(fmt.Sprintf("%v", result))
}

}

func HandleGetQuote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	srcChain, _ :=getString(args, "srcChain")
	srcToken, _ :=getString(args, "srcToken")
	destChain, _ :=getString(args, "destChain")
	destToken, _ :=getString(args, "destToken")
	amount, _ :=getString(args, "amount")
	if srcChain == "" || srcToken == "" || destChain == "" || destToken == "" || amount == "" {
		return err("missing required parameters: srcChain, srcToken, destChain, destToken, amount")
}

	url := fmt.Sprintf("https://api.debridge.finance/api/quote?srcChain=%s&srcToken=%s&destChain=%s&destToken=%s&amount=%s", srcChain, srcToken, destChain, destToken, amount)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch quote")
}

	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	return ok(fmt.Sprintf("%v", result))
}