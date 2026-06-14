package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetLocationInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	zip, _ :=getString(args, "zip")
	if zip == "" {
		return err("zip is required")
}

	country, _ :=getString(args, "country")
	if country == "" {
		country = "us"
	}
	url := fmt.Sprintf("https://api.zippopotam.us/%s/%s", country, zip)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("status code " + fmt.Sprint(resp.StatusCode))
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Location data: %+v", data))
}

func HandleGetTimezoneInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	tz, _ :=getString(args, "timezone")
	if tz == "" {
		return err("timezone is required")
}

	url := fmt.Sprintf("http://worldtimeapi.org/api/timezone/%s", tz)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("status code " + fmt.Sprint(resp.StatusCode))
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Timezone data: %+v", data))
}