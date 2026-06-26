package tools

import (
	"context"
	"io"
	"net/http"
	"os"
)

func HandleDownloadRules(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("url is required")
}

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch rules: " + e.Error())
}

	defer resp.Body.Close()
	data, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	filename, _ :=getString(args, "filename")
	if filename == "" {
		filename = "rules.json"
	}
	e = os.WriteFile(filename, data, 0644)
	if e != nil {
		return err("failed to write file: " + e.Error())
}

	return success("rules downloaded to " + filename)
}

func HandleInitProject(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("project name is required")
}

	dirs := []string{name + "/src", name + "/config", name + "/scripts"}
	for _, d := range dirs {
		e := os.MkdirAll(d, 0755)
		if e != nil {
			return err("failed to create directory " + d + ": " + e.Error())

	}
	return ok("project " + name + " initialized")
}
}