package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleMintNFT(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	symbol, _ :=getString(args, "symbol")
	url := "https://api.verbwire.com/v1/nft/mint?name=" + name + "&symbol=" + symbol
	req, e := http.NewRequestWithContext(ctx, "POST", url, nil)
	if e != nil {
		return err("request failed: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("json parse failed: " + e.Error())
}

	return ok(fmt.Sprintf("Minted NFT: %v", result["transaction_hash"]))
}

func HandleGetBalance_verbwire_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	url := "https://api.verbwire.com/v1/account/balance?address=" + address
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("request failed: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(fmt.Sprintf("Balance: %s", string(body)))
}