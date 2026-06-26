package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"io"
)

func HandleListTopics_kafka_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		baseURL = "http://localhost:8082" // default Kafka REST proxy
	}
	resp, e := http.DefaultClient.Get(fmt.Sprintf("%s/topics", baseURL))
	if e != nil {
		return err("failed to list topics: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var topics []string
	if e := json.Unmarshal(body, &topics); e != nil {
		return err("failed to parse topics: " + e.Error())
}

	return ok(fmt.Sprintf("Topics: %v", topics))
}

func HandleProduceMessage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		baseURL = "http://localhost:8082"
	}
	topic, _ :=getString(args, "topic")
	if topic == "" {
		return err("topic is required")
}

	value, _ :=getString(args, "value")
	if value == "" {
		return err("value is required")
}

	payload := map[string]interface{}{
		"records": []map[string]string{{"value": value}},
	}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal payload: " + e.Error())
}

	url := fmt.Sprintf("%s/topics/%s", baseURL, topic)
	resp, e := http.DefaultClient.Post(url, "application/vnd.kafka.json.v2+json", nil)
	if e != nil {
		return err("failed to produce message: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return err("kafka responded with status " + resp.Status)
}

	return success("Message produced to topic " + topic)
}