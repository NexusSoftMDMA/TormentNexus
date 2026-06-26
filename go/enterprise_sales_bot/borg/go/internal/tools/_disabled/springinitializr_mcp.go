package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleGetMetadata(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://start.spring.io/dependencies.json")
	if e != nil {
		return err("Failed to fetch metadata: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("Failed to read response: " + e.Error())
}

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("Failed to parse JSON: " + e.Error())
}

	jsonStr, _ := json.MarshalIndent(data, "", "  ")
	return ok(string(jsonStr))
}