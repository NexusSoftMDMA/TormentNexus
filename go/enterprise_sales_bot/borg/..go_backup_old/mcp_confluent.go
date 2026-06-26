package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListTopics(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	clusterID, _ :=getString(args, "cluster_id")
	apiKey, _ :=getString(args, "api_key")
	apiSecret, _ :=getString(args, "api_secret")
	if clusterID == "" || apiKey == "" || apiSecret == "" {
		return err("missing cluster_id, api_key, or api_secret")
}

	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.confluent.cloud/kafka/v3/clusters/%s/topics", clusterID), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth(apiKey, apiSecret)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	data, _ := json.Marshal(result["data"])
	return ok(string(data))
}

func HandleGetCluster(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	apiSecret, _ :=getString(args, "api_secret")
	if apiKey == "" || apiSecret == "" {
		return err("missing api_key or api_secret")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.confluent.cloud/kafka/v3/clusters", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth(apiKey, apiSecret)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	data, _ := json.Marshal(result["data"])
	return ok(string(data))
}