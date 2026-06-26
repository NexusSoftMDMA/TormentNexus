package mcpimpl

import (
	"context"
	"net/http"
	"io"
)

func HandleTestEpa(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		name = "EPA"
	}
	resp, e := http.DefaultClient.Get("https://api.epa.gov/test?name=" + name)
	if e != nil {
		return err("failed to request: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read: " + e.Error())
}

	return ok(string(body))
}