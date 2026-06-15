package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleSearchSkills(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	resp, e := http.DefaultClient.Get("https://api.example.com/skills?q=" + query)
	if e != nil {
		return err("failed to fetch skills: " + e.Error())
}

	defer resp.Body.Close()
	var result interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to parse response")
}

	return success("Found skills: " + fmt.Sprint(result))
}

func HandleGetSkillDetails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id is required")
}

	return success("Details for skill " + id + ": placeholder")
}