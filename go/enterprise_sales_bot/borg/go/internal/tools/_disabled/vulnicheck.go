package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleCheckVulnerability(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pkg, _ :=getString(args, "package")
	if pkg == "" {
		return err("package is required")
}

	resp, e := http.DefaultClient.Get("https://vulni-check.example.com/" + pkg)
	if e != nil {
		return err("failed to check: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid response")
}

	return ok("Vulnerability check complete: " + string(body))
}

func HandleScanPackage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pkg, _ :=getString(args, "package")
	if pkg == "" {
		return err("package is required")
}

	return success("Scanning package: " + pkg)
}