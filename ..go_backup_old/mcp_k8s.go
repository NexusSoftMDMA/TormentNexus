package tools

import "context"

func HandleK8sInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	namespace, _ :=getString(args, "namespace")
	return ok("K8s namespace: " + namespace)
}

func HandleK8sCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	return ok("Executing command: " + command)
}