package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleRecordAttendance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	employeeID, _ :=getString(args, "employee_id")
	eventType, _ :=getString(args, "type")
	timestamp, _ :=getString(args, "timestamp")
	if employeeID == "" || eventType == "" {
		return err("missing required fields")
}

	body := fmt.Sprintf(`{"employee_id":"%s","type":"%s","timestamp":"%s"}`, employeeID, eventType, timestamp)
	req, e := http.NewRequestWithContext(ctx, "POST", "http://localhost:8080/attendance", strings.NewReader(body))
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return err("attendance server error")
}

	return ok("attendance recorded")
}

func HandleGetAttendance(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	employeeID, _ :=getString(args, "employee_id")
	if employeeID == "" {
		return err("employee_id required")
}

	url := fmt.Sprintf("http://localhost:8080/attendance/%s", employeeID)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	data, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	return success(string(data))
}