package tools

import (
	"bytes"
	"context"
	"os/exec"
)

func HandleGetKubeConfigContexts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	cmd := exec.CommandContext(ctx, "kubectl", "config", "get-contexts", "-o", "name")
	var out bytes.Buffer
	cmd.Stdout = &out
	e := cmd.Run()
	if e != nil {
		return err("failed to get kube contexts: " + e.Error())
}

	return ok("Contexts:\n" + out.String())
}