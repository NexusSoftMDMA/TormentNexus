package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleListJobs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	addr, _ :=getString(args, "address")
	if addr == "" {
		addr = "http://localhost:4646"
	}
	resp, e := http.DefaultClient.Get(addr + "/v1/jobs")
	if e != nil {
		return err("failed to list jobs: " + e.Error())
}

	defer resp.Body.Close()
	var jobs []map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&jobs); e != nil {
		return err("failed to parse jobs: " + e.Error())
}

	var names []string
	for _, j := range jobs {
		if name, found := j["Name"].(string); found {
			names = append(names, name)

	}
	return ok(strings.Join(names, "\n"))
}

}

func HandleReadJob(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	addr, _ :=getString(args, "address")
	if addr == "" {
		addr = "http://localhost:4646"
	}
	jobID, _ :=getString(args, "job_id")
	if jobID == "" {
		return err("job_id is required")
}

	resp, e := http.DefaultClient.Get(addr + "/v1/job/" + jobID)
	if e != nil {
		return err("failed to read job: " + e.Error())
}

	defer resp.Body.Close()
	var job map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&job); e != nil {
		return err("failed to parse job: " + e.Error())
}

	data, _ := json.MarshalIndent(job, "", "  ")
	return ok(string(data))
}