package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleFetchUrl(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("HTTP request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(string(body))
}

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	var result string
	for _, v := range args {
		result += fmt.Sprintf("%v ", v)

	return ok(strings.TrimSpace(result))
}
}