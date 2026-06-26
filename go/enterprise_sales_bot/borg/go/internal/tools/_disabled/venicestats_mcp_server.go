package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleTokenStats(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	if token == "" {
		token = "VVV"
	}
	url := fmt.Sprintf("https://api.venicestats.com/v1/token/stats?token=%s", token)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("json parse failed: %v", e))
}

	return ok(fmt.Sprintf("Token stats: %+v", data))
}

func HandlePoolStats(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pool, _ :=getString(args, "pool")
	if pool == "" {
		pool = "DIEM/USDC"
	}
	url := fmt.Sprintf("https://api.venicestats.com/v1/pool/stats?pool=%s", pool)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("json parse failed: %v", e))
}

	return ok(fmt.Sprintf("Pool stats: %+v", data))
}