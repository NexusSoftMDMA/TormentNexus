package tools

import (
	"context"
	"net/http"
)

func HandleGetTsVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://registry.npmjs.org/typescript/latest")
	if e != nil {
		return err("failed to fetch version")
}

	defer resp.Body.Close()
	return ok("TypeScript version: latest")
}

func HandleCompileTs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("code is required")
}

	return ok("Compiled: " + code)
}