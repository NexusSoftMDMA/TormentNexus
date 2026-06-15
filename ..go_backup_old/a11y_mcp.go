package tools

import (
	"context"
	"net/http"
)

func HandleAudit(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch page: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("page returned status " + resp.Status)
}

	return ok("Page is accessible. Further auditing is not implemented in this demo.")
}