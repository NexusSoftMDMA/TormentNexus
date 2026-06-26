package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleCreateEmployee(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	role, _ :=getString(args, "role")
	if name == "" || role == "" {
		return err("name and role are required")
}

	baseURL := "http://paperclip-internal/api"
	payload := map[string]string{"name": name, "role": role}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal payload")
}

	req, e := http.NewRequestWithContext(ctx, "POST", baseURL+"/employees", bytes.NewReader(body))
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
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	var result struct {
		EmployeeID string `json:"employee_id"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	return success(fmt.Sprintf("created employee %s", result.EmployeeID))
}

func HandleGetEmployeeStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	employeeID, _ :=getString(args, "employee_id")
	if employeeID == "" {
		return err("employee_id is required")
}

	baseURL := "http://paperclip-internal/api"
	req, e := http.NewRequestWithContext(ctx, "GET", baseURL+"/employees/"+employeeID, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("status %d: %s", resp.StatusCode, string(body)))
}

	var result struct {
		Status string `json:"status"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	return ok(fmt.Sprintf("employee status: %s", result.Status))
}