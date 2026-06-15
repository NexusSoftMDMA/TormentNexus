package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleVaultInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	vaultID, _ :=getString(args, "vaultId")
	if vaultID == "" {
		return err("vaultId is required")
}

	chain, _ :=getString(args, "chain")
	if chain == "" {
		return err("chain is required")
}

	url := fmt.Sprintf("https://api.usdd.com/v1/vaults/%s?chain=%s", vaultID, chain)
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode error: " + e.Error())
}

	return ok(fmt.Sprintf("Vault info: %v", data))
}

func HandlePSMSwapQuote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	from, _ :=getString(args, "from")
	to, _ :=getString(args, "to")
	amount, _ :=getString(args, "amount")
	if from == "" || to == "" || amount == "" {
		return err("from, to, amount are required")
}

	chain, _ :=getString(args, "chain")
	url := fmt.Sprintf("https://api.usdd.com/v1/psm/quote?from=%s&to=%s&amount=%s&chain=%s", from, to, amount, chain)
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode error: " + e.Error())
}

	return ok(fmt.Sprintf("PSM swap quote: %v", data))
}