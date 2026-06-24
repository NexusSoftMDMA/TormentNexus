package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"bytes"
)

func HandleListChanges(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	project, _ :=getString(args, "project")
	resp, e := http.Get(fmt.Sprintf("http://polygraph.local/api/changes?project=%s", project))
	if e != nil {
		return err("failed to fetch changes: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("polygraph API returned status " + resp.Status)
}

	var changes []map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&changes); e != nil {
		return err("failed to decode changes: " + e.Error())
}

	return ok(fmt.Sprintf("Found %d changes", len(changes)))
}

func HandleCreateChange(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	desc, _ :=getString(args, "description")
	repos, _ :=getString(args, "repos")
	body, _ := json.Marshal(map[string]string{
		"title":       title,
		"description": desc,
		"repos":       repos,
	})
	resp, e := http.DefaultClient.Post("http://polygraph.local/api/changes", "application/json", bytes.NewReader(body))
	if e != nil {
		return err("failed to create change: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		return err("polygraph API returned status " + resp.Status)
}

	return success("Change created successfully")
}