package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleFetchIcal(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("unexpected status: " + resp.Status)
}

	return ok(string(body))
}

func HandleParseIcal(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ics, _ :=getString(args, "ics")
	if ics == "" {
		return err("ics is required")
}

	count := strings.Count(ics, "BEGIN:VEVENT")
	return ok(fmt.Sprintf("Found %d events", count))
}