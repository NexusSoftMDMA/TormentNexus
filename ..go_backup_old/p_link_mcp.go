package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetDeviceInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	if host == "" {
		return err("host parameter required")
}

	resp, e := http.DefaultClient.Get(fmt.Sprintf("http://%s/info", host))
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse failed: %v", e))
}

	output, _ := json.Marshal(result)
	return ok(string(output))
}

func HandleReboot(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	if host == "" {
		return err("host parameter required")
}

	resp, e := http.DefaultClient.Post(fmt.Sprintf("http://%s/reboot", host), "application/json", nil)
	if e != nil {
		return err(fmt.Sprintf("reboot failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("reboot returned %d", resp.StatusCode))
}

	return success("device rebooting")
}