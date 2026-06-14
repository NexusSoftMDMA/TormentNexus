package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleFetchOpenApiSpec(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("missing 'url' argument")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("HTTP request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read error: %v", e))
}

	return ok(string(body))
}

func HandleListPaths(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("missing 'url' argument")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("HTTP request failed: %v", e))
}

	defer resp.Body.Close()
	var spec map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&spec); e != nil {
		return err(fmt.Sprintf("JSON decode error: %v", e))
}

	paths, found := spec["paths"].(map[string]interface{})
	if !found {
		return err("no paths in spec")
}

	var list string
	for p := range paths {
		list += p + "\n"
	}
	return ok(list)
}// touch 1781132128
