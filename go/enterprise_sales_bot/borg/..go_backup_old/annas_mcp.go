package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleAnnaSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	q, _ :=getString(args, "query")
	if q == "" {
		return err("query is required")
}

	u := fmt.Sprintf("https://annas-archive.org/search?q=%s", url.QueryEscape(q))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("search failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(string(body))
}

func HandleAnnaDownload(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	u := fmt.Sprintf("https://annas-archive.org/download/%s", url.PathEscape(id))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("download failed: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("parse failed: " + e.Error())
}

	link, found := data["link"].(string)
	if !found {
		return err("no link in response")
}

	return ok(link)
}