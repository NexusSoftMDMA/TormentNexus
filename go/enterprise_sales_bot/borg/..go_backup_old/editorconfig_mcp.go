package tools

import (
	"context"
	"encoding/json"
	"os"
	"strings"
)

func HandleGetEditorconfigProperties(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	file, _ :=getString(args, "file")
	if file == "" {
		return err("missing file argument")
	}
	data, e := os.ReadFile(".editorconfig")
	if e != nil {
		return err("cannot read .editorconfig: " + e.Error())
	}
	props := map[string]string{}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			props[key] = val
		}
	}
	jsonBytes, e := json.Marshal(props)
	if e != nil {
		return err("json marshal error: " + e.Error())
	}
	return ok(string(jsonBytes))
}