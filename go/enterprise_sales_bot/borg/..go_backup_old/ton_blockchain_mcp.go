package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func GetAccountBalance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	if address == "" {
		return err("address is required")
}

	url := fmt.Sprintf("https://toncenter.com/api/v2/getAccount?address=%s", address)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	var result struct {
		Ok     bool `json:"ok"`
		Result struct {
			Balance int64 `json:"balance"`
		} `json:"result"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse json failed: %v", e))
}

	if !result.Ok {
		return err("API response not ok")
}

	return ok(fmt.Sprintf("Balance: %d", result.Result.Balance))
}

func GetTransactionByHash(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	hash, _ :=getString(args, "hash")
	if hash == "" {
		return err("hash is required")
}

	url := fmt.Sprintf("https://toncenter.com/api/v2/getTransaction?hash=%s", hash)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	var result struct {
		Ok     bool            `json:"ok"`
		Result json.RawMessage `json:"result"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse json failed: %v", e))
}

	if !result.Ok {
		return err("API response not ok")
}

	return ok(fmt.Sprintf("Transaction data: %s", string(result.Result)))
}