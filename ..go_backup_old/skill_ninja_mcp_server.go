package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// HandleSearchSkills searches for available skills by query.
func HandleSearchSkills(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	url := fmt.Sprintf("https://api.skill-ninja.com/v1/skills?q=%s", query)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("search request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read response failed: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse response failed: %v", e))
}

	skills, found := result["skills"].([]interface{})
	if !found {
		return ok("[]")
}

	return ok(fmt.Sprintf("%v", skills))
}

// HandleInstallSkill installs a skill by its ID.
func HandleInstallSkill(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	skillID, _ :=getString(args, "skillId")
	url := fmt.Sprintf("https://api.skill-ninja.com/v1/install/%s", skillID)
	resp, e := http.DefaultClient.Post(url, "application/json", nil)
	if e != nil {
		return err(fmt.Sprintf("install request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read response failed: %v", e))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("parse response failed: %v", e))
}

	status, found := result["status"].(string)
	if !found {
		return ok("installed")
}

	return ok(status)
}