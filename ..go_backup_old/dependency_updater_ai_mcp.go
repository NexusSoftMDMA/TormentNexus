package tools

import "context"

func HandleCheckOutdated(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pkg, _ :=getString(args, "package")
	if pkg == "" {
		return err("missing package")
}

	return ok("checked outdated for " + pkg + ": none")
}

func HandleSuggestUpdates(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pkg, _ :=getString(args, "package")
	if pkg == "" {
		return err("missing package")
}

	return ok("suggested updates for " + pkg + ": upgrade to latest")
}