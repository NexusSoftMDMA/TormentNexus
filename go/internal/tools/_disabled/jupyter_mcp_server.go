package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleListKernels(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serverURL, _ :=getString(args, "server_url")
	req, e := http.NewRequestWithContext(ctx, "GET", serverURL+"/api/kernels", nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("unexpected status")
}

	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed")
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}

func HandleExecuteCode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serverURL, _ :=getString(args, "server_url")
	kernelID, _ :=getString(args, "kernel_id")
	code, _ :=getString(args, "code")
	body := map[string]string{"code": code}
	b, _ := json.Marshal(body)
	req, e := http.NewRequestWithContext(ctx, "POST", serverURL+"/api/kernels/"+kernelID+"/execute", bytes.NewReader(b))
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		return err("unexpected status")
}

	return success("code executed")
}