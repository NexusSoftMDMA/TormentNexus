package tools

import (
	"context"
	"fmt"
)

func HandleGenerateMcpServer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	code := fmt.Sprintf("package main\n\nfunc main() {\n\tprintln(\"Hello from %s\")\n}", name)
	return ok(code)
}