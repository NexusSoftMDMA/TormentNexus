package mcpimpl

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

func HandleFlaiwheel(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	city, _ :=getString(args, "city")
	u := "https://api.flaiwheel.com/v1/stations?city=" + url.QueryEscape(city)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("Failed to fetch: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("Failed to read: " + e.Error())
}

	return success(string(body))
}