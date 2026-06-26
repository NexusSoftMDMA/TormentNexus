package mcpimpl

import "context"

func HandleInstall(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pkg, _ :=getString(args, "package")
	if pkg == "" {
		return err("package name is required")
}

	version, _ :=getString(args, "version")
	if version != "" {
		return ok("installed " + pkg + " version " + version)
}

	return ok("installed " + pkg)
}