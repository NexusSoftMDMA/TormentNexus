package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleGenerateCRUDPage(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	entityName, _ :=getString(args, "entityName")
	fields, _ :=getString(args, "fields")
	if entityName == "" {
		return err("entityName is required")
}

	body := fmt.Sprintf(`{"entityName":"%s","fields":"%s"}`, entityName, fields)
	req, e := http.NewRequestWithContext(ctx, "POST", "http://localhost:3000/api/crud/generate", strings.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("http request failed: " + e.Error())
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read response failed: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("server returned " + resp.Status)
}

	var result map[string]interface{}
	json.Unmarshal(respBody, &result)
	output, _ := json.Marshal(result)
	return ok("CRUD page generated: " + string(output))
}

func HandleGetEntitySchema(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	entity, _ :=getString(args, "entity")
	if entity == "" {
		return err("entity is required")
}

	url := "http://localhost:3000/api/schema/" + entity
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("http request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read response failed: " + e.Error())
}

	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("invalid JSON: " + e.Error())
}

	out, _ := json.Marshal(data)
	return ok("Schema: " + string(out))
}