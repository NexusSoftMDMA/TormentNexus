package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListPods_kubernetes_mcp_server_linux_amd64(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	namespace, _ :=getString(args, "namespace")
	if namespace == "" {
		namespace = "default"
	}
	url := fmt.Sprintf("https://kubernetes.default.svc/api/v1/namespaces/%s/pods", namespace)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("create request: " + e.Error())
	}
	req.Header.Set("Authorization", "Bearer "+getString(args, "token"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode failed: " + e.Error())
	}
	return ok("pods listed")
}

func HandleGetPodLogs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	namespace, _ :=getString(args, "namespace")
	if namespace == "" {
		namespace = "default"
	}
	pod, _ :=getString(args, "pod")
	if pod == "" {
		return err("pod name required")
	}
	url := fmt.Sprintf("https://kubernetes.default.svc/api/v1/namespaces/%s/pods/%s/log", namespace, pod)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("create request: " + e.Error())
	}
	req.Header.Set("Authorization", "Bearer "+getString(args, "token"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	var logs string
	if e := json.NewDecoder(resp.Body).Decode(&logs); e != nil {
		return err("decode failed: " + e.Error())
	}
	return success(logs)
}

func HandleApplyManifest(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	manifest, _ :=getString(args, "manifest")
	if manifest == "" {
		return err("manifest required")
	}
	url := "https://kubernetes.default.svc/apis/apps/v1/namespaces/default/deployments"
	req, e := http.NewRequestWithContext(ctx, "POST", url, nil)
	if e != nil {
		return err("create request: " + e.Error())
	}
	req.Header.Set("Authorization", "Bearer "+getString(args, "token"))
	req.Header.Set("Content-Type", "application/yaml")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	return ok("manifest applied")
}// touch 1781132130
