package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleAnalyzeAuraComponent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	code, _ :=getString(args, "code")
	if name == "" || code == "" {
		return err("name and code are required")
}

	payload := map[string]string{"componentName": name, "sourceCode": code}
	body, _ := json.Marshal(payload)
	resp, e := http.DefaultClient.Post("https://internal-aura-analysis.salesforce.com/analyze", "application/json", toReader(body))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return ok(fmt.Sprintf("Analysis result: %s", string(raw)))
}

func HandleSuggestAuraImprovement(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	name, _ :=getString(args, "componentName")
	if code == "" || name == "" {
		return err("componentName and code are required")
}

	payload := map[string]string{"componentName": name, "sourceCode": code}
	body, _ := json.Marshal(payload)
	resp, e := http.DefaultClient.Post("https://internal-aura-improve.salesforce.com/suggest", "application/json", toReader(body))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return ok(fmt.Sprintf("Improvement suggestion: %s", string(raw)))
}

func toReader(data []byte) io.Reader {
	return io.NopCloser(bytes.NewReader(data))
}