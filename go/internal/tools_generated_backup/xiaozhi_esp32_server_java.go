package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleXiaozhiEsp32(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	action, _ :=getString(args, "action")
	if baseURL == "" || action == "" {
		return err("base_url and action are required")
}

	url := baseURL + action
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return success(string(body))
}