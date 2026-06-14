package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleListContacts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key is required")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.capsulecrm.com/v2/contacts/", nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
	}
	return ok(fmt.Sprintf("Contacts: %v", result))
}

func HandleCreateContact(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key is required")
	}
	firstName, _ :=getString(args, "first_name")
	lastName, _ :=getString(args, "last_name")
	body, e := json.Marshal(map[string]interface{}{
		"contact": map[string]interface{}{
			"firstName": firstName,
			"lastName":  lastName,
		},
	})
	if e != nil {
		return err(fmt.Sprintf("marshal failed: %v", e))
	}
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.capsulecrm.com/v2/contacts/", bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("create request failed: %v", e))
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return err(fmt.Sprintf("API error: %d", resp.StatusCode))
	}
	return success("Contact created")
}