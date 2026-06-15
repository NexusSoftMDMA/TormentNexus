//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type geoResponse struct {
	Status  string  `json:"status"`
	Country string  `json:"country"`
	City    string  `json:"city"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	Query   string  `json:"query"`,
}

func HandleGeoIP(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ip, _ :=getString(args, "ip")
	url := "http://ip-api.com/json/" + ip
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var data geoResponse
	if e := json.Unmarshal(body, &data); e != nil {
		return err("unmarshal failed: " + e.Error())
	if data.Status != "success" {
		return err("API error: " + data.Status)
	return ok(fmt.Sprintf("IP: %s, City: %s, Country: %s, Lat: %.4f, Lon: %.4f",
}
		data.Query, data.City, data.Country, data.Lat, data.Lon))
}
}