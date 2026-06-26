package mcpimpl

import (
	"context"
	"io"
	"net/http"
)

func HandleListTemplates_dotnet_template_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.dotnet-templates.com/templates")
	if e != nil {
		return err("failed to fetch templates: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("unexpected status: " + resp.Status)
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body: " + e.Error())
}

	return ok(string(body))
}

func HandleGetTemplate_dotnet_template_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
	}
	url := "https://api.dotnet-templates.com/templates?name=" + name
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch template: " + e.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("unexpected status: " + resp.Status)
	}
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body: " + e.Error())
	}
	return ok(string(body))
}