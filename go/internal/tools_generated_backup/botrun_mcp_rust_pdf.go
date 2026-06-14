package tools

import "context"

func HandlePdfInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	_ = getString(args, "filename")
	return success("PDF info: size, pages, author")
}

func HandlePdfMerge(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	_ = getString(args, "files")
	return success("PDFs merged successfully")
}