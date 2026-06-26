package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

func HandleSearchJobs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter required")
}

	jobs := []map[string]interface{}{
		{"id": "1", "title": "Software Engineer", "company": "Tech Corp", "location": "Remote"},
		{"id": "2", "title": "Product Manager", "company": "Biz Inc", "location": "New York"},
	}
	data, e := json.Marshal(jobs)
	if e != nil {
		return err(fmt.Sprintf("failed to marshal jobs: %v", e))
}

	return ok(string(data))
}

func HandleGetJob(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("id parameter required")
}

	job := map[string]interface{}{
		"id":          id,
		"title":       "Software Engineer",
		"company":     "Tech Corp",
		"description": "Develop and maintain software applications.",
		"location":    "Remote",
	}
	data, e := json.Marshal(job)
	if e != nil {
		return err(fmt.Sprintf("failed to marshal job: %v", e))
}

	return ok(string(data))
}