package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleGetRevitElements(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	category, _ :=getString(args, "category")
	if category == "" {
		return err("category is required")
}

	req, e := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/revit/api/elements?category="+category, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to connect to Revit: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("Revit error: " + string(body))
}

	return ok(string(body))
}