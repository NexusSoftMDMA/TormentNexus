package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleListCameras(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "unifi_host")
	key, _ :=getString(args, "api_key")
	url := fmt.Sprintf("https://%s/proxy/protect/api/cameras", host)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error()), e
	}
	req.Header.Set("Authorization", "Bearer "+key)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error()), e
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body failed: " + e.Error()), e
	}
	return ok(string(body))
}

func HandleGetCamera(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "unifi_host")
	key, _ :=getString(args, "api_key")
	id, _ :=getString(args, "camera_id")
	if id == "" {
		return err("camera_id is required"), fmt.Errorf("missing camera_id")
}

	url := fmt.Sprintf("https://%s/proxy/protect/api/cameras/%s", host, id)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error()), e
	}
	req.Header.Set("Authorization", "Bearer "+key)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error()), e
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body failed: " + e.Error()), e
	}
	return ok(string(body))
}