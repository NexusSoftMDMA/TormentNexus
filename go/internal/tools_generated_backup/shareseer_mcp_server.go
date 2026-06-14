package tools

import "context"

func HandleShareFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filename, _ :=getString(args, "filename")
	public, _ :=getBool(args, "public")
	if filename == "" {
		return err("filename is required")
}

	msg := "shared file: " + filename
	if public {
		msg += " (public)"
	}
	return ok(msg)
}

func HandleListShares(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	list := []string{"file1.txt", "file2.pdf", "image.png"}
	return success("shared files: " + strings.Join(list, ", "))
}