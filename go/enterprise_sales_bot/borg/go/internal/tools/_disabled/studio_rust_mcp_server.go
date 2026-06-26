package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleGetStudioStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(string(body))
}

func HandleExecuteScript(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	script, _ :=getString(args, "script")
	if url == "" || script == "" {
		return err("url and script are required")
}

	resp, e := http.DefaultClient.Post(url, "text/plain", nil)
	if e != nil {
		return err("post failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok("executed: " + string(body))
}