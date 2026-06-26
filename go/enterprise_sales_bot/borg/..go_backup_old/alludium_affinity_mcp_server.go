package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleListContacts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	baseURL := "https://api.affinity.co"
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/contacts", nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("X-API-Key", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch contacts")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return success(string(body))
}

func HandleGetContact(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "apiKey")
	contactId, _ :=getString(args, "contactId")
	baseURL := "https://api.affinity.co"
	url := fmt.Sprintf("%s/contacts/%s", baseURL, contactId)
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("X-API-Key", apiKey)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch contact")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return success(string(body))
}