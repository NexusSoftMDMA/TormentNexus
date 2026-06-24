package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleSearchPineScriptDocs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	url := fmt.Sprintf("https://pine-script-docs.example.com/search?q=%s", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to search docs: " + e.Error())
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return ok("Search results for '" + query + "': " + string(body))
}

func HandleGetPineScriptReference(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name parameter is required")
}

	url := fmt.Sprintf("https://pine-script-docs.example.com/reference/%s", name)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to get reference: " + e.Error())
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return ok("Reference for '" + name + "': " + string(body))
}