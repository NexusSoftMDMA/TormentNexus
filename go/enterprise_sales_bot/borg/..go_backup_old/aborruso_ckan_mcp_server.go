package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// HandleSearchDatasets searches CKAN datasets using package_search
func HandleSearchDatasets(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "url")
	if base == "" {
		return err("missing 'url' argument")
}

	query, _ :=getString(args, "query")
	u, e := url.Parse(base + "/api/3/action/package_search")
	if e != nil {
		return err("invalid URL")
}

	q := u.Query()
	if query != "" {
		q.Set("q", query)

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
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("JSON decode: " + e.Error())
}

	return success(fmt.Sprintf("Search results: %s", string(body)))
}

}

// HandleShowDataset shows a specific CKAN dataset using package_show
func HandleShowDataset(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "url")
	if base == "" {
		return err("missing 'url' argument")
}

	id, _ :=getString(args, "id")
	if id == "" {
		return err("missing 'id' argument")
}

	u, e := url.Parse(base + "/api/3/action/package_show")
	if e != nil {
		return err("invalid URL")
}

	q := u.Query()
	q.Set("id", id)
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
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("JSON decode: " + e.Error())
}

	return success(fmt.Sprintf("Dataset: %s", string(body)))
}