package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

func HandleMeilisearchSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	apiKey, _ :=getString(args, "api_key")
	index, _ :=getString(args, "index")
	query, _ :=getString(args, "query")
	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 20
	}
	u, e := url.Parse(host + "/indexes/" + index + "/search")
	if e != nil {
		return err("invalid host: " + e.Error())
}

	q := u.Query()
	q.Set("q", query)
	q.Set("limit", strconv.Itoa(limit))
	u.RawQuery = q.Encode()
	req, e := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("Meilisearch returned status " + strconv.Itoa(resp.StatusCode) + ": " + string(body))
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	out, _ := json.Marshal(result)
	return ok(string(out))
}