//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleGetComponentDocs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "component")
	if name == "" {
		return err("missing required field: component")
}

	url := fmt.Sprintf("https://ark-ui.com/api/docs/%s", strings.ToLower(name))
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch docs: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}

func HandleListComponents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://ark-ui.com/api/components")
	if e != nil {
		return err("failed to list components: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var components []string
	if e := json.Unmarshal(body, &components); e != nil {
		return err("failed to parse response: " + e.Error())
}

	return success(components)
}
