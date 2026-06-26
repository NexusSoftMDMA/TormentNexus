package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleExecuteQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sql, _ :=getString(args, "sql")
	if sql == "" {
		return err("sql is required")
}

	url, _ :=getString(args, "url")
	if url == "" {
		url = "http://localhost:8080"
	}
	user, _ :=getString(args, "user")
	password, _ :=getString(args, "password")

	body, e := json.Marshal(map[string]string{"query": sql})
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", url+"/v1/statement", bytes.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	if user != "" || password != "" {
		req.SetBasicAuth(user, password)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()

	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("trino returned status %d: %s", resp.StatusCode, string(respBody)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(respBody, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	if data, found := result["data"]; found {
		return ok(fmt.Sprintf("%v", data))
}

	return ok(string(respBody))
}
}