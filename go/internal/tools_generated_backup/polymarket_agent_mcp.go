package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func HandleGetMarkets(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	closed, _ :=getBool(args, "closed")
	url := fmt.Sprintf("https://clob.polymarket.com/markets?limit=%d&closed=%t", limit, closed)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch markets: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err("unexpected status: " + strconv.Itoa(resp.StatusCode) + " body: " + string(body))
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	jsonResult, e := json.Marshal(result)
	if e != nil {
		return err("failed to marshal result: " + e.Error())
}

	return success(string(jsonResult))
}