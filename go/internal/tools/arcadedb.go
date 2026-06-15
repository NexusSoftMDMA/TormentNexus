//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	query, _ :=getString(args, "query")
	reqURL := fmt.Sprintf("%s?query=%s", url, query)
	resp, e := http.DefaultClient.Get(reqURL)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("json error: %v", e))
}

	return ok(fmt.Sprintf("result: %v", result))
}

func HandleExecute(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	command, _ :=getString(args, "command")
	reqURL := fmt.Sprintf("%s/execute", url)
	req, e := http.NewRequestWithContext(ctx, "POST", reqURL, nil)
	if e != nil {
		return err(fmt.Sprintf("create request: %v", e))
}

	req.Body = io.NopCloser(strReader(command))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("execute failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	return ok(fmt.Sprintf("executed: %s", string(body)))
}

func strReader(s string) io.Reader {
	return &stringReader{s}
}

type stringReader struct{ s string }

func (r *stringReader) Read(p []byte) (int, error) {
	n := copy(p, r.s)
	r.s = r.s[n:]
	if len(r.s) == 0 {
		return n, io.EOF
	}
	return n, nil
}
