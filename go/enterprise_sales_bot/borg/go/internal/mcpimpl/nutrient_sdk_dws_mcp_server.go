package mcpimpl

import "context"

func HandleProcessDocument(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	docID, _ :=getString(args, "document_id")
	if docID == "" {
		return err("document_id is required")
}

	return ok("Document " + docID + " processed successfully")
}