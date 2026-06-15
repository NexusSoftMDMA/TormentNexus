package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func HandleGapAnalysis(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	projectID, _ :=getString(args, "project_id")
	if projectID == "" {
		return err("project_id is required")
}

	threshold, _ :=getInt(args, "threshold")
	if threshold == 0 {
		threshold = 50
	}

	u, e := url.Parse("https://api.qulib.io/gaps")
	if e != nil {
		return err("failed to parse URL: " + e.Error())
}

	q := u.Query()
	q.Set("project_id", projectID)
	q.Set("threshold", fmt.Sprintf("%d", threshold))
	u.RawQuery = q.Encode()

	resp, e := http.DefaultClient.Get(u.String())
	if e != nil {
		return err("API request failed: " + e.Error())
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse JSON: " + e.Error())
}

	return success(fmt.Sprintf("Gap analysis result: %v", result["summary"]))
}

func HandleCreateAssessment(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	title, _ :=getString(args, "title")
	if title == "" {
		return err("title is required")
}

	payload, e := json.Marshal(map[string]string{"title": title})
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	resp, e := http.DefaultClient.Post("https://api.qulib.io/assessments", "application/json", io.NopCloser(reader(payload)))
	if e != nil {
		return err("API request failed: " + e.Error())
}

	defer resp.Body.Close()

	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return ok(string(body))
}

// helper for converting []byte to io.Reader
type reader []byte

func (r reader) Read(p []byte) (int, error) {
	return copy(p, r), nil
}