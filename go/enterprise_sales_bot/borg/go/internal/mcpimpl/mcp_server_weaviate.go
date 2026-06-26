package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleWeaviateQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("missing query argument")
}

	url := fmt.Sprintf("http://localhost:8080/v1/graphql")
	body := fmt.Sprintf(`{"query":"{Get{Meta{Count}}}"}`)
	req, e := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(body))
	if e != nil {
		return err(e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(e.Error())
}

	return success(fmt.Sprintf("Weaviate response: %v", result))
}

func HandleWeaviateHealth(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url := "http://localhost:8080/v1/.well-known/ready"
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return success("Weaviate is ready")
}

	return err(fmt.Sprintf("Weaviate health check failed with status: %d", resp.StatusCode))
}// touch 1781132133
