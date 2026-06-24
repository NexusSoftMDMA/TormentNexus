package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetLiquidity(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	poolId, _ :=getString(args, "pool_id")
	if poolId == "" {
		return err("pool_id is required")
}

	url := fmt.Sprintf("https://api.near.org/v1/pools/%s", poolId)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("parse failed: %v", e))
}

	return ok(fmt.Sprintf("Liquidity for pool %s: %v", poolId, data))
}