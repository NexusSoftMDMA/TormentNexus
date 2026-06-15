package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleListAdvertisers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	apiKey, _ :=getString(args, "api_key")
	limit, _ :=getInt(args, "limit")
	offset, _ :=getInt(args, "offset")
	if baseURL == "" || apiKey == "" {
		return err("base_url and api_key required")
}

	url := fmt.Sprintf("%s/advertisers?limit=%d&offset=%d", baseURL, limit, offset)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("request error: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("json error: " + e.Error())
}

	return success(string(body))
}