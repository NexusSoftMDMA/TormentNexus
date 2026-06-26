package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGlobalpingPing(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	if host == "" {
		return err("host is required")
}

	count, _ :=getInt(args, "count")
	if count <= 0 {
		count = 3
	}
	url := fmt.Sprintf("https://api.globalping.io/v1/ping?host=%s&count=%d", host, count)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse failed: " + e.Error())
}

	return ok(fmt.Sprintf("Ping result: %v", result))
}

func HandleGlobalpingTraceroute(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	if host == "" {
		return err("host is required")
}

	protocol, _ :=getString(args, "protocol")
	if protocol == "" {
		protocol = "icmp"
	}
	url := fmt.Sprintf("https://api.globalping.io/v1/traceroute?host=%s&protocol=%s", host, protocol)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse failed: " + e.Error())
}

	return ok(fmt.Sprintf("Traceroute result: %v", result))
}