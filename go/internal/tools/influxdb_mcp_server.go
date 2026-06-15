//go:build ignore
// +build ignore

package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url") + "/api/v2/query"
	token, _ :=getString(args, "token")
	org, _ :=getString(args, "org")
	query, _ :=getString(args, "query")
	if query == "" || url == "" {
		return err("missing required args: url and query")
}

	body := map[string]interface{}{"query": query, "type": "flux"}
	if org != "" {
		body["org"] = org
	}
	b, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal query: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(b))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Token "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("http request failed: " + e.Error())
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("influxdb error %d: %s", resp.StatusCode, string(respBody)))
}

	return success(string(respBody))
}

func HandleWrite(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url") + "/api/v2/write?bucket=" + getString(args, "bucket")
	token, _ :=getString(args, "token")
	org, _ :=getString(args, "org")
	if org != "" {
		url += "&org=" + org
	}
	data, _ :=getString(args, "data")
	if data == "" || url == "" {
		return err("missing required args: url, bucket, data")
}

	req, e := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBufferString(data))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Token "+token)
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("http request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("write failed %d: %s", resp.StatusCode, string(body)))
}

	return success("written")
}
