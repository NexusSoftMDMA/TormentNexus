package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleListTables(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", url+"/tables", nil)
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

	if resp.StatusCode != 200 {
		return err("bad status: " + resp.Status)
}

	var result map[string]interface{}
	json.Unmarshal(body, &result)
	msg, _ := json.Marshal(result)
	return ok(string(msg))
}

func HandleExecuteQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	sql, _ :=getString(args, "sql")
	if url == "" || sql == "" {
		return err("url and sql are required")
}

	reqBody, _ := json.Marshal(map[string]string{"sql": sql})
	req, e := http.NewRequestWithContext(ctx, "POST", url+"/query", nil)
	if e != nil {
		return err(e.Error())
}

	req.Body = io.NopCloser(io.Reader(bytes.NewReader(reqBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	if resp.StatusCode != 200 {
		return err("bad status: " + resp.Status)
}

	return ok(string(body))
}