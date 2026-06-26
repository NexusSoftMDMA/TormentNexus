package mcpimpl

import (
	"context"
	"io"
	"net/http"
	"strconv"
)

func HandleListRoutes_apisix(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		return err("base_url is required")
}

	resp, e := http.DefaultClient.Get(baseURL + "/apisix/admin/routes")
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body failed: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("API error: " + strconv.Itoa(resp.StatusCode))
}

	return ok(string(body))
}