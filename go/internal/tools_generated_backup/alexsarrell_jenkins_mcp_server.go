package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type jobsResponse struct {
	Jobs []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"jobs"`
}

type jobResponse struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

func HandleListJobs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "url")
	if base == "" {
		return err("missing 'url' argument")
}

	req, e := http.NewRequestWithContext(ctx, "GET", base+"/api/json?tree=jobs[name,url]", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch jobs: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var jr jobsResponse
	if e := json.Unmarshal(body, &jr); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	if len(jr.Jobs) == 0 {
		return ok("No jobs found")
}

	var out string
	for _, j := range jr.Jobs {
		out += fmt.Sprintf("- %s (%s)\n", j.Name, j.URL)

	return ok(out)
}

}

func HandleGetJob(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "url")
	jobName, _ :=getString(args, "jobName")
	if base == "" || jobName == "" {
		return err("missing 'url' and/or 'jobName' arguments")
}

	req, e := http.NewRequestWithContext(ctx, "GET", base+"/job/"+jobName+"/api/json", nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("failed to fetch job: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("Jenkins returned status " + resp.Status)
}

	var job jobResponse
	if e := json.Unmarshal(body, &job); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return ok(fmt.Sprintf("Name: %s\nURL: %s\nDescription: %s\nColor: %s", job.Name, job.URL, job.Description, job.Color))
}