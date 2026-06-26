package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleListDocs_yuque_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	repoID, _ :=getString(args, "repo_id")
	if repoID == "" {
		return err("repo_id is required")
}

	token := os.Getenv("YUQUE_TOKEN")
	if token == "" {
		return err("YUQUE_TOKEN environment variable not set")
}

	url := fmt.Sprintf("https://yuque.com/api/v2/repos/%s/docs", repoID)
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err("API returned status " + fmt.Sprint(resp.StatusCode) + ": " + string(body))
}

	var result struct {
		Data []struct {
			ID    int    `json:"id"`
			Title string `json:"title"`
			Slug  string `json:"slug"`
		} `json:"data"`
	}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	if len(result.Data) == 0 {
		return ok("No documents found in this repo")
}

	var list string
	for _, doc := range result.Data {
		list += fmt.Sprintf("- %s (slug: %s, id: %d)\n", doc.Title, doc.Slug, doc.ID)

	return success("Documents in repo " + repoID + ":\n" + list)
}
}