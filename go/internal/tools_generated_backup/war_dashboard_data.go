package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleGetWarDashboard(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		url = "https://example.com/api/war-dashboard"
	}
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch data: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("server returned status " + resp.Status)
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return success(string(body))
}