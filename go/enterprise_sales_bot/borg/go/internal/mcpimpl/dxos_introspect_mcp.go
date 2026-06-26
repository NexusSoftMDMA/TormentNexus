package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetPackageInfo_dxos_introspect_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("missing 'name' parameter")
}

	url := fmt.Sprintf("https://registry.npmjs.org/%s", name)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("failed to fetch: %v", e))
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err(fmt.Sprintf("failed to parse: %v", e))
}

	version, found := data["version"].(string)
	if !found {
		return err("version not found")
}

	return ok(fmt.Sprintf("Package %s version: %s", name, version))
}