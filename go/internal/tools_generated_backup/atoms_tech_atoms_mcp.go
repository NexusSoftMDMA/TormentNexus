package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleCreateRequirement(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	description, _ :=getString(args, "description")
	apiUrl, _ :=getString(args, "api_url")
	if apiUrl == "" {
		return err("api_url is required")
}

	body := map[string]string{"title": title, "description": description}
	jsonBytes, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	resp, e := http.DefaultClient.Post(apiUrl+"/requirements", "application/json", bytes.NewReader(jsonBytes))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return err("unexpected status: " + resp.Status)
}

	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(fmt.Sprintf("Requirement created: %s", string(respBody)))
}

func HandleListRequirements(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiUrl, _ :=getString(args, "api_url")
	if apiUrl == "" {
		return err("api_url is required")
}

	resp, e := http.DefaultClient.Get(apiUrl + "/requirements")
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("unexpected status: " + resp.Status)
}

	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return success(string(respBody))
}