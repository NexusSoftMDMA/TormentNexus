package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListClusters(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	token, _ :=getString(args, "token")
	if host == "" || token == "" {
		return err("host and token are required")
}

	url := fmt.Sprintf("%s/api/2.0/clusters/list", host)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
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
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	clusters, found := result["clusters"].([]interface{})
	if !found {
		return err("invalid response format")
}

	return ok(fmt.Sprintf("Found %d clusters", len(clusters)))
}

func HandleGetCluster(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	token, _ :=getString(args, "token")
	clusterID, _ :=getString(args, "cluster_id")
	if host == "" || token == "" || clusterID == "" {
		return err("host, token, and cluster_id are required")
}

	url := fmt.Sprintf("%s/api/2.0/clusters/get?cluster_id=%s", host, clusterID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
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

	return ok(string(body))
}