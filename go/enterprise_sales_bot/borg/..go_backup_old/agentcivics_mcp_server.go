package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleRegisterIdentity(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	owner, _ :=getString(args, "owner")
	if name == "" || owner == "" {
		return err("name and owner are required")
}

	body := fmt.Sprintf(`{"name":"%s","owner":"%s"}`, name, owner)
	req, e := http.NewRequestWithContext(ctx, "POST", "https://agentcivics.ai/api/identities", strings.NewReader(body))
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return err("unexpected status: " + resp.Status)
}

	return ok("Identity registered")
}

func HandleGetIdentity(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://agentcivics.ai/api/identities/"+id, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("identity not found")
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	return success(fmt.Sprintf("Identity: %+v", result))
}