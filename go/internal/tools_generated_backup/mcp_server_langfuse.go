package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func HandleGetTrace(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	base, _ :=getString(args, "baseUrl")
	if base == "" {
		return err("baseUrl is required")
	}
	projectID, _ :=getString(args, "projectId")
	traceID, _ :=getString(args, "traceId")
	limit, _ :=getInt(args, "limit")
	if limit == 0 {
		limit = 10
	}
	url := fmt.Sprintf("%s/api/public/traces?projectId=%s&limit=%d", base, projectID, limit)
	if traceID != "" {
		url = fmt.Sprintf("%s/api/public/traces/%s", base, traceID)

	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err(fmt.Sprintf("HTTP request failed: %v", e))
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body: %v", e))
	}
	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API returned %d: %s", resp.StatusCode, string(body)))
	}
	var data interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err(fmt.Sprintf("unmarshal: %v", e))
	}
	return success(fmt.Sprintf("Traces: %s", string(body)))
}
}