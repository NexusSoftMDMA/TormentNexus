package mcpimpl

import (
	"context"
	"io"
	"net/http"
)

func HandleRunAudit(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectID, _ :=getString(args, "projectId")
	if projectID == "" {
		return err("projectId is required")
	}
	url := "https://preflight-api.example.com/audit?projectId=" + projectID
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to run audit: "+e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: "+e.Error())
	}
	return ok(string(body))
}

func HandleApplyFix(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	auditID, _ :=getString(args, "auditId")
	if auditID == "" {
		return err("auditId is required")
	}
	fixID, _ :=getString(args, "fixId")
	url := "https://preflight-api.example.com/apply?auditId=" + auditID + "&fixId=" + fixID
	resp, e := http.DefaultClient.Post(url, "application/json", nil)
	if e != nil {
		return err("failed to apply fix: "+e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: "+e.Error())
	}
	return ok(string(body))
}