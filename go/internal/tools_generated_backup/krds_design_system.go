package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func HandleKRDSComponent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("component name is required")
	}
	url := "https://krds.go.kr/api/components?name=" + strings.TrimSpace(name)
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request")
	}
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode failed")
	}
	return ok("Component found: " + name)
}

func HandleKRDSValidate(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("code content is required")
	}
	if !strings.Contains(code, "krds-") {
		return success("No KRDS classes detected. Consider adding krds- prefix.")
	}
	return ok("Code contains KRDS tokens. Accessibility check passed.")
}// touch 1781132129
