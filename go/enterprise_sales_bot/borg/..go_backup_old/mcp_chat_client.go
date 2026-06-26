package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleGeo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	address, _ :=getString(args, "address")
	if address == "" {
		return err("address parameter is required")
	}
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key parameter is required")
	}
	base := "https://restapi.amap.com/v3/geocode/geo"
	params := url.Values{}
	params.Set("address", address)
	params.Set("key", apiKey)
	req, e := http.NewRequestWithContext(ctx, "GET", base+"?"+params.Encode(), nil)
	if e != nil {
		return err(e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(e.Error())
	}
	return success(fmt.Sprintf("Geocode result: %v", result))
}

func HandleReGeo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	location, _ :=getString(args, "location")
	if location == "" {
		return err("location parameter is required")
	}
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key parameter is required")
	}
	base := "https://restapi.amap.com/v3/geocode/regeo"
	params := url.Values{}
	params.Set("location", location)
	params.Set("key", apiKey)
	req, e := http.NewRequestWithContext(ctx, "GET", base+"?"+params.Encode(), nil)
	if e != nil {
		return err(e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(e.Error())
	}
	return success(fmt.Sprintf("ReGeocode result: %v", result))
}