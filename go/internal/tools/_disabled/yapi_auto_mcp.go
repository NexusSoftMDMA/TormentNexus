package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleGetInterface(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectID, _ :=getString(args, "project_id")
	interfaceID, _ :=getString(args, "interface_id")
	token, _ :=getString(args, "token")
	baseURL, _ :=getString(args, "base_url")
	if baseURL == "" {
		baseURL = "https://yapi.example.com"
	}
	url := fmt.Sprintf("%s/api/interface/get?id=%s&project_id=%s&token=%s", baseURL, interfaceID, projectID, token)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch interface: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	return ok(string(body))
}