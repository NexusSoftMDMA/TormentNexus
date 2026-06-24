package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base_url")
	query, _ :=getString(args, "query")
	body, e := json.Marshal(map[string]string{"query": query})
	if e != nil {
		return err(fmt.Sprintf("marshal error: %v", e))
}

	resp, e := http.DefaultClient.Post(base+"/query", "application/json", bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("request error: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode error: %v", e))
}

	return ok(fmt.Sprintf("Query result: %v", result))
}

func HandleExecute(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base_url")
	query, _ :=getString(args, "query")
	body, e := json.Marshal(map[string]string{"query": query})
	if e != nil {
		return err(fmt.Sprintf("marshal error: %v", e))
}

	resp, e := http.DefaultClient.Post(base+"/execute", "application/json", bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("request error: %v", e))
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode error: %v", e))
}

	return success(fmt.Sprintf("Execute result: %v", result))
}