package tools

import "context"

func HandleReadDocument(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	collection, _ :=getString(args, "collection")
	docId, _ :=getString(args, "docId")
	if collection == "" || docId == "" {
		return err("Missing required parameters")
}

	return ok("Document read: " + collection + "/" + docId)
}

func HandleWriteDocument(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	collection, _ :=getString(args, "collection")
	docId, _ :=getString(args, "docId")
	data, _ :=getString(args, "data")
	if collection == "" || docId == "" || data == "" {
		return err("Missing required parameters")
}

	return success("Document written successfully")
}