//go:build ignore
// +build ignore

package tools

import (
	"context"
	"fmt"
	"net/http"
	"io"
	"encoding/json"
)

func HandleEcho(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	msg, _ :=getString(args, "message")
	return ok(msg)
}

func HandleArchInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.archlinux.org/version")
	if e != nil {
		return err("failed to fetch arch version")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse json")
}

	return success(fmt.Sprintf("Arch version: %v", data["version"]))
}
