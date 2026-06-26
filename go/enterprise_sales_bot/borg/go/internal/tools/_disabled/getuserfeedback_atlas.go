package tools

import (
	"context"
	"io"
	"net/http"
	"os"
)

func HandleGetFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	owner, _ :=getString(args, "owner")
	repo, _ :=getString(args, "repo")
	path, _ :=getString(args, "path")
	if owner == "" || repo == "" || path == "" {
		return err("owner, repo, and path are required")
}

	base := os.Getenv("ATLAS_BASE_URL")
	if base == "" {
		base = "https://api.github.com"
	}
	url := base + "/repos/" + owner + "/" + repo + "/contents/" + path
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(string(body))
}

func HandleListDir(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	owner, _ :=getString(args, "owner")
	repo, _ :=getString(args, "repo")
	path, _ :=getString(args, "path")
	if owner == "" || repo == "" {
		return err("owner and repo are required")
}

	base := os.Getenv("ATLAS_BASE_URL")
	if base == "" {
		base = "https://api.github.com"
	}
	url := base + "/repos/" + owner + "/" + repo + "/contents/"
	if path != "" {
		url += path
	}
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("read failed: " + e.Error())
}

	return ok(string(body))
}