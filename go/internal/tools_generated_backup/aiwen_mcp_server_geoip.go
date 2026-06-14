package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetGeoIP(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ip, _ :=getString(args, "ip")
	if ip == "" {
		return err("ip parameter is required")
}

	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=country,regionName,city,isp,query", ip)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch geoip data: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse geoip data: " + e.Error())
}

	if status, found := result["status"].(string); found && status == "fail" {
		if msg, found := result["message"].(string); found {
			return err("geoip error: " + msg)
}

		return err("geoip lookup failed")
}

	country, _ := result["country"].(string)
	region, _ := result["regionName"].(string)
	city, _ := result["city"].(string)
	isp, _ := result["isp"].(string)
	ipAddr, _ := result["query"].(string)
	msg := fmt.Sprintf("IP: %s\nCountry: %s\nRegion: %s\nCity: %s\nISP: %s", ipAddr, country, region, city, isp)
	return ok(msg)
}