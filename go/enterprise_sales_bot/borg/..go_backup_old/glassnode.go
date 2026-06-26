package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetOnChainMetrics(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	asset, _ :=getString(args, "asset")
	if asset == "" {
		asset = "BTC"
	}
	metric, _ :=getString(args, "metric")
	if metric == "" {
		metric = "marketcap"
	}
	url := fmt.Sprintf("https://api.glassnode.com/v1/metrics/%s?api_key=%s&a=%s", metric, apiKey, asset)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %d", resp.StatusCode))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("parse failed: %v", e))
}

	return success(fmt.Sprintf("%+v", data))
}