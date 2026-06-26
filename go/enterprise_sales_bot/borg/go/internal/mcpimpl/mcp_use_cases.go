package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"io"
	"strings"
)

func HandleListUseCases(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// Mock list of use cases
	useCases := []string{"Translation Assistant", "Code Review", "Meeting Summarizer"}
	body, _ := json.Marshal(useCases)
	return ok(fmt.Sprintf("Available use cases: %s", string(body)))
}

func HandleGetUseCase(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name parameter is required")
}

	// Simulate fetching details from an external API
	url := fmt.Sprintf("https://api.mcpusecases.com/use-cases/%s", name)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %s", e.Error()))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %s", e.Error()))
}

	defer resp.Body.Close()
	data, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %s", e.Error()))
}

	// For demo, we simply return the raw response as a string
	detail := strings.TrimSpace(string(data))
	if detail == "" {
		detail = "No details available at this time."
	}
	return success(detail)
}