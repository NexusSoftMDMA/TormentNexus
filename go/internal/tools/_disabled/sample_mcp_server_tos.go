package tools

import (
	"context"
	"net/http"
)

// HandleGreet returns a greeting message.
func HandleGreet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "World"
	}
	msg := "Hello, " + name + "!"
	return ok(msg)
}

// HandleFetchPage fetches a URL and returns its content length (or error).
func HandleFetchPage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("missing url")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("fetch failed: " + e.Error())
}

	defer resp.Body.Close()
	body := make([]byte, 1024)
	n, e := resp.Body.Read(body)
	if e != nil && e.Error() != "EOF" {
		return err("read failed: " + e.Error())
}

	return ok("fetched " + string(body[:n]))
}