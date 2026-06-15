package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleLogWork(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	issueKey, _ :=getString(args, "issueKey")
	timeSpent, _ :=getString(args, "timeSpent")
	comment, _ :=getString(args, "comment")
	if issueKey == "" || timeSpent == "" {
		return err("issueKey and timeSpent are required")
	}
	payload := map[string]interface{}{"timeSpent": timeSpent}
	if comment != "" {
		payload["comment"] = comment
	}
	body, e := json.Marshal(payload)
	if e != nil {
		return err(e.Error())
	}
	req, e := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("https://api.tempo.io/core/3/worklogs?issueKey=%s", issueKey), strings.NewReader(string(body)))
	if e != nil {
		return err(e.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return err(fmt.Sprintf("Tempo API error: %d", resp.StatusCode))
	}
	return success("Worklog added successfully")
}

func HandleGetWorklogs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	issueKey, _ :=getString(args, "issueKey")
	if issueKey == "" {
		return err("issueKey is required")
	}
	req, e := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.tempo.io/core/3/worklogs?issueKey=%s", issueKey), nil)
	if e != nil {
		return err(e.Error())
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return err(fmt.Sprintf("Tempo API error: %d", resp.StatusCode))
	}
	return ok("Worklogs retrieved successfully")
}// touch 1781132128
