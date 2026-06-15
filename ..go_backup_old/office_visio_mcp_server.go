package tools

import "context"

func HandleListVisioDocuments(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	documents := map[string]interface{}{
		"documents": []map[string]string{
			{"id": "1", "name": "Diagram1.vsdx"},
			{"id": "2", "name": "Diagram2.vsdx"},
		},
	}
	return ok(documents)
}

func HandleGetVisioDocument(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	docID, _ :=getString(args, "document_id")
	docInfo := map[string]string{
		"id":    docID,
		"name":  "Diagram " + docID + ".vsdx",
		"pages": "3",
	}
	return ok(docInfo)
}