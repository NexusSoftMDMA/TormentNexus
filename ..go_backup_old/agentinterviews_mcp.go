package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGetInterview(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "agent")
	if name == "" {
		return err("missing agent parameter")
}

	url := fmt.Sprintf("https://example.com/interviews?agent=%s", name)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	var data map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&data); e != nil {
		return err("decode failed: " + e.Error())
}

	result, found := data["result"].(string)
	if !found {
		return err("result not found")
}

	return ok(result)
}