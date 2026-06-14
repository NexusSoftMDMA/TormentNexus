package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleReadDocx(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "file_path")
	f, e := os.Open(path)
	if e != nil {
		return err(fmt.Sprintf("failed to open file: %v", e))
}

	defer f.Close()
	data, e := io.ReadAll(f)
	if e != nil {
		return err(fmt.Sprintf("failed to read file: %v", e))
}

	return ok(fmt.Sprintf("read %d bytes", len(data)))
}

func HandleWriteDocx(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "file_path")
	content, _ :=getString(args, "content")
	e := os.WriteFile(path, []byte(content), 0644)
	if e != nil {
		return err(fmt.Sprintf("failed to write file: %v", e))
}

	return success("file written")
}