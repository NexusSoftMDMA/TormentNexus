package tools

import (
	"context"
	"encoding/json"
	"net/http"
)

// HandleParseResume parses a resume and extracts skills.
func HandleParseResume(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resume, _ :=getString(args, "resume")
	if resume == "" {
		return err("resume argument is required")
}

	// Simulate API call (use http.DefaultClient for real integration)
	resp, e := http.DefaultClient.Get("https://api.resume-parser.example.com/parse?text=" + resume)
	if e != nil {
		return err("failed to call resume parser: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	output, _ := json.Marshal(result)
	return success(string(output))
}

// HandleMatchJob matches a resume to a job description.
func HandleMatchJob(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resume, _ :=getString(args, "resume")
	job, _ :=getString(args, "job_description")
	if resume == "" || job == "" {
		return err("resume and job_description arguments are required")
}

	// Simulate API call
	resp, e := http.DefaultClient.Get("https://api.resume-parser.example.com/match?resume=" + resume + "&job=" + job)
	if e != nil {
		return err("failed to call match service: " + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	output, _ := json.Marshal(result)
	return success(string(output))
}