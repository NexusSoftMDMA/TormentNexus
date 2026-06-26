package mcpimpl

import (
	"context"
	"encoding/json"
	"net/http"
)

func HandleListTopics_ros_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "http://localhost:9090/topics"
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to get topics: " + e.Error())
}

	defer resp.Body.Close()
	var topics []string
	if e := json.NewDecoder(resp.Body).Decode(&topics); e != nil {
		return err("failed to decode topics: " + e.Error())
}

	return ok("topics: " + joinStrings(topics))
}

func HandleTopicInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	topic, _ :=getString(args, "topic")
	url := "http://localhost:9090/topic_info?name=" + topic
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to get topic info: " + e.Error())
}

	defer resp.Body.Close()
	var info map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&info); e != nil {
		return err("failed to decode info: " + e.Error())
}

	return success("topic info: " + prettyJSON(info))
}

func joinStrings_ros_mcp_server(s []string) string {
	result := ""
	for i, v := range s {
		if i > 0 {
			result += ", "
		}
		result += v
	}
	return result
}

func prettyJSON(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}