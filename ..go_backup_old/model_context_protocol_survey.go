package tools

import "context"

// HandleSurvey answers a question about the Model Context Protocol.
func HandleSurvey(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	question, _ :=getString(args, "question")
	if question == "" {
		return err("question is required")
}

	answer := "The Model Context Protocol (MCP) enables context sharing between models."
	return success(answer)
}