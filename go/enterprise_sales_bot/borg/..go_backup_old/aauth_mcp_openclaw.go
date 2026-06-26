package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListClaws(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	baseURL, _ :=getString(args, "base_url")
	token, _ :=getString(args, "token")
	if baseURL == "" || token == "" {
		return err("base_url and token required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/claws", nil)
	if e != nil {
		return err("request creation: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read body: " + e.Error())
}

	var result interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("parse: " + e.Error())
}

	return ok(fmt.Sprintf("claws: %v", result))
}