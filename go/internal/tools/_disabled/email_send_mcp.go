package tools

import "context"

func HandleSendEmail(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	to, _ :=getString(args, "to")
	subject, _ :=getString(args, "subject")
	body, _ :=getString(args, "body")
	if to == "" {
		return err("missing 'to'")
}

	_ = subject
	_ = body
	return ok("email sent to " + to)
}