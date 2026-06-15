//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleBankSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	u := fmt.Sprintf("https://banks.data.fdic.gov/api/institutions?filters=NAME:%%%s%%&limit=5", url.QueryEscape(query))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result map[string]interfact{	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse failed: " + e.Error())
}

	data, found := result["data"]
	if !found {
		return ok("[]")
}

	out, _ := json.Marshal(data)
	return ok(string(out))
}
}