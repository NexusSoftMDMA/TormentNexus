package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListNodes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "api_base")
	if base == "" {
		return err("api_base is required")
	}
	url := fmt.Sprintf("%s/api2/json/nodes", base)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
	}
	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
	}
	data, found := result["data"]
	if !found {
		return err("no data in response")
	}
	return ok(fmt.Sprintf("Nodes: %v", data))
}

func HandleGetVMStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "api_base")
	if base == "" {
		return err("api_base is required")
	}
	node, _ :=getString(args, "node")
	if node == "" {
		return err("node is required")
	}
	vmid, _ :=getString(args, "vmid")
	if vmid == "" {
		return err("vmid is required")
	}
	url := fmt.Sprintf("%s/api2/json/nodes/%s/qemu/%s/status/current", base, node, vmid)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
	}
	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
	}
	data, found := result["data"]
	if !found {
		return err("no data in response")
	}
	return ok(fmt.Sprintf("VM status: %v", data))
}