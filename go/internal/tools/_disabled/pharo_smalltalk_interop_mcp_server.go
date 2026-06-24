package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleExecutePharoCode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("code argument is required")
}

	reqBody, e := json.Marshal(map[string]string{"code": code})
	if e != nil {
		return err(fmt.Sprintf("failed to marshal request: %v", e))
}

	resp, e := http.DefaultClient.Post("http://localhost:8080/evaluate", "application/json", bytes.NewBuffer(reqBody))
	if e != nil {
		return err(fmt.Sprintf("HTTP request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("failed to read response: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("Pharo server returned status %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err(fmt.Sprintf("failed to parse response: %v", e))
}

	output, found := result["output"]
	if !found {
		return err("response missing output field")
}

	return ok(fmt.Sprintf("Pharo output: %v", output))
}