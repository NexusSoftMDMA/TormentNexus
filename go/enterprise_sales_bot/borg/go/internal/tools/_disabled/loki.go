package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

func HandleQueryLogs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	server, _ :=getString(args, "server")
	if server == "" {
		server = "http://localhost:3100"
	}
	u, e := url.Parse(server + "/loki/api/v1/query_range")
	if e != nil {
		return err("invalid server URL: " + e.Error())
}

	q := u.Query()
	q.Set("query", query)
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

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse failed: " + e.Error())
}

	return ok("query result: " + string(body))
}