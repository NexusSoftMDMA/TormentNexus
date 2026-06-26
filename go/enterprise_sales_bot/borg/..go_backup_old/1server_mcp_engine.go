package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleListServers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "https://api.1server.io/servers"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to list servers: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("server returned status " + http.StatusText(resp.StatusCode))
}

	var result []map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid JSON: " + e.Error())
}

	return ok("servers: " + string(body))
}

func HandleInstallServer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("server name is required")
}

	url := "https://api.1server.io/install"
	resp, e := http.DefaultClient.Post(url, "application/json", nil)
	if e != nil {
		return err("failed to install server: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("installation failed: " + string(body))
}

	return success("installed server " + name + ": " + string(body))
}