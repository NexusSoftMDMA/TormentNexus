package mcpimpl

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
)

func HandleProxy_mcpo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	method, _ :=getString(args, "method")
	if method == "" {
		method = "GET"
	}
	body, _ :=getString(args, "body")

	var req *http.Request
	var e error
	if body != "" {
		req, e = http.NewRequestWithContext(ctx, method, url, bytes.NewBufferString(body))
	} else {
		req, e = http.NewRequestWithContext(ctx, method, url, nil)

	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()

	respBody, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return success(string(respBody))
}
}