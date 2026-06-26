package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListComponents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://registry.shadcn.com/api/components")
	if e != nil {
		return err("failed to fetch components: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("registry returned status " + resp.Status)
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var components []string
	if e := json.Unmarshal(body, &components); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	out, _ := json.Marshal(components)
	return ok("Components: " + string(out))
}

func HandleGetComponent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("missing component name")
}

	url := fmt.Sprintf("https://registry.shadcn.com/api/components/%s", name)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch component: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("component not found or registry error: " + resp.Status)
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok("Component: " + string(body))
}