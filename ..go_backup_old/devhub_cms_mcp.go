package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleGetContent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	req, e := http.NewRequestWithContext(ctx, "GET", "https://api.devhub.com/content/"+id, nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch content")
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	return success(string(body))
}