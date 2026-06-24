package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleCbi(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	target, _ :=getString(args, "target")
	if target == "" {
		return err("target is required")
}

	resp, e := http.DefaultClient.Get(target)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(string(body))
}