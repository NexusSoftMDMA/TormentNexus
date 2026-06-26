package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

func HandleSearchLogs_mcp_server_graylog(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	baseURL, _ :=getString(args, "server_url")
	apiKey, _ :=getString(args, "api_key")
	limit, _ :=getInt(args, "limit")
	if limit < 1 {
		limit = 20
	}
	rangeSec, _ :=getInt(args, "range_seconds")
	if rangeSec < 1 {
		rangeSec = 3600
	}
	u, e := url.Parse(baseURL + "/api/search/universal/relative")
	if e != nil {
		return err("bad url: " + e.Error())
}

	q := u.Query()
	q.Set("query", query)
	q.Set("range", strconv.Itoa(rangeSec))
	q.Set("limit", strconv.Itoa(limit))
	u.RawQuery = q.Encode()
	req, e := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if e != nil {
		return err("req fail: " + e.Error())
}

	req.Header.Set("X-Requested-By", "mcp-server-graylog")
	if apiKey != "" {
		req.Header.Set("Authorization", "Basic "+apiKey)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("http fail: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read fail: " + e.Error())
}

	var tmp interface{}
	if e := json.Unmarshal(body, &tmp); e != nil {
		return err("json bad: " + e.Error())
}

	return ok(string(body))
}

}

func HandleGetMessage_mcp_server_graylog(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msgID, _ :=getString(args, "message_id")
	if msgID == "" {
		return err("message_id required")
}

	baseURL, _ :=getString(args, "server_url")
	apiKey, _ :=getString(args, "api_key")
	u := fmt.Sprintf("%s/api/search/messages/%s", baseURL, msgID)
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return err("req fail: " + e.Error())
}

	req.Header.Set("X-Requested-By", "mcp-server-graylog")
	if apiKey != "" {
		req.Header.Set("Authorization", "Basic "+apiKey)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("http fail: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read fail: " + e.Error())
}

	var tmp interface{}
	if e := json.Unmarshal(body, &tmp); e != nil {
		return err("json bad: " + e.Error())
}

	return ok(string(body))
}
}