package mcpimpl

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleListCertifications(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.example.com/certs")
	if e != nil {
		return err(fmt.Sprintf("failed to fetch certifications: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	return ok(string(body))
}