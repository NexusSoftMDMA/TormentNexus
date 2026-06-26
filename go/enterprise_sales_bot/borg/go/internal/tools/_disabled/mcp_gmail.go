package tools

import (
	"context"
)

func HandleSearchEmails(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	return ok("Search performed for query: " + query)
}

func HandleSendEmail(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	to, _ :=getString(args, "to")
	subject, _ :=getString(args, "subject")
	body, _ :=getString(args, "body")
	return ok("Email sent to " + to + " with subject: " + subject + " body: " + body)
}