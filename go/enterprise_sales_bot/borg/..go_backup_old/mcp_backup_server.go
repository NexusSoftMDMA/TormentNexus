package tools

import "context"

func HandleCreateBackup(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	source, _ :=getString(args, "source")
	if source == "" {
		return err("source is required")
}

	return ok("backup " + name + " created from " + source)
}

func HandleListBackups(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("backups: example1, example2")
}