package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleGetPlant(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("plant name is required")
}

	baseURL := "https://api.flora.example.com"
	url := fmt.Sprintf("%s/plants?name=%s", baseURL, name)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
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
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("FLORA API returned status " + resp.Status)
}

	return ok(fmt.Sprintf("Plant info: %s", string(body)))
}