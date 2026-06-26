package mcpimpl

import "context"

func HandleCreateNewsletter(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	subject, _ :=getString(args, "subject")
	body, _ :=getString(args, "body")
	if subject == "" {
		return err("subject is required")
}

	if body == "" {
		return err("body is required")
}

	return success("Newsletter created")
}

func HandleCaptureSubscriber(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	email, _ :=getString(args, "email")
	name, _ :=getString(args, "name")
	if email == "" {
		return err("email is required")
}

	if name == "" {
		name = "Unknown"
	}
	return success("Subscriber captured")
}