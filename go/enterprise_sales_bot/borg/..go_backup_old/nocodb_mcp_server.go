package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListTables(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	apiToken, _ :=getString(args, "api_token")
	if baseURL == "" || apiToken == "" {
		return err("base_url and api_token are required")
}

	url := fmt.Sprintf("%s/api/v1/db/meta/tables", baseURL)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("request creation failed: " + e.Error())
}

	req.Header.Set("xc-token", apiToken)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("json parse failed: " + e.Error())
}

	return ok(fmt.Sprintf("tables: %v", data))
}