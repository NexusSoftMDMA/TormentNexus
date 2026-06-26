package tools

import (
	"context"
	"fmt"
	"net/http"
)

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	count, _ :=getInt(args, "count")
	if count <= 0 {
		count = 10
	}

	url := fmt.Sprintf("https://openapi.naver.com/v1/search/webkr?query=%s&display=%d", query, count)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("X-Naver-Client-Id", "YOUR_CLIENT_ID")
	req.Header.Set("X-Naver-Client-Secret", "YOUR_CLIENT_SECRET")

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %d", resp.StatusCode))
}

	body, e := http.ReadResponseBody(resp)
	if e != nil {
		return err("read body failed: " + e.Error())
}

	return ok(string(body))
}