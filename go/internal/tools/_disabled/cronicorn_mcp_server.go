package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func HandleListCronJobs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("http://localhost:8080/cron")
	if e != nil {
		return err("failed to list cron jobs: " + e.Error())
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return ok(string(body))
}

func HandleCreateCronJob(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	schedule, _ :=getString(args, "schedule")
	command, _ :=getString(args, "command")
	if schedule == "" || command == "" {
		return err("schedule and command are required")
}

	payload, _ := json.Marshal(map[string]string{"schedule": schedule, "command": command})
	resp, e := http.DefaultClient.Post("http://localhost:8080/cron", "application/json", bytes.NewReader(payload))
	if e != nil {
		return err("failed to create cron job: " + e.Error())
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return success("cron job created: " + string(body))
}