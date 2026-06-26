package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleScanDependency(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	version, _ :=getString(args, "version")
	if name == "" {
		return err("name is required")
}

	u := fmt.Sprintf("https://cryptoguard.dev/api/dependency?name=%s&version=%s", url.QueryEscape(name), url.QueryEscape(version))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode error")
}

	return ok(fmt.Sprintf("Dependency analysis: %v", result))
}

func HandleScanFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path is required")
}

	u := fmt.Sprintf("https://cryptoguard.dev/api/file?path=%s", url.QueryEscape(path))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode error")
}

	return ok(fmt.Sprintf("File scan result: %v", result))
}