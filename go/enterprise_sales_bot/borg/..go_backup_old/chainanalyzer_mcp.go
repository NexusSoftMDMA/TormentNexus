package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetChainInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	chain, _ :=getString(args, "chain")
	if chain == "" {
		chain = "ethereum"
	}
	url := fmt.Sprintf("https://api.chainanalyzer.com/chain/%s", chain)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch chain info: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("Chain %s: %v", chain, data))
}

func HandleGetTransaction(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	txHash, _ :=getString(args, "txHash")
	if txHash == "" {
		return err("txHash is required")
}

	chain, _ :=getString(args, "chain")
	if chain == "" {
		chain = "ethereum"
	}
	url := fmt.Sprintf("https://api.chainanalyzer.com/tx/%s?chain=%s", txHash, chain)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch transaction: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return ok(fmt.Sprintf("Transaction %s: %v", txHash, data))
}