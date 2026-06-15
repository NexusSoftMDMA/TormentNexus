package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleGetInstallData(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	appID, _ :=getString(args, "app_id")
	deviceID, _ :=getString(args, "device_id")
	if apiKey == "" || appID == "" || deviceID == "" {
		return err("missing api_key, app_id, or device_id")
}

	u := fmt.Sprintf("https://api.appsflyer.com/raw-data/v1.0/install/%s?device_id=%s", appID, url.QueryEscape(deviceID))
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return err("request creation failed: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var data interface{}
	if e = json.Unmarshal(body, &data); e != nil {
		return err("parse failed: " + e.Error())
}

	return success(fmt.Sprintf("install data: %v", data))
}

func HandleGetConversionData(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	appID, _ :=getString(args, "app_id")
	deviceID, _ :=getString(args, "device_id")
	if apiKey == "" || appID == "" || deviceID == "" {
		return err("missing api_key, app_id, or device_id")
}

	u := fmt.Sprintf("https://api.appsflyer.com/raw-data/v1.0/conversion/%s?device_id=%s", appID, url.QueryEscape(deviceID))
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		return err("request creation failed: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var data interface{}
	if e = json.Unmarshal(body, &data); e != nil {
		return err("parse failed: " + e.Error())
}

	return success(fmt.Sprintf("conversion data: %v", data))
}