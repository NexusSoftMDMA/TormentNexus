package mcpimpl

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
)

func HandleMdsFetch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	method, _ :=getString(args, "method")
	if method == "" {
		method = "GET"
	}
	req, e := http.NewRequestWithContext(ctx, method, url, nil)
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
		return err("failed to read body: " + e.Error())
}

	return success(string(body))
}

func HandleMdsProcess(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	input, _ :=getString(args, "input")
	if input == "" {
		return ok("no input provided")
}

	result := strings.ToUpper(input)
	return ok(result)
}