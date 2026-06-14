package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleStitchRequest(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("missing url")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	var result interface{}
	e = json.NewDecoder(resp.Body).Decode(&result)
	if e != nil {
		return err(e.Error())
	}
	return success("request completed")
}

func HandleStitchPost(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	body, _ :=getString(args, "body")
	if url == "" {
		return err("missing url")
	}
	req, e := http.NewRequestWithContext(ctx, "POST", url, nil)
	if e != nil {
		return err(e.Error())
	}
	if body != "" {
		req.Body = nil
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	return ok("post executed")
}// touch 1781132131
