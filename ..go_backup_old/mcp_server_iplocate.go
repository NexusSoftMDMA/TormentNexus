package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleLookupIP(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ip := strings.TrimSpace(getString(args, "ip"))
	url := "http://ip-api.com/json/"
	if ip != "" {
		url += ip
	}
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to reach ip-api: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var data map[string]interface{}
	if e = json.Unmarshal(body, &data); e != nil {
		return err("invalid JSON response")
}

	status, found := data["status"].(string)
	if !found || status != "success" {
		msg, _ := data["message"].(string)
		return err("API error: " + msg)
}

	parts := []string{
		fmt.Sprintf("IP: %v", data["query"]),
		fmt.Sprintf("Country: %v", data["country"]),
		fmt.Sprintf("Region: %v", data["regionName"]),
		fmt.Sprintf("City: %v", data["city"]),
		fmt.Sprintf("Zip: %v", data["zip"]),
		fmt.Sprintf("Lat: %v", data["lat"]),
		fmt.Sprintf("Lon: %v", data["lon"]),
		fmt.Sprintf("ISP: %v", data["isp"]),
		fmt.Sprintf("Org: %v", data["org"]),
	}
	return ok(strings.Join(parts, "\n"))
}