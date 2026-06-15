package tools

import (
	"context"
	"fmt"
	"net/http"
)

func HandleGetInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	resp, e := http.Get("https://api.example.com/iranti?q=" + query)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("status %d", resp.StatusCode))
}

	return success(fmt.Sprintf("Info for '%s' retrieved", query))
}