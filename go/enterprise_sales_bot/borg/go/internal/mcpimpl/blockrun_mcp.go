package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleListLanguages_blockrun_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.blockrun.dev/v1/languages")
	if e != nil {
		return err("failed to fetch languages:" + e.Error())
}

	defer resp.Body.Close()
	var langs []string
	if e := json.NewDecoder(resp.Body).Decode(&langs); e != nil {
		return err("decode error:" + e.Error())
}

	return ok(fmt.Sprintf("Available languages: %s", strings.Join(langs, ", ")))
}

func HandleExecuteCode_blockrun_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	lang, _ :=getString(args, "language")
	code, _ :=getString(args, "code")
	if lang == "" || code == "" {
		return err("language and code are required")
}

	body := fmt.Sprintf(`{"language":"%s","code":%q}`, lang, code)
	resp, e := http.DefaultClient.Post("https://api.blockrun.dev/v1/execute", "application/json", strings.NewReader(body))
	if e != nil {
		return err("request error:" + e.Error())
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("decode error:" + e.Error())
}

	if out, found := result["output"]; found {
		return ok(fmt.Sprintf("%v", out))
}

	return ok("execution completed (no output)")
}