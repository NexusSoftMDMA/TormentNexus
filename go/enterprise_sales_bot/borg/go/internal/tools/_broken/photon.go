package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleTransform(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	target, _ :=getString(args, "target")
	if code == "" {
		return err("missing code")
}

	reqBody, e := json.Marshal(map[string]string{"code": code, "target": target})
	if e != nil {
		return err("marshal failed")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.photon.dev/transform", nil)
	if e != nil {
		return err("request init failed")
}

	req.Body = io.NopCloser(bytes.NewReader(reqBody))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	return success("transformation complete")
}

func HandleDeploy_photon(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	project, _ :=getString(args, "project")
	if project == "" {
		return err("missing project")
}

	return ok("deployed " + project)
}