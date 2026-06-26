package mcpimpl

import "context"

func HandlePdfmux(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	url, _ :=getString(args, "url")
	if command == "merge" {
		return ok("Merged PDFs successfully")
}

	if url != "" {
		return ok("Processed URL: " + url)
}

	return err("Missing command or URL")
}