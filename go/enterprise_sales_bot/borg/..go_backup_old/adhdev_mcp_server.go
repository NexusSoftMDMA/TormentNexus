package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleListAgents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filter, _ :=getString(args, "filter")
	url := "http://localhost:8080/agents"
	if filter != "" {
		url += "?filter=" + filter
	}
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to list agents: " + e.Error())
}

	defer resp.Body.Close()
	var agents []map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&agents); e != nil {
		return err("failed to decode agents: " + e.Error())
}

	data, _ := json.Marshal(agents)
	return ok(string(data))
}

func HandleExecuteAgent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	agentName, _ :=getString(args, "agentName")
	task, _ :=getString(args, "task")
	if agentName == "" || task == "" {
		return err("agentName and task are required")
}

	body, _ := json.Marshal(map[string]string{"task": task})
	resp, e := http.DefaultClient.Post("http://localhost:8080/agents/"+agentName+"/execute", "application/json", bytes.NewReader(body))
	if e != nil {
		return err("failed to execute agent: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode result: " + e.Error())
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}