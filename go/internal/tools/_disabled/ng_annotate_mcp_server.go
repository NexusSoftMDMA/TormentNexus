package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleAnnotate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	source, _ :=getString(args, "source")
	if source == "" {
		return err("missing 'source' argument")
}

	body, _ := json.Marshal(map[string]string{"source": source, "action": "annotate"})
	resp, e := http.DefaultClient.Post("http://localhost:8080/annotate", "application/json", bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result struct{ Result string }
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	return ok(result.Result)
}

func HandleRemove(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	source, _ :=getString(args, "source")
	if source == "" {
		return err("missing 'source' argument")
}

	body, _ := json.Marshal(map[string]string{"source": source, "action": "remove"})
	resp, e := http.DefaultClient.Post("http://localhost:8080/remove", "application/json", bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result struct{ Result string }
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	return ok(result.Result)
}