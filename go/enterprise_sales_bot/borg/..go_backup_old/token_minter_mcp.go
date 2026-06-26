package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
)

func HandleMintToken(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	amount, _ :=getInt(args, "amount")
	if address == "" || amount <= 0 {
		return err("address and amount are required")
}

	reqBody, _ := json.Marshal(map[string]interface{}{"address": address, "amount": amount})
	resp, e := http.DefaultClient.Post(os.Getenv("MINT_ENDPOINT"), "application/json", reqBody)
	if e != nil {
		return err("mint request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("mint returned status " + resp.Status)
}

	io.Copy(io.Discard, resp.Body)
	return ok("token minted successfully")
}

func HandleGetBalance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	if address == "" {
		return err("address is required")
}

	resp, e := http.DefaultClient.Get(os.Getenv("BALANCE_ENDPOINT") + "?address=" + address)
	if e != nil {
		return err("balance request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("balance returned status " + resp.Status)
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("invalid response")
}

	return success("balance retrieved", result)
}