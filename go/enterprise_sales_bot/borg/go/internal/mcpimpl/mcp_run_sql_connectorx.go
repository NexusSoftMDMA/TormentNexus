package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleRunSql(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	urlStr, _ :=getString(args, "url")
	if urlStr == "" {
		return err("missing 'url' argument")
}

	u, e := url.Parse(urlStr)
	if e != nil {
		return err("invalid url: " + e.Error())
}

	q := u.Query()
	q.Set("sql", query)
	u.RawQuery = q.Encode()
	req, e := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("status %d", resp.StatusCode))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("reading response: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e == nil {
		return success(result)
}

	return success(string(body))
}