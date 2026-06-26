package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleCreateSandbox(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	templateID, _ :=getString(args, "template_id")
	body := fmt.Sprintf(`{"template_id":"%s"}`, templateID)
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.e2b.dev/v1/sandboxes", strings.NewReader(body))
	if e != nil {
		return err(e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	var result struct{ ID string `json:"id"` }
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(e.Error())
}

	return ok(fmt.Sprintf("Sandbox created: %s", result.ID))
}

func HandleExecuteCode_e2b_sandbox_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	sandboxID, _ :=getString(args, "sandbox_id")
	code, _ :=getString(args, "code")
	language, _ :=getString(args, "language")
	body := fmt.Sprintf(`{"code":%q,"language":%q}`, code, language)
	url := fmt.Sprintf("https://api.e2b.dev/v1/sandboxes/%s/code", sandboxID)
	req, e := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(body))
	if e != nil {
		return err(e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(e.Error())
}

	defer resp.Body.Close()
	out, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(e.Error())
}

	return ok(string(out))
}