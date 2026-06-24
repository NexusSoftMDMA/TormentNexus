package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleListVms(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	url := "https://api.example.com/vms"
	if name != "" {
		url += "?name=" + name
	}
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch VMs: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return success(string(body))
}