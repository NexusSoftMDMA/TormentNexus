package mcpimpl

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

func HandleGeocode_magiclane_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	apiKey := os.Getenv("MAGIC_LANE_API_KEY")
	u := fmt.Sprintf("https://api.magiclane.com/v1/geocode?address=%s&api_key=%s", url.QueryEscape(address), apiKey)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("geocode request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}

func HandleReverseGeocode_magiclane_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	lat, _ :=getString(args, "lat")
	lng, _ :=getString(args, "lng")
	apiKey := os.Getenv("MAGIC_LANE_API_KEY")
	u := fmt.Sprintf("https://api.magiclane.com/v1/reverse?lat=%s&lng=%s&api_key=%s", url.QueryEscape(lat), url.QueryEscape(lng), apiKey)
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("reverse geocode request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}