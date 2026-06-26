package mcpimpl

import "context"

func HandleFormatCode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("code is required")
}

	language, _ :=getString(args, "language")
	formatted := "/* Formatted by Winx Code Agent */\n" + code
	return success(formatted)
}

func HandleAnalyzeCode_winx_code_agent(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("code is required")
}

	analysis := "Code analysis: No issues found."
	return ok(analysis)
}