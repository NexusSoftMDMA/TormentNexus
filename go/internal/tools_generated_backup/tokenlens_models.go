package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleModels(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	provider, _ :=getString(args, "provider")
	url := fmt.Sprintf("https://models.dev/api/catalog?provider=%s", provider)
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("invalid JSON: " + e.Error())
}

	out, e := json.MarshalIndent(data, "", "  ")
	if e != nil {
		return err("marshal error: " + e.Error())
}

	return ok(string(out))
}