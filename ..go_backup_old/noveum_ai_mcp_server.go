package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleListEndpoints(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch spec: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body: " + e.Error())
}

	var spec map[string]interface{}
	if e := json.Unmarshal(body, &spec); e != nil {
		return err("invalid JSON: " + e.Error())
}

	paths, found := spec["paths"].(map[string]interface{})
	if !found {
		return err("no paths found in spec")
}

	var endpoints []string
	for p := range paths {
		endpoints = append(endpoints, p)

	return ok("endpoints: " + joinStrings(endpoints))
}

}

func joinStrings(s []string) string {
	if len(s) == 0 {
		return "[]"
	}
	result := "["
	for i, v := range s {
		if i > 0 {
			result += ", "
		}
		result += v
	}
	return result + "]"
}