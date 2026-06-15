//go:build ignore
// +build ignore

package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleListContainers(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	account, _ :=getString(args, "storageAccount")
	req, e := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("https://%s.dfs.core.windows.net/?resource=account", account), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(string(body))
}

func HandleReadFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	account, _ :=getString(args, "storageAccount")
	container, _ :=getString(args, "container")
	filePath, _ :=getString(args, "filePath")
	req, e := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("https://%s.dfs.core.windows.net/%s/%s", account, container, filePath), nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(string(body))
}
