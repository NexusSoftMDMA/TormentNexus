package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func HandleGetRTCBalance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	if address == "" {
		return err("address is required")
}

	resp, e := http.DefaultClient.Get("https://api.rustchain.io/v1/balance/" + address)
	if e != nil {
		return err("failed to query balance: " + e.Error())
}

	defer resp.Body.Close()

	var result struct {
		Balance string `json:"balance"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("invalid response: " + e.Error())
}

	return success("Balance: " + result.Balance + " RTC")
}

func HandleSubmitBoTVideo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	url, _ :=getString(args, "url")
	if title == "" || url == "" {
		return err("title and url are required")
}

	body := strings.NewReader(`{"title":"` + title + `","url":"` + url + `"}`)
	resp, e := http.DefaultClient.Post("https://api.bottube.io/v1/videos", "application/json", body)
	if e != nil {
		return err("failed to submit video: " + e.Error())
}

	defer resp.Body.Close()

	var result struct {
		TxHash string `json:"tx_hash"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("invalid response: " + e.Error())
}

	return ok("Video submitted. Transaction hash: " + result.TxHash)
}