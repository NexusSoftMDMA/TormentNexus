package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	db, _ :=getString(args, "db")
	user, _ :=getString(args, "user")
	pass, _ :=getString(args, "pass")
	model, _ :=getString(args, "model")
	domain, _ :=getString(args, "domain")
	limit := int64(getInt(args, "limit"))
	if limit == 0 {
		limit = 80
	}
	body := fmt.Sprintf(`{"jsonrpc":"2.0","method":"call","params":{"service":"object","method":"execute_kw","args":["%s",%d,"%s","search_read",%s,{"limit":%d}]}}`, db, 1, user, pass, model, domain, limit)
	req, e := http.NewRequestWithContext(ctx, "POST", url+"/jsonrpc", strings.NewReader(body))
	if e != nil {
		return err("request creation failed: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("http error: " + e.Error())
}

	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var result struct {
		Result []map[string]interface{} `json:"result"`
	}
	if e := json.Unmarshal(raw, &result); e != nil {
		return err("json parse error: " + e.Error())
}

	if result.Result == nil {
		return ok("no records found")
}

	return success(fmt.Sprintf("found %d records", len(result.Result)))
}

func HandleCreate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	db, _ :=getString(args, "db")
	user, _ :=getString(args, "user")
	pass, _ :=getString(args, "pass")
	model, _ :=getString(args, "model")
	data, _ :=getString(args, "data")
	body := fmt.Sprintf(`{"jsonrpc":"2.0","method":"call","params":{"service":"object","method":"execute_kw","args":["%s",%d,"%s","create",[%s]]}}`, db, 1, user, pass, model, data)
	req, e := http.NewRequestWithContext(ctx, "POST", url+"/jsonrpc", strings.NewReader(body))
	if e != nil {
		return err("request creation failed: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("http error: " + e.Error())
}

	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var result struct {
		Result int `json:"result"`
	}
	if e := json.Unmarshal(raw, &result); e != nil {
		return err("json parse error: " + e.Error())
}

	return success(fmt.Sprintf("created record id %d", result.Result))
}