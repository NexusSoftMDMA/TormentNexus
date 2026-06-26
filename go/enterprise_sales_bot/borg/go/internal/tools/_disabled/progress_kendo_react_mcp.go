package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleGetKendoComponents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "https://kendoreact.com/api/components.json"
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}

func HandleGetKendoComponent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	component, _ :=getString(args, "component")
	if component == "" {
		return err("missing required argument 'component'")
}

	url := fmt.Sprintf("https://kendoreact.com/api/components/%s.json", component)
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}