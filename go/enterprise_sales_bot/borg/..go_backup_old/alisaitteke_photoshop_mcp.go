package tools

import "context"

func HandleApplyFilter(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filterName, _ :=getString(args, "filterName")
	if filterName == "" {
		return err("Missing required argument 'filterName'")
}

	return ok("Applied filter: " + filterName)
}

func HandleOpenDocument(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filePath, _ :=getString(args, "filePath")
	if filePath == "" {
		return err("Missing required argument 'filePath'")
}

	return ok("Opened document: " + filePath)
}