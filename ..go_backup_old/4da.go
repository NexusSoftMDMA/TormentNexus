package tools

import (
	"context"
	"net/http"
)

func Handle4Da(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return err("no message provided")
}

	resp, e := http.DefaultClient.Get("https://httpbin.org/get?msg=" + msg)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	return ok("sent: " + msg)
}