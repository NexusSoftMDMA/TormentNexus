package tools

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
)

func HandleExecuteCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	cmd, _ :=getString(args, "command")
	if url == "" || cmd == "" {
		return err("missing url or command")
}

	req, e := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(cmd))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "text/plain")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err("server returned " + resp.Status + ": " + string(body))
}

	return ok(string(body))
}

func HandleListFiles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	path, _ :=getString(args, "path")
	if url == "" {
		return err("missing url")
}

	if path != "" {
		url = url + "?path=" + path
	}
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err("server returned " + resp.Status + ": " + string(body))
}

	return ok(string(body))
}