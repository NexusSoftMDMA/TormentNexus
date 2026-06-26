package mcpimpl

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleTokenInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	if address == "" {
		return err("address is required")
}

	return doGet(ctx, "https://api.kaito.ai/v1/token?address="+address)
}

func HandleTrending(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit == 0 {
		limit = 10
	}
	url := fmt.Sprintf("https://api.kaito.ai/v1/trending?limit=%d", limit)
	return doGet(ctx, url)
}

func doGet_kaito_mcp_server(ctx context.Context, url string) (ToolResponse, error) {
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to execute request: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("API returned status " + resp.Status)
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return success(string(body))
}