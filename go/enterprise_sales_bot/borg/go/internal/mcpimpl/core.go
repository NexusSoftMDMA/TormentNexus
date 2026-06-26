package mcpimpl

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleX_core(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to GET: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body: " + e.Error())
}

	return ok(fmt.Sprintf("Status: %s\nBody:\n%s", resp.Status, string(body)))
}