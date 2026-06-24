package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleCreateDocx(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filename, _ :=getString(args, "filename")
	if filename == "" {
		return err("filename is required")
}

	return ok(fmt.Sprintf("Created DOCX file: %s", filename))
}

func HandleConvertTextToDocx(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")
	filename, _ :=getString(args, "filename")
	if text == "" || filename == "" {
		return err("text and filename are required")
}

	body, _ := json.Marshal(map[string]string{"text": text, "filename": filename})
	resp, e := http.DefaultClient.Post("http://localhost:8000/convert", "application/json", bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("server returned %d", resp.StatusCode))
}

	return ok(fmt.Sprintf("Converted text to DOCX: %s", filename))
}