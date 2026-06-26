package mcpimpl

import (
	"context"
	"fmt"
	"strings"
)

func HandleReviewCode_code_reviewer_ai_mcp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("code is required")
}

	suggestions := []string{}
	if strings.Contains(code, "fmt.Println") {
		suggestions = append(suggestions, "Consider using a logger instead of fmt.Println")

	if strings.Contains(code, "TODO") {
		suggestions = append(suggestions, "TODO comments should be resolved before production")

	if len(suggestions) == 0 {
		return success("Code looks clean, no suggestions.")
}

	return ok(fmt.Sprintf("Review suggestions: %s", strings.Join(suggestions, "; ")))
}

}
}

func HandleCheckSecurity(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if code == "" {
		return err("code is required")
}

	issues := []string{}
	if strings.Contains(code, "exec(") || strings.Contains(code, "Cmd(") {
		issues = append(issues, "Potential command injection risk")

	if strings.Contains(code, "fmt.Sprintf(\"%v\"") {
		issues = append(issues, "Use parameterized queries to prevent SQL injection")

	if len(issues) == 0 {
		return success("No security issues found.")
}

	return ok(fmt.Sprintf("Security issues: %s", strings.Join(issues, "; ")))
}
}
}