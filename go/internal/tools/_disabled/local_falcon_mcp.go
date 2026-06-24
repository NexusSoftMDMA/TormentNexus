package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleGetReports(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.localfalcon.com/v1/reports?api_key="+apiKey, nil)
	if e != nil {
		return err("request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("status: " + resp.Status)
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read: " + e.Error())
}

	return ok(string(body))
}