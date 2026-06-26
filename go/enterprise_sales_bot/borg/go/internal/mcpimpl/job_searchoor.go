package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func HandleSearchJobs_job_searchoor(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	location, _ :=getString(args, "location")
	if query == "" {
		return err("query is required")
}

	u := fmt.Sprintf("https://api.jobsearchoor.com/search?q=%s&location=%s", url.QueryEscape(query), url.QueryEscape(location))
	resp, e := http.DefaultClient.Get(u)
	if e != nil {
		return err("failed to fetch jobs")
}

	defer resp.Body.Close()
	var result struct {
		Jobs []map[string]interface{} `json:"jobs"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("invalid response")
}

	if len(result.Jobs) == 0 {
		return ok("no jobs found")
}

	return success(fmt.Sprintf("found %d jobs", len(result.Jobs)))
}