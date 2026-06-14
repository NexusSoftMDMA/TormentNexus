package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListDevices(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base_url")
	url := fmt.Sprintf("%s/devices", base)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("unexpected status: " + resp.Status)
}

	b, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(b, &result); e != nil {
		return err("parse failed: " + e.Error())
}

	return success(fmt.Sprintf("devices: %+v", result))
}

func HandleGetDeviceInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base_url")
	id, _ :=getString(args, "device_id")
	url := fmt.Sprintf("%s/devices/%s", base, id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("unexpected status: " + resp.Status)
}

	b, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(b, &result); e != nil {
		return err("parse failed: " + e.Error())
}

	return success(fmt.Sprintf("device info: %+v", result))
}