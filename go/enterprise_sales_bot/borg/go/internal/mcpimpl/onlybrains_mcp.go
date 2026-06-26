package mcpimpl

import (
	"context"
	"net/http"
)

func HandleTrade(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	amount, _ :=getInt(args, "amount")
	if token == "" {
		return err("token required")
}

	url := "https://api.onlybrains.com/trade?token=" + token + "&amount=" + string(amount)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("api error: " + e.Error())
}

	resp.Body.Close()
	return ok("trade executed")
}

func HandleNotarize(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	data, _ :=getString(args, "data")
	if data == "" {
		return err("data required")
}

	url := "https://api.onlybrains.com/notarize?data=" + data
	resp, e := http.DefaultClient.Post(url, "text/plain", nil)
	if e != nil {
		return err("notarization failed: " + e.Error())
}

	resp.Body.Close()
	return success("data notarized")
}