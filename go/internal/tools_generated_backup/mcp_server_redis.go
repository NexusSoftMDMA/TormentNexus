package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func HandleGet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	if key == "" {
		return err("key is required")
}

	out, e := exec.Command("redis-cli", "GET", key).Output()
	if e != nil {
		return err(fmt.Sprintf("redis error: %v", e))
}

	value := strings.TrimSpace(string(out))
	if value == "" {
		return ok(fmt.Sprintf("key '%s' not found", key))
}

	return success(map[string]string{"value": value})
}

func HandleSet(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	if key == "" || value == "" {
		return err("key and value are required")
}

	e := exec.Command("redis-cli", "SET", key, value).Run()
	if e != nil {
		return err(fmt.Sprintf("redis error: %v", e))
}

	return ok(fmt.Sprintf("set key '%s'", key))
}