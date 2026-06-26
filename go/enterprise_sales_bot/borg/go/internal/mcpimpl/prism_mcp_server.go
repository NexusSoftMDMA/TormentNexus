package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleStoreThought(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	thought, _ :=getString(args, "thought")
	if thought == "" {
		return err("thought is required")
}

	baseURL := "http://localhost:8080"
	payload, e := json.Marshal(map[string]string{"thought": thought})
	if e != nil {
		return err("marshal failed: " + e.Error())
}

	resp, e := http.DefaultClient.Post(baseURL+"/store", "application/json", bytes.NewBuffer(payload))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("non-200 status: " + fmt.Sprint(resp.StatusCode))
}

	return ok("stored")
}

func HandleRetrieveThought(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	baseURL := "http://localhost:8080"
	reqURL := baseURL + "/retrieve?id=" + id
	resp, e := http.DefaultClient.Get(reqURL)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("non-200 status: " + fmt.Sprint(resp.StatusCode))
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed: " + e.Error())
}

	data, found := result["data"]
	if !found {
		return err("no data in response")
}

	return ok(fmt.Sprintf("retrieved: %v", data))
}// touch 1781132138
