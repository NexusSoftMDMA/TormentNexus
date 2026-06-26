package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleCreateRuntime(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	manifest, _ :=getString(args, "manifest")
	if manifest == "" {
		return err("manifest is required")
}

	body, _ := json.Marshal(map[string]string{"manifest": manifest})
	resp, e := http.DefaultClient.Post("http://localhost:8080/runtime", "application/json", bytes.NewBuffer(body))
	if e != nil {
		return err("failed to create runtime: " + e.Error())
}

	defer resp.Body.Close()
	return ok("Runtime created successfully")
}

func HandleExecuteAction(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	runtimeID, _ :=getString(args, "runtime_id")
	action, _ :=getString(args, "action")
	if runtimeID == "" || action == "" {
		return err("runtime_id and action are required")
}

	body, _ := json.Marshal(map[string]string{"action": action})
	req, e := http.NewRequestWithContext(ctx, "POST", "http://localhost:8080/runtime/"+runtimeID+"/action", bytes.NewBuffer(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to execute action: " + e.Error())
}

	defer resp.Body.Close()
	return ok("Action executed successfully")
}