package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleListTraces(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://traces.example.com/api/traces")
	if e != nil {
		return err(fmt.Sprintf("failed to list traces: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	return ok(fmt.Sprintf("Traces: %s", string(body)))
}

func HandleGetTrace(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	traceID, _ :=getString(args, "traceId")
	resp, e := http.DefaultClient.Get(fmt.Sprintf("https://traces.example.com/api/traces/%s", traceID))
	if e != nil {
		return err(fmt.Sprintf("failed to get trace: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	return ok(fmt.Sprintf("Trace data: %s", string(body)))
}