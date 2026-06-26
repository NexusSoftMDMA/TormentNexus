package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleVictorialogsQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query argument is required")
	}
	base, _ :=getString(args, "base_url")
	if base == "" {
		base = "http://localhost:9428"
	}
	u, e := url.Parse(base + "/select/0/prometheus/api/v1/query")
	if e != nil {
		return err(e.Error())
	}
	q := u.Query()
	q.Set("query", query)
	u.RawQuery = q.Encode()
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if e != nil {
		return err(e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("status %d: %s", resp.StatusCode, string(body)))
	}
	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(e.Error())
	}
	return ok(string(body))
}

func HandleVictorialogsLabels(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base_url")
	if base == "" {
		base = "http://localhost:9428"
	}
	u, e := url.Parse(base + "/api/v1/labels")
	if e != nil {
		return err(e.Error())
	}
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if e != nil {
		return err(e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("status %d: %s", resp.StatusCode, string(body)))
	}
	return ok(string(body))
}// touch 1781132134
