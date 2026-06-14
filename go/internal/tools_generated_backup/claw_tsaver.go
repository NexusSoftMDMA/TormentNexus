package tools

import (
	"context"
	"net/http"
)

func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	if msg == "" {
		return err("message required")
}

	resp, e := http.DefaultClient.Get("https://httpbin.org/anything?msg=" + msg)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	return ok("sent: " + msg)
}