package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleSimctlList(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	deviceType, _ :=getString(args, "device_type")
	if deviceType == "" {
		deviceType = "iPhone"
	}
	url := fmt.Sprintf("https://simctl-api.example.com/list?type=%s", deviceType)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(e.Error())
	}
	return success(fmt.Sprintf("Found devices for %s", deviceType))
}

func HandleSimctlBoot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	udid, _ :=getString(args, "udid")
	if udid == "" {
		return err("udid is required")
	}
	url := "https://simctl-api.example.com/boot"
	payload := strings.NewReader(fmt.Sprintf(`{"udid":"%s"}`, udid))
	req, e := http.NewRequestWithContext(ctx, "POST", url, payload)
	if e != nil {
		return err(e.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("boot failed with status %d", resp.StatusCode))
	}
	return success(fmt.Sprintf("Device %s booted successfully", udid))
}// touch 1781132140
