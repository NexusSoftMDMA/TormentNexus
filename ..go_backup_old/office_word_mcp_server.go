package tools

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

func HandleCreateDocument(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fileName, _ :=getString(args, "file_name")
	content, _ :=getString(args, "content")
	if fileName == "" {
		return err("file_name is required")
}

	url := fmt.Sprintf("http://localhost:8080/word/create?name=%s", fileName)
	resp, e := http.DefaultClient.Post(url, "text/plain", strings.NewReader(content))
	if e != nil {
		return err("failed to create document: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("document creation failed")
}

	return ok("document created: " + fileName)
}

func HandleReadDocument(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fileName, _ :=getString(args, "file_name")
	if fileName == "" {
		return err("file_name is required")
}

	url := fmt.Sprintf("http://localhost:8080/word/read?name=%s", fileName)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to read document: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("document not found")
}

	return success("document content retrieved")
}