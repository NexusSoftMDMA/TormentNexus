package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleSystemHealth(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok("Health: " + string(body))
}

func HandleSystemInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	if host == "" {
		host = "localhost"
	}
	resp, e := http.DefaultClient.Get("http://" + host + ":8080/info")
	if e != nil {
		return err("system info unavailable: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read info failed: " + e.Error())
}

	return success("System Info: " + string(body))
}