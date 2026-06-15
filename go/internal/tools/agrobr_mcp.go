//go:build ignore
// +build ignore

package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleGetWeather(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	loc, _ :=getString(args, "location")
	if loc == "" {
		return err("location required")
}

	resp, e := http.DefaultClient.Get("https://api.weather.com/current?loc=" + loc)
	if e != nil {
		return err("http error: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	return ok(string(body))
}

func HandleGetCropData(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	crop, _ :=getString(args, "crop")
	region, _ :=getString(args, "region")
	if crop == "" || region == "" {
		return err("crop and region required")
}

	resp, e := http.DefaultClient.Get("https://api.agrobr.com/yield?crop=" + crop + "&region=" + region)
	if e != nil {
		return err("http error: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	return ok(string(body))
}
