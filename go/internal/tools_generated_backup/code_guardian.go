package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleAnalyzeCode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	language, _ :=getString(args, "language")
	if code == "" {
		return err("code is required")
}

	resp, e := http.DefaultClient.Get(fmt.Sprintf("https://api.codeguardian.dev/analyze?code=%s&lang=%s", code, language))
	if e != nil {
		return err("failed to analyze code")
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("analysis service returned error")
}

	return ok("Code analysis submitted")
}

func HandleGetReport(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	reportID, _ :=getString(args, "report_id")
	if reportID == "" {
		return err("report_id is required")
}

	resp, e := http.DefaultClient.Get(fmt.Sprintf("https://api.codeguardian.dev/report/%s", reportID))
	if e != nil {
		return err("failed to fetch report")
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("report not found")
}

	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("invalid report data")
}

	return success("Report retrieved")
}