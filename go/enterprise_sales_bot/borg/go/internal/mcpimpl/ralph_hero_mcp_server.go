package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleCreateProjectItem(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectID, _ :=getString(args, "project_id")
	title, _ :=getString(args, "title")
	if projectID == "" || title == "" {
		return err("project_id and title are required")
}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return err("GITHUB_TOKEN not set")
}

	query := fmt.Sprintf(`mutation { addProjectV2ItemById(input: {projectId: "%s", contentId: "%s"}) { item { id } } }`, projectID, title)
	payload := map[string]string{"query": query}
	body, _ := json.Marshal(payload)
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.github.com/graphql", bytes.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return err("API error: " + string(respBody))
}

	return ok("Project item created successfully")
}