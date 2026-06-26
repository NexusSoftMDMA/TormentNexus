package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func HandleKubectlRun(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	if command == "" {
		return err("command is required")
}

	parts := strings.Fields(command)
	cmd := exec.CommandContext(ctx, "kubectl", parts...)
	out, e := cmd.CombinedOutput()
	if e != nil {
		return err(fmt.Sprintf("kubectl error: %v", e))
}

	return ok(string(out))
}

func HandleListPods(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	namespace, _ :=getString(args, "namespace")
	if namespace == "" {
		namespace = "default"
	}
	cmd := exec.CommandContext(ctx, "kubectl", "get", "pods", "-n", namespace, "-o", "wide")
	out, e := cmd.CombinedOutput()
	if e != nil {
		return err(fmt.Sprintf("kubectl error: %v", e))
}

	return ok(string(out))
}