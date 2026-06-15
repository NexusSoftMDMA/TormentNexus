//go:build ignore
// +build ignore

package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleListServers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filter, _ :=getString(args, "filter")
	url := "https://raw.githubusercontent.com/suchipi/awesome-mcp-servers/main/README.md"

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch: " + e.Error())
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read: " + e.Error())
}

	msg := string(body)
	if filter != "" {
		msg = "Filter: " + filter + "\n\n" + msg
	}
	return ok(msg)
}
