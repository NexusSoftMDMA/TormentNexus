package mcpimpl

import (
	"context"
	"io"
	"net/http"
)

func HandleQuarkusHealth(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		url = "http://localhost:8080/q/health"
	}
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to call Quarkus health: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("health returned status " + http.StatusText(resp.StatusCode))
}

	return ok("Quarkus health: " + string(body))
}

func HandleQuarkusInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		url = "http://localhost:8080/q/info"
	}
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to call Quarkus info: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("info returned status " + http.StatusText(resp.StatusCode))
}

	return ok("Quarkus info: " + string(body))
}