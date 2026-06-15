package tools

import (
	"context"
	"net/http"
	"strings"
)

func HandleSSEPing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
	}
	req, e := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader("data: ping\n\n"))
	if e != nil {
		return err("failed to create request: " + e.Error())
	}
	req.Header.Set("Content-Type", "text/event-stream")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("unexpected status: " + resp.Status)
	}
	return success("pong received")
}

func HandleSSESend(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	event, _ :=getString(args, "event")
	data, _ :=getString(args, "data")
	if url == "" || data == "" {
		return err("url and data are required")
	}
	body := "data: " + data + "\n"
	if event != "" {
		body = "event: " + event + "\n" + body
	}
	body += "\n"
	req, e := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
	}
	req.Header.Set("Content-Type", "text/event-stream")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("unexpected status: " + resp.Status)
	}
	return ok("event sent")
}