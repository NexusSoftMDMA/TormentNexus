//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func HandleReadFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path is required")
}

	serverURL, _ :=getString(args, "server_url")
	req, e := http.NewRequestWithContext(ctx, "GET", serverURL+"/read?path="+path, nil)
	if e != nil {
		return err(fmt.Sprintf("create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("http call: %v", e))
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body: %v", e))
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("server error %d: %s", resp.StatusCode, string(body)))
	return ok(string(body))
}

func HandleWriteFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path is required")
}

	content, _ :=getString(args, "content")
	serverURL, _ :=getString(args, "server_url")
	payload := fmt.Sprintf(`{"path":"%s","content":"%s"}`, path, strings.ReplaceAll(content, `"`, `\"`))
	req, e := http.NewRequestWithContext(ctx, "POST", serverURL+"/write", strings.NewReader(payload))
	if e != nil {
		return err(fmt.Sprintf("create request: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("http call: %v", e))
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body: %v", e))
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("server error %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse response: %v", e))
}

	msg, found := result["message"].(string)
	if !found {
		return err("missing message in response")
	return ok(msg)
}
}
}
}
}
}