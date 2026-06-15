package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func HandleCreateKnowledge(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	base, _ :=getString(args, "base_url")
	if base == "" {
		base = os.Getenv("CODEX_API_BASE")

	reqBody, e := json.Marshal(map[string]string{"key": key, "value": value})
	if e != nil {
		return err(fmt.Sprintf("marshal failed: %v", e))
}

	req, e := http.NewRequestWithContext(ctx, "POST", strings.TrimRight(base, "/")+"/knowledge", strings.NewReader(string(reqBody)))
	if e != nil {
		return err(fmt.Sprintf("request creation failed: %v", e))
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return err(fmt.Sprintf("unexpected status: %s", resp.Status))
}

	return ok("knowledge created")
}

}

func HandleQueryKnowledge(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	base, _ :=getString(args, "base_url")
	if base == "" {
		base = os.Getenv("CODEX_API_BASE")

	url := fmt.Sprintf("%s/knowledge/%s", strings.TrimRight(base, "/"), key)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("request creation failed: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status: %s", resp.Status))
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	return success(fmt.Sprintf("%v", result))
}
}