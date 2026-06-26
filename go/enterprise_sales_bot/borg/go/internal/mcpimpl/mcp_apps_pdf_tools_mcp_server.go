package mcpimpl

import "context"

func HandleExtractText_mcp_apps_pdf_tools_mcp_server(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
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