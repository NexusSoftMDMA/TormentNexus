package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListTopics_confluentinc_mcp_confluent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	clusterID, _ :=getString(args, "cluster_id")
	if clusterID == "" {
		return err("cluster_id is required")
}

	url := fmt.Sprintf("https://api.confluent.cloud/kafka/v3/clusters/%s/topics", clusterID)
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

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	return ok(fmt.Sprintf("topics: %v", data))
}

func HandleDescribeTopic(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	clusterID, _ :=getString(args, "cluster_id")
	topicName, _ :=getString(args, "topic_name")
	if clusterID == "" || topicName == "" {
		return err("cluster_id and topic_name are required")
}

	url := fmt.Sprintf("https://api.confluent.cloud/kafka/v3/clusters/%s/topics/%s", clusterID, topicName)
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

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", e))
}

	return ok(fmt.Sprintf("topic details: %v", data))
}