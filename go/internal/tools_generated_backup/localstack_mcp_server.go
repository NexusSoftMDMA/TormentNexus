package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func HandleListServices(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "endpoint")
	if url == "" {
		url = "http://localhost:4566"
	}
	resp, e := http.DefaultClient.Get(url + "/_localstack/health")
	if e != nil {
		return err("failed to connect: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	services, found := result["services"].(map[string]interface{})
	if !found {
		return err("no services in response")
}

	var list []string
	for name := range services {
		list = append(list, name)

	return ok("Services: " + strings.Join(list, ", "))
}
}