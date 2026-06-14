package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleCreateMultiSigWallet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ownersStr, _ :=getString(args, "owners")
	threshold, _ :=getInt(args, "threshold")
	chain, _ :=getString(args, "chain")

	body := map[string]interface{}{
		"owners":    ownersStr,
		"threshold": threshold,
		"chain":     chain,
	}
	jsonBody, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request")
}

	url := "https://api.multisig.example.com/create"
	req, e := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Content-Type", "application/json")

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	address, found := result["address"].(string)
	if !found {
		return err("response missing address")
}

	return success("Multi-sig wallet created: " + address)
}