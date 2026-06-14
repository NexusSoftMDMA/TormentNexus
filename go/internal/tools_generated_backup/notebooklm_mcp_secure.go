package tools

import (
	"context"
	"net/http"
	"io"
)

func HandleSearchNotebooks(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	url := "https://notebooklm.googleapis.com/v1/notebooks?query=" + query
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to search notebooks: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return success(string(body))
}

func HandleGetNotebook(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	url := "https://notebooklm.googleapis.com/v1/notebooks/" + id
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to get notebook: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return success(string(body))
}