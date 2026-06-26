package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleGetCryptoNews(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	category, _ :=getString(args, "category")
	count, _ :=getInt(args, "count")
	if count < 1 || count > 50 {
		count = 10
	}
	url := "https://min-api.cryptocompare.com/data/v2/news/?lang=EN&limit=" + itoa(count)
	if category != "" {
		url += "&categories=" + category
	}
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch news: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON")
}

	if msg, found := result["Message"]; found && msg != "" {
		return err(msg.(string))
}

	raw, _ := json.MarshalIndent(result, "", "  ")
	return ok(string(raw))
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}