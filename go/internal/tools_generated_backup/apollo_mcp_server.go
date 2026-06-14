package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func HandleApolloQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	endpoint, _ :=getString(args, "endpoint")
	if endpoint == "" {
		return err("endpoint is required")
}

	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	payload, e := json.Marshal(map[string]string{"query": query})
	if e != nil {
		return err("failed to marshal query: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(payload))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}