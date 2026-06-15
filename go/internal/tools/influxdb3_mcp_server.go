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
	"net/url"
	"strings"
)

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "url")
	if base == "" {
		return err("url required")
}

	database, _ :=getString(args, "database")
	query, _ :=getString(args, "query")
	if database == "" || query == "" {
		return err("database and query required")
}

	token, _ :=getString(args, "token")
	u, e := url.Parse(fmt.Sprintf("%s/api/v2/query?org=%s", base, database))
	if e != nil {
		return err("invalid url: " + e.Error())
}

	body := map[string]string{"query": query, "type": "flux"}
	b, _ := json.Marshal(body)
	req, e := http.NewRequestWithContext(ctx, "POST", u.String(), bytes.NewReader(b))
	if e != nil {
		return err("request error: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Token "+token)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("http error: " + e.Error())
}

	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return err("query failed: " + string(data))
}

	return ok(string(data))
}

}

func HandleWrite(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "url")
	if base == "" {
		return err("url required")
}

	database, _ :=getString(args, "database")
	data, _ :=getString(args, "data")
	if database == "" || data == "" {
		return err("database and data required")
}

	token, _ :=getString(args, "token")
	u, e := url.Parse(fmt.Sprintf("%s/api/v2/write?org=%s&bucket=%s", base, database, database))
	if e != nil {
		return err("invalid url: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", u.String(), strings.NewReader(data))
	if e != nil {
		return err("request error: " + e.Error())
}

	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	if token != "" {
		req.Header.Set("Authorization", "Token "+token)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("http error: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 204 {
		respBody, _ := io.ReadAll(resp.Body)
		return err("write failed: " + string(respBody))
}

	return ok("write successful")
}
}
