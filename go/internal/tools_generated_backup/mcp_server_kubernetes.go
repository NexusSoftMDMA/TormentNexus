package tools

import "context"

func HandleListPods(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	namespace, _ :=getString(args, "namespace")
	if namespace == "" {
		namespace = "default"
	}
	return ok("Listed pods in namespace: " + namespace)
}

func HandleGetPodStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pod, _ :=getString(args, "pod")
	if pod == "" {
		return err("pod name is required")
}

	namespace, _ :=getString(args, "namespace")
	if namespace == "" {
		namespace = "default"
	}
	return ok("Pod status for " + pod + " in namespace " + namespace + ": Running")
}