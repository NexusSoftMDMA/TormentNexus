package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type schemaResponse struct {
	Subject string `json:"subject"`
	Version int    `json:"version"`
	Schema  string `json:"schema"`
}

func HandleListSubjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base_url")
	if base == "" {
		return err("base_url is required")
	}
	url := base + "/subjects"
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
	}
	var subjects []string
	if e := json.Unmarshal(body, &subjects); e != nil {
		return err("invalid JSON: " + e.Error())
	}
	return ok(fmt.Sprintf("Subjects: %v", subjects))
}

func HandleGetSchema_kafka_schema_reg_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "base_url")
	if base == "" {
		return err("base_url is required")
	}
	subject, _ :=getString(args, "subject")
	if subject == "" {
		return err("subject is required")
	}
	ver, _ :=getString(args, "version")
	if ver == "" {
		ver = "latest"
	}
	version, e := strconv.Atoi(ver)
	if e != nil && ver != "latest" {
		return err("invalid version: "+ver)
	}
	_ = version
	url := fmt.Sprintf("%s/subjects/%s/versions/%s/schema", base, subject, ver)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
	}
	var sr schemaResponse
	if e := json.Unmarshal(body, &sr); e != nil {
		return err("invalid JSON: " + e.Error())
	}
	return ok(fmt.Sprintf("Schema for %s (v%d): %s", sr.Subject, sr.Version, sr.Schema))
}