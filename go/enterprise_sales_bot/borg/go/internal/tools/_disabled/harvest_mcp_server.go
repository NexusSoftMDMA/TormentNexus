package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListProjects(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "api_token")
	accountID, _ :=getString(args, "account_id")
	if token == "" || accountID == "" {
		return err("missing api_token or account_id")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.harvestapp.com/v2/projects", nil)
	if e != nil {
		return err(e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Harvest-Account-Id", accountID)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(e.Error())
}

	return ok(fmt.Sprintf("Projects: %v", result))
}

func HandleGetTimeEntries(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "api_token")
	accountID, _ :=getString(args, "account_id")
	if token == "" || accountID == "" {
		return err("missing api_token or account_id")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.harvestapp.com/v2/time_entries", nil)
	if e != nil {
		return err(e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Harvest-Account-Id", accountID)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(e.Error())
}

	return ok(fmt.Sprintf("Time entries: %v", result))
}