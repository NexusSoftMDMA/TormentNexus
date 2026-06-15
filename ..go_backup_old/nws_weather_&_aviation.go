package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetGridForecast(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	gridId, _ :=getString(args, "gridId")
	gridX, _ :=getString(args, "gridX")
	gridY, _ :=getString(args, "gridY")
	if gridId == "" || gridX == "" || gridY == "" {
		return err("gridId, gridX, and gridY are required")
}

	url := fmt.Sprintf("https://api.weather.gov/gridpoints/%s/%s,%s/forecast", gridId, gridX, gridY)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("HTTP request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("API error: " + resp.Status)
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body failed: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("json parse failed: " + e.Error())
}

	properties, found := result["properties"].(map[string]interface{})
	if !found {
		return err("missing properties")
}

	periods, found := properties["periods"].([]interface{})
	if !found {
		return err("missing periods")
}

	data, _ := json.MarshalIndent(periods, "", "  ")
	return ok(string(data))
}

func HandleGetAlerts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	area, _ :=getString(args, "area")
	if area == "" {
		return err("area is required (e.g., state code)")
}

	url := fmt.Sprintf("https://api.weather.gov/alerts/active?area=%s", area)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("HTTP request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("API error: " + resp.Status)
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body failed: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("json parse failed: " + e.Error())
}

	features, found := result["features"].([]interface{})
	if !found {
		return err("missing features")
}

	data, _ := json.MarshalIndent(features, "", "  ")
	return ok(string(data))
}