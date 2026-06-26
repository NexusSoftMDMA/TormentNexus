package tools

import (
	"context"
	"fmt"
	"os/exec"
)

func HandlePhpSdkVersion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.Command("php", "-r", "echo PHP_VERSION;")
	out, e := cmd.Output()
	if e != nil {
		return err("failed to get PHP version: " + e.Error())
}

	return ok(string(out))
}

func HandlePhpSdkEval(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("code is required")
}

	cmd := exec.Command("php", "-r", code)
	out, e := cmd.Output()
	if e != nil {
		return err("execution error: " + e.Error())
}

	return success(string(out))
}