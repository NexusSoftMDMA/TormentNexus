package tools

import (
	"context"
	"os/exec"
)

func HandleKubectlGetPods(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	namespace, _ :=getString(args, "namespace")
	var cmd *exec.Cmd
	if namespace == "" {
		cmd = exec.CommandContext(ctx, "kubectl", "get", "pods", "-o", "wide")
	} else {
		cmd = exec.CommandContext(ctx, "kubectl", "get", "pods", "-n", namespace, "-o", "wide")

	out, e := cmd.Output()
	if e != nil {
		return err("Failed to get pods: " + e.Error())
}

	return ok(string(out))
}

}

func HandleKubectlGetDeployments(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	namespace, _ :=getString(args, "namespace")
	var cmd *exec.Cmd
	if namespace == "" {
		cmd = exec.CommandContext(ctx, "kubectl", "get", "deployments", "-o", "wide")
	} else {
		cmd = exec.CommandContext(ctx, "kubectl", "get", "deployments", "-n", namespace, "-o", "wide")

	out, e := cmd.Output()
	if e != nil {
		return err("Failed to get deployments: " + e.Error())
}

	return ok(string(out))
}
}