package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

func HandleCallSubagent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	msg, _ :=getString(args, "message")
	if msg == "" {
		msg = "{}"
	}
	resp, e := http.DefaultClient.Post(url, "application/json", strings.NewReader(msg))
	if e != nil {
		return err("failed to call subagent: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return ok(string(body))
}

	return ok(result)
}