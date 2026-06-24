package tools

import (
	"context"
	"encoding/json"
)

func HandleGetDependencies(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pkg, _ :=getString(args, "package")
	if pkg == "" {
		return err("package is required")
}

	deps := map[string]string{pkg: "fmt, os, strings"}
	data, e := json.Marshal(deps)
	if e != nil {
		return err("marshal error")
}

	return ok(string(data))
}

func HandleListPackages(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pkgs := []string{"depwire", "wire", "inject"}
	data, e := json.Marshal(pkgs)
	if e != nil {
		return err("marshal error")
}

	return ok(string(data))
}