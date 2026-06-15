package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleSearchAngularDocs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter required")
}

	u := fmt.Sprintf("https://angular.io/api?search=%s", url.QueryEscape(query))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("search request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read response failed: " + e.Error())
}

	return success(string(body))
}

func HandleGetAngularDoc(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path parameter required")
}

	u := fmt.Sprintf("https://angular.io/%s", path)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("doc request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read doc failed: " + e.Error())
}

	return success(string(body))
}