package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleDocQA(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	flowID, _ :=getString(args, "flow_id")
	apiKey, _ :=getString(args, "api_key")

	if query == "" || flowID == "" {
		return err("query and flow_id are required")
}

	body, _ := json.Marshal(map[string]interface{}{"input": query})
	req, e := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("https://api.langflow.com/v1/process/%s", flowID),
		bytes.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	output, found := result["output"]
	if !found {
		return err("no output in response")
}

	return ok(fmt.Sprint(output))
}
}