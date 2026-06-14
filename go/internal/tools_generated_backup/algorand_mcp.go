package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetAlgorandAccount(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	if address == "" {
		return err("address is required")
}

	url := fmt.Sprintf("https://testnet-algorand.api.purestake.io/idx2/v2/accounts/%s", address)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("x-api-key", getString(args, "apiKey"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch account: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var data map[string]interface{}
	if e = json.Unmarshal(body, &data); e != nil {
		return err("invalid JSON: " + e.Error())
}

	out, _ := json.MarshalIndent(data, "", "  ")
	return success(string(out))
}