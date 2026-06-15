//go:build ignore
// +build ignore

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListApps(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	server, _ :=getString(args, "server")
	token, _ :=getString(args, "token")
	if server == "" || token == "" {
		return err("server and token are required")
}

	url := fmt.Sprintf("%s/api/v1/applications", server)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("unexpected status %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(e.Error())
}

	items, found := result["items"]
	if !found {
		return ok("[]")
}

	b, _ := json.Marshal(items)
	return ok(string(b))
}

func HandleGetApp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	server, _ :=getString(args, "server")
	token, _ :=getString(args, "token")
	name, _ :=getString(args, "name")
	if server == "" || token == "" || name == "" {
		return err("server, token and name are required")
}

	url := fmt.Sprintf("%s/api/v1/applications/%s", server, name)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("unexpected status %d: %s", resp.StatusCode, string(body)))
}

	return ok(string(body))
}
