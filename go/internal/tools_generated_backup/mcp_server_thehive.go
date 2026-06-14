package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleListCases(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "server_url")
	if base == "" {
		return err("server_url is required")
}

	resp, e := http.DefaultClient.Get(base + "/api/case")
	if e != nil {
		return err("failed to fetch cases")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return ok(string(body))
}

func HandleGetCase(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "server_url")
	id, _ :=getString(args, "caseId")
	if base == "" || id == "" {
		return err("server_url and caseId are required")
}

	resp, e := http.DefaultClient.Get(base + "/api/case/" + id)
	if e != nil {
		return err("failed to fetch case")
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse response")
}

	out, _ := json.Marshal(result)
	return ok(string(out))
}