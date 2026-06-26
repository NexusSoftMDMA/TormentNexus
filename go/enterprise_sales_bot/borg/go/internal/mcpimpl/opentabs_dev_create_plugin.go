package mcpimpl

import (
	"context"
	"os"
	"path/filepath"
)

func HandleCreatePlugin(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	dir := filepath.Join(".", name)
	e := os.Mkdir(dir, 0755)
	if e != nil {
		return err("failed to create directory: " + e.Error())
}

	manifest := `{"name":"` + name + `","version":"0.1.0"}`
	e = os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(manifest), 0644)
	if e != nil {
		return err("failed to write manifest: " + e.Error())
}

	return ok("plugin " + name + " created successfully")
}