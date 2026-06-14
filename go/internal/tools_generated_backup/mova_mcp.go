package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleInvoiceOcr(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fileURL, _ :=getString(args, "file_url")
	if fileURL == "" {
		return err("file_url is required")
	}
	reqBody, e := json.Marshal(map[string]string{"url": fileURL})
	if e != nil {
		return err("failed to marshal request")
	}
	resp, e := http.DefaultClient.Post("https://api.mova.ai/v1/ocr", "application/json", bytes.NewReader(reqBody))
	if e != nil {
		return err("OCR service unavailable")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("OCR failed")
	}
	var result map[string]interface{}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("OCR parsing error")
	}
	return success(fmt.Sprintf("OCR result: %v", result))
}