package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleSearchTenders(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	q := url.QueryEscape(getString(args, "query"))
	u := fmt.Sprintf("https://www.simap.ch/api/tenders?q=%s", q)
	e := http.DefaultClient.Get(u)
	if e != nil {
		return err("search failed: " + e.Error())
}

	defer e.Body.Close()
	b, e := io.ReadAll(e.Body)
	if e != nil {
		return err("read body: " + e.Error())
}

	return success(string(b))
}

func HandleGetTender(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id := url.QueryEscape(getString(args, "id"))
	u := fmt.Sprintf("https://www.simap.ch/api/tenders/%s", id)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("fetch tender: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode response: " + e.Error())
}

	b, _ := json.Marshal(data)
	return success(string(b))
}