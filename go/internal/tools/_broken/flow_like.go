package tools

import (
	"context"
	"fmt"
	"os/exec"
)

// ToolResponse, ok, e, getString, getInt, getBool は parity.go で定義されていると仮定

func HandleDevDesktop(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.Command("bun", "run", "--cwd", "./apps/desktop", "dev:auto")
	output, e := cmd.CombinedOutput()
	if e != nil {
		return err(e)
}

	return ok(string(output))
}

func HandleDevWeb(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.Command("bun", "run", "--cwd", "./apps/web", "dev")
	output, e := cmd.CombinedOutput()
	if e != nil {
		return err(e)
}

	return ok(string(output))
}

func HandleDevApi(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.Command("bun", "run", "--cwd", "./apps/backend/local/api", "dev")
	output, e := cmd.CombinedOutput()
	if e != nil {
		return err(e)
}

	return ok(string(output))
}

func HandleDevRuntime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.Command("bun", "run", "--cwd", "./apps/backend/local/runtime", "dev")
	output, e := cmd.CombinedOutput()
	if e != nil {
		return err(e)
}

	return ok(string(output))
}

func HandleDevEmbedded(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.Command("bun", "run", "--cwd", "./apps/embedded", "dev")
	output, e := cmd.CombinedOutput()
	if e != nil {
		return err(e)
}

	return ok(string(output))
}