package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleListContexts_arkon(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "server_url")
	if base == "" {
		base = "http://localhost:8080"
	}
	req, e := http.NewRequestWithContext(ctx, "GET", base+"/api/contexts", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var contexts []map[string]interface{}
	if e := json.Unmarshal(body, &contexts); e != nil {
		return err("failed to parse response: " + e.Error())
}

	data, _ := json.Marshal(contexts)
	return ok(fmt.Sprintf("Found %d contexts: %s", len(contexts), string(data)))
}