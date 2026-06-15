package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

func HandleListEndpoints(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	dir := "endpoints"
	entries, e := os.ReadDir(dir)
	if e != nil {
		return err("failed to read endpoints directory: " + e.Error())
}

	var names []string
	for _, entry := range entries {
		if !entry.IsDir() {
			names = append(names, strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())))

	}
	data, e := json.Marshal(names)
	if e != nil {
		return err("failed to marshal names")
}

	return ok(string(data))
}

}

func HandleGetEndpoint(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("missing 'name' argument")
}

	filePath := filepath.Join("endpoints", name+".json")
	data, e := os.ReadFile(filePath)
	if e != nil {
		return err("failed to read endpoint documentation: " + e.Error())
}

	return ok(string(data))
}