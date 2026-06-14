package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

func HandleListDevices(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	baseURL, _ :=getString(args, "baseURL")
	if baseURL == "" {
		baseURL = "https://api.mobitru.com"
	}
	url := baseURL + "/devices"
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("request error: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	return ok(string(body))
}