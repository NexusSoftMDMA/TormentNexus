package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

func HandleMarketaux(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	symbols, _ :=getString(args, "symbols")
	limit, _ :=getInt(args, "limit")

	u, e := url.Parse("https://api.marketaux.com/v1/news/all")
	if e != nil {
		return err("failed to parse URL: " + e.Error())
}

	q := u.Query()
	q.Set("api_token", apiKey)
	q.Set("symbols", symbols)
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))

	u.RawQuery = q.Encode()

	resp, e := http.DefaultClient.Get(u.String())
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err("API returned " + strconv.Itoa(resp.StatusCode) + ": " + string(body))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("JSON decode error: " + e.Error())
}

	data, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(data))
}
}