package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetMarket(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	marketId, _ :=getString(args, "marketId")
	if marketId == "" {
		return err("marketId is required")
}

	url := fmt.Sprintf("https://clob.polymarket.com/markets/%s", marketId)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch market: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse JSON")
}

	return ok(fmt.Sprintf("%+v", data))
}