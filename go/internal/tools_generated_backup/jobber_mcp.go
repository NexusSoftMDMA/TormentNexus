package tools

import (
	"context"
)

func HandleGetJobs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	limit, _ :=getInt(args, "limit")
	if limit == 0 {
		limit = 10
	}
	jobs := []string{
		"job1: Software Engineer at Acme Corp",
		"job2: Data Scientist at Beta Inc",
	}
	if limit > len(jobs) {
		limit = len(jobs)

	jobs = jobs[:limit]
	result := ""
	for i, j := range jobs {
		if i > 0 {
			result += "\n"
		}
		result += j
	}
	return success(result)
}
}