package mcpimpl

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleQuery_falkordb_mcpserver(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	resp, e := http.DefaultClient.Post("http://localhost:6379/query", "text/plain", nil)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	return success(string(body))
}

func HandleExecute_falkordb_mcpserver(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	req, e := http.NewRequestWithContext(ctx, "POST", "http://localhost:6379/execute", nil)
	if e != nil {
		return err(fmt.Sprintf("new request failed: %v", e))
}

	req.Header.Set("Content-Type", "text/plain")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("execute failed: %v", e))
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read failed: %v", e))
}

	return success(string(body))
}