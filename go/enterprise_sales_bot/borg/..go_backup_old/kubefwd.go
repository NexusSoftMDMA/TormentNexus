package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

type serviceList struct {
	Items []struct {
		Metadata struct {
			Name string `json:"name"`
		} `json:"metadata"`
	} `json:"items"`
}

func HandleGetServices(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	namespace, _ :=getString(args, "namespace")
	if namespace == "" {
		namespace = "default"
	}
	cmd := exec.CommandContext(ctx, "kubectl", "get", "svc", "-n", namespace, "-o", "json")
	out, e := cmd.Output()
	if e != nil {
		return err("failed to list services: " + e.Error())
}

	var list serviceList
	if e := json.Unmarshal(out, &list); e != nil {
		return err("failed to parse services: " + e.Error())
}

	names := []string{}
	for _, svc := range list.Items {
		names = append(names, svc.Metadata.Name)

	return success(fmt.Sprintf("Services in %s: %v", namespace, names))
}

}

func HandlePortForward(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	service, _ :=getString(args, "service")
	localPort, _ :=getString(args, "localPort")
	remotePort, _ :=getString(args, "remotePort")
	namespace, _ :=getString(args, "namespace")
	if namespace == "" {
		namespace = "default"
	}
	if service == "" || localPort == "" || remotePort == "" {
		return err("service, localPort, and remotePort are required")
}

	cmd := exec.CommandContext(ctx, "kubectl", "port-forward", "service/"+service, localPort+":"+remotePort, "-n", namespace)
	if e := cmd.Start(); e != nil {
		return err("failed to start port-forward: " + e.Error())
}

	return ok("port-forward started for service " + service + " on " + localPort + ":" + remotePort)
}