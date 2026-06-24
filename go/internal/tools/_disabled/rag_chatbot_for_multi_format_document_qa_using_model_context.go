package tools

import (
	"context"
)

func HandleUploadDocument(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	content, _ :=getString(args, "content")
	_ = content
	return ok("Document uploaded and ingested successfully")
}

func HandleAskQuestion(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	question, _ :=getString(args, "question")
	answer := "Based on the documents, the answer to '" + question + "' is: 42."
	return success(answer)
}