package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleGetAccount(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "public_key")
	if key == "" {
		return err("public_key is required")
}

	u := fmt.Sprintf("https://horizon-testnet.stellar.org/accounts/%s", url.PathEscape(key))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("failed to fetch account: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("horizon returned status %d", resp.StatusCode))
}

	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to decode response: " + e.Error())
}

	return ok(fmt.Sprintf("Account: %s", key))
}