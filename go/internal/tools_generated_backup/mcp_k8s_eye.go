package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleGetPods(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	namespace, _ :=getString(args, "namespace")
	if namespace == "" {
		namespace = "default"
	}
	host := os.Getenv("K8S_HOST")
	token := os.Getenv("K8S_TOKEN")
	if host == "" || token == "" {
		return err("missing K8S_HOST or K8S_TOKEN")
}

	req, e := http.NewRequestWithContext(ctx, "GET", host+"/api/v1/namespaces/"+namespace+"/pods", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	json.Unmarshal(body, &result)
	return ok(fmt.Sprintf("pods: %v", result))
}

func HandleGetClusterInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host := os.Getenv("K8S_HOST")
	token := os.Getenv("K8S_TOKEN")
	if host == "" || token == "" {
		return err("missing K8S_HOST or K8S_TOKEN")
}

	req, e := http.NewRequestWithContext(ctx, "GET", host+"/version", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	json.Unmarshal(body, &result)
	return ok(fmt.Sprintf("cluster info: %v", result))
}