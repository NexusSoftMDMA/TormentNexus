package tools

import "context"

func HandleCompileLatex(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	latex, _ :=getString(args, "latex_code")
	if latex == "" {
		return err("latex_code is required")
}

	return ok("LaTeX compilation started")
}