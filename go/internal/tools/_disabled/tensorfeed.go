package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleGetTrending(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit < 1 {
		limit = 10
	}
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.tensorfeed.com/v1/trending?limit="+itoa(limit), nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch trending")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return success(string(body))
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	s := ""
	for i > 0 {
		s = string(rune('0'+i%10)) + s
		i /= 10
	}
	return s
}