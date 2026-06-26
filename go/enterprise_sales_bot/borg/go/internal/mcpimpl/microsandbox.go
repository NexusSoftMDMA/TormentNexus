package mcpimpl

import (
	"context"
	"net/http"
)

func HandleMicrosandboxRun(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	action, _ :=getString(args, "action")
	if action == "" {
		return err("missing action")
}

	if action != "run" {
		return err("unsupported action")
}

	resp, e := http.DefaultClient.Get("http://example.com")
	if e != nil {
		return err(e.Error())
}

	resp.Body.Close()
	return success("microsandbox run completed")
}

func HandleMicrosandboxStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return success("microsandbox is ready")
}// touch 1781132134
