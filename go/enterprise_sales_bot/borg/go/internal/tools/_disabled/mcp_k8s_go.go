package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func HandleListPods(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	namespace, _ :=getString(args, "namespace")
	if namespace == "" {
		namespace = "default"
	}
	token, e := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if e != nil {
		return err("failed to read service account token: " + e.Error())
}

	url := fmt.Sprintf("https://kubernetes.default.svc/api/v1/namespaces/%s/pods", namespace)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(string(token)))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err("API error: " + resp.Status + " - " + string(body))
}

	var result struct {
		Items []struct {
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
		} `json:"items"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	pods := make([]string, 0, len(result.Items))
	for _, item := range result.Items {
		pods = append(pods, item.Metadata.Name)

	return success(fmt.Sprintf("Pods in %s: %s", namespace, strings.Join(pods, ", ")))
}

}

func HandleListNodes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, e := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if e != nil {
		return err("failed to read service account token: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://kubernetes.default.svc/api/v1/nodes", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(string(token)))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err("API error: " + resp.Status + " - " + string(body))
}

	var result struct {
		Items []struct {
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
		} `json:"items"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	nodes := make([]string, 0, len(result.Items))
	for _, item := range result.Items {
		nodes = append(nodes, item.Metadata.Name)

	return success(fmt.Sprintf("Nodes: %s", strings.Join(nodes, ", ")))
}
}