package tools

import (
	"context"
	"io"
	"net/http"
	"strconv"
)

func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	username, _ :=getString(args, "username")
	if username == "" {
		return err("username is required")
}

	url := "https://x.com/" + username
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("unexpected status: " + resp.Status)
}

	body, _ := io.ReadAll(resp.Body)
	return success("Fetched profile for @" + username + " (" + strconv.Itoa(len(body)) + " bytes)")
}