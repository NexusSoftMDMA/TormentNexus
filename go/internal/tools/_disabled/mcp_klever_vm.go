package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleGetAccount(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	if address == "" {
		return err("address is required")
}

	url := fmt.Sprintf("https://api.klever.io/v1/account/%s", address)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	return ok(string(body))
}

func HandleGetTransaction(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	txid, _ :=getString(args, "txid")
	if txid == "" {
		return err("txid is required")
}

	url := fmt.Sprintf("https://api.klever.io/v1/transaction/%s", txid)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	return ok(string(body))
}