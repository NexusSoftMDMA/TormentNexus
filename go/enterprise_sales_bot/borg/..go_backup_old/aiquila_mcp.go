package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleNextcloudUserInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	server, _ :=getString(args, "server_url")
	user, _ :=getString(args, "username")
	pass, _ :=getString(args, "password")
	url := fmt.Sprintf("%s/ocs/v2.php/cloud/user?format=json", server)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth(user, pass)
	req.Header.Set("OCS-APIRequest", "true")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status %d", resp.StatusCode))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(result)
}

func HandleNextcloudListFiles(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	server, _ :=getString(args, "server_url")
	user, _ :=getString(args, "username")
	pass, _ :=getString(args, "password")
	path, _ :=getString(args, "path")
	url := fmt.Sprintf("%s/remote.php/dav/files/%s/%s", server, user, path)
	req, e := http.NewRequestWithContext(ctx, "PROPFIND", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.SetBasicAuth(user, pass)
	req.Header.Set("Depth", "1")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMultiStatus {
		return err(fmt.Sprintf("unexpected status %d", resp.StatusCode))
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}