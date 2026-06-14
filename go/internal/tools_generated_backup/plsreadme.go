package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleReadReadme(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	repo, _ :=getString(args, "repo")
	if repo == "" {
		return err("missing 'repo' argument (format: owner/repo)")
}

	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/master/README.md", repo)
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if e != nil {
		return err("request error: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("fetch error: " + e.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err("unexpected status: " + resp.Status)
}

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read error: " + e.Error())
}

	return ok(string(body))
}