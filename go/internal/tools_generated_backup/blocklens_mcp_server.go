package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetTokenBalance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	if address == "" {
		return err("address is required")
}

	url := fmt.Sprintf("https://api.blocklens.io/v1/tokens/%s", address)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch token info: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("failed to parse response")
}

	return success(data)
}