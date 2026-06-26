package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleExecuteCode_codecortex(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	lang, _ :=getString(args, "language")
	if code == "" || lang == "" {
		return err("code and language are required")
}

	resp, e := http.DefaultClient.PostForm("https://api.codecortex.com/execute", nil)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var result struct {
		Output string `json:"output"`
		Error  string `json:"error"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	if result.Error != "" {
		return err(result.Error)
}

	return ok(result.Output)
}

func HandleListLanguages_codecortex(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.codecortex.com/languages")
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	var langs []string
	if e := json.NewDecoder(resp.Body).Decode(&langs); e != nil {
		return err(fmt.Sprintf("decode failed: %v", e))
}

	return success(langs)
}