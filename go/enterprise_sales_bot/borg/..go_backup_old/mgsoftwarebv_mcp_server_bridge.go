package tools

import (
	"context"
	"io"
	"net/http"
)

func HandleGetTicket(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	ticketID, _ :=getString(args, "ticketId")
	if ticketID == "" {
		return err("ticketId is required")
}

	url := "https://api.mgtickets.com/tickets/" + ticketID
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

	return ok(string(body))
}

func HandleExploreGitHub(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	repo, _ :=getString(args, "repo")
	path, _ :=getString(args, "path")
	if repo == "" || path == "" {
		return err("repo and path are required")
}

	url := "https://api.github.com/repos/" + repo + "/contents/" + path
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

	return ok(string(body))
}