package mcpimpl

import (
	"context"
	"os"
)

func HandleReadSpec(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path is required")
}

	content, e := os.ReadFile(path)
	if e != nil {
		return err("failed to read spec: " + e.Error())
}

	return ok(string(content))
}

func HandleGenerateCode_mcp_server_spec_driven_development(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	spec, _ :=getString(args, "spec")
	if spec == "" {
		return err("spec is required")
}

	code := "// Generated from spec\npackage main\nfunc main() {\n\tprintln(\"Hello, Spec-Driven Development!\")\n}"
	return success(code)
}