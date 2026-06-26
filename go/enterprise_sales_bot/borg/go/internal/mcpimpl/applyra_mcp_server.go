package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetApp_applyra_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	appID, _ :=getString(args, "app_id")
	if appID == "" {
		return err("missing app_id")
}

	url := fmt.Sprintf("https://api.applyra.com/v1/apps/%s", appID)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed")
}

	_, found := args["format"]
	if found {
		var result map[string]interface{}
		if e := json.Unmarshal(body, &result); e != nil {
			return err("parse failed")
}

		return success(string(body))
}

	return ok(string(body))
}

func HandleSearchKeywords(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("missing query")
}

	url := fmt.Sprintf("https://api.applyra.com/v1/keywords?q=%s", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed")
}

	return ok(string(body))
}