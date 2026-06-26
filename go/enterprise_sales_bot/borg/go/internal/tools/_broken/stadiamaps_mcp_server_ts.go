package mcpimpl

import (
	"encoding/json"
	"io"
	"net/http"
)

func HandleForwardGeocode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	apiKey, _ :=getString(args, "api_key")
	if query == "" || apiKey == "" {
		return err("missing query or api_key")
}

	url := "https://api.stadiamaps.com/geocoding/v1/forward?q=" + query + "&api_key=" + apiKey
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("http error: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	var result map[string]interface{}
	e = json.Unmarshal(body, &result)
	if e != nil {
		return err("json parse error: " + e.Error())
}

	return ok(string(body))
}

func HandleReverseGeocode_stadiamaps_mcp_server_ts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	lat, _ :=getString(args, "lat")
	lon, _ :=getString(args, "lon")
	apiKey, _ :=getString(args, "api_key")
	if lat == "" || lon == "" || apiKey == "" {
		return err("missing lat, lon, or api_key")
}

	url := "https://api.stadiamaps.com/geocoding/v1/reverse?lat=" + lat + "&lon=" + lon + "&api_key=" + apiKey
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("http error: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	var result map[string]interface{}
	e = json.Unmarshal(body, &result)
	if e != nil {
		return err("json parse error: " + e.Error())
}

	return ok(string(body))
}