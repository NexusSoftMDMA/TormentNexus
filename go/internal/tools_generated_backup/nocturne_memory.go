package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleSetMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	if key == "" || value == "" {
		return err("key and value are required")
}

	body, _ := json.Marshal(map[string]string{"key": key, "value": value})
	resp, e := http.DefaultClient.Post("http://localhost:8080/memory", "application/json", bytes.NewReader(body))
	if e != nil {
		return err("failed to set memory: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("memory server returned status " + fmt.Sprint(resp.StatusCode))
}

	return success("memory stored")
}

func HandleGetMemory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	if key == "" {
		return err("key is required")
}

	resp, e := http.DefaultClient.Get("http://localhost:8080/memory?key=" + key)
	if e != nil {
		return err("failed to get memory: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		return err("memory not found")
}

	if resp.StatusCode != 200 {
		return err("memory server returned status " + fmt.Sprint(resp.StatusCode))
}

	data, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	var result map[string]string
	if e := json.Unmarshal(data, &result); e != nil {
		return err("failed to parse response")
}

	return ok(result["value"])
}