package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func HandleEvaluate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	expr, _ :=getString(args, "expression")
	if expr == "" {
		return err("expression is required")
}

	req, e := http.NewRequestWithContext(ctx, "POST", "http://127.0.0.1:5000/evaluate?expression="+expr, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("failed to evaluate: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	return ok(string(body))
}