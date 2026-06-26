package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleRunTask(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	task, _ :=getString(args, "task")
	input, _ :=getString(args, "input")
	body, _ := json.Marshal(map[string]string{"task": task, "input": input})
	resp, e := http.DefaultClient.Post("http://localhost:8080/run", "application/json", bytes.NewReader(body))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}

func HandleGetResult_valkey_ai_tasks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	resp, e := http.DefaultClient.Get("http://localhost:8080/result?id=" + id)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	data, _ := json.Marshal(result)
	return ok(string(data))
}