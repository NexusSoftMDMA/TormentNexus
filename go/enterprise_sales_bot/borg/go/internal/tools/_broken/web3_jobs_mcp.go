package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleListWeb3Jobs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	keyword, _ :=getString(args, "keyword")
	location, _ :=getString(args, "location")
	remote, _ :=getBool(args, "remote")

	url := "https://api.web3.career/jobs?keyword=" + keyword + "&location=" + location
	if remote {
		url += "&remote=true"
	}
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch jobs: " + e.Error())
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var jobs []map[string]interface{}
	if e := json.Unmarshal(body, &jobs); e != nil {
		return err("failed to parse jobs: " + e.Error())
}

	result := ""
	for _, job := range jobs {
		title, _ := job["title"].(string)
		company, _ := job["company"].(string)
		result += fmt.Sprintf("- %s at %s\n", title, company)

	if result == "" {
		result = "No jobs found."
	}
	return ok(result)
}

}

func HandleGetWeb3Job(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getInt(args, "id")
	url := fmt.Sprintf("https://api.web3.career/jobs/%d", id)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch job: " + e.Error())
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var job map[string]interface{}
	if e := json.Unmarshal(body, &job); e != nil {
		return err("failed to parse job: " + e.Error())
}

	title, _ := job["title"].(string)
	company, _ := job["company"].(string)
	description, _ := job["description"].(string)
	result := fmt.Sprintf("Title: %s\nCompany: %s\nDescription: %s", title, company, description)
	return ok(result)
}