package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleCreateForm(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
	}
	body := map[string]string{"name": name}
	data, _ := json.Marshal(body)
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.clipform.io/v1/forms", bytes.NewReader(data))
	if e != nil {
		return err("failed to create request: " + e.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
	}
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %s", string(respBody)))
	}
	var result map[string]interface{}
	if e := json.Unmarshal(respBody, &result); e != nil {
		return err("failed to parse response: " + e.Error())
	}
	formID, found := result["id"].(string)
	if !found {
		formID = "unknown"
	}
	return success(fmt.Sprintf("Form created with ID: %s", formID))
}

func HandleListForms(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.clipform.io/v1/forms", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
	}
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %s", string(respBody)))
	}
	var forms []interface{}
	if e := json.Unmarshal(respBody, &forms); e != nil {
		return err("failed to parse response: " + e.Error())
	}
	return success(fmt.Sprintf("Found %d forms", len(forms)))
}