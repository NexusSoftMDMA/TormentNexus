package tools

import "context"

func HandleCodeAssistant(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	language, _ :=getString(args, "language")
	if code == "" || language == "" {
		return err("Missing required parameters: code and language")
}

	return success("Processed " + language + " code successfully")
}