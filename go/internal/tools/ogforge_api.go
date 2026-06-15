//go:build ignore
// +build ignore

package tools

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func HandleFetchOgTitle(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("request error: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("fetch error: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	title := ""
	start := strings.Index(string(body), "<title>")
	if start != -1 {
		end := strings.Index(string(body[start:]), "</title>")
		if end != -1 {

},
}
}