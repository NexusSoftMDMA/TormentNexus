package tools

import "context"

func HandleExtractText(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	file, _ :=getString(args, "file")
	if file == "" {
		return err("file is required")
}

	return ok("extracted text from " + file)
}

func HandleMergePDFs(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	files, _ :=getString(args, "files")
	if files == "" {
		return err("files is required")
}

	return ok("merged PDFs: " + files)
}