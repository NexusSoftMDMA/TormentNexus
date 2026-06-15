package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetServerInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	if host == "" {
		return err("host is required")
}

	port, _ :=getString(args, "port")
	if port == "" {
		port = "9990"
	}
	url := fmt.Sprintf("http://%s:%s/management/server", host, port)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to connect: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	return ok(string(body))
}

func HandleListDeployments(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	if host == "" {
		return err("host is required")
}

	port, _ :=getString(args, "port")
	if port == "" {
		port = "9990"
	}
	url := fmt.Sprintf("http://%s:%s/management/deployments", host, port)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to connect: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	deployments, found := result["deployment"].([]interface{})
	if !found {
		return err("no deployments found")
}

	return ok(fmt.Sprintf("%v", deployments))
}