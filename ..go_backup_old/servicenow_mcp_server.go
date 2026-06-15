package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleCreateIncident(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	instance, _ :=getString(args, "instance")
	user, _ :=getString(args, "username")
	pass, _ :=getString(args, "password")
	desc, _ :=getString(args, "short_description")
	url := fmt.Sprintf("https://%s.service-now.com/api/now/table/incident", instance)
	body := map[string]string{"short_description": desc}
	payload, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth(user, pass)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return err(fmt.Sprintf("unexpected status %d", resp.StatusCode))
}

	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(respBody))
}

func HandleGetIncident(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	instance, _ :=getString(args, "instance")
	user, _ :=getString(args, "username")
	pass, _ :=getString(args, "password")
	sysID, _ :=getString(args, "sys_id")
	url := fmt.Sprintf("https://%s.service-now.com/api/now/table/incident/%s", instance, sysID)
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth(user, pass)
	req.Header.Set("Accept", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status %d", resp.StatusCode))
}

	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(respBody))
}