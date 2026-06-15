package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
)

func HandleGenerateDocument(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	templateID, _ :=getString(args, "templateId")
	dataStr, _ :=getString(args, "data")
	format, _ :=getString(args, "outputFormat")
	if format == "" {
		format = "pdf"
	}
	apiBase := os.Getenv("CARBONE_API_URL")
	if apiBase == "" {
		apiBase = "https://api.carbone.io/v1"
	}
	body, _ := json.Marshal(map[string]interface{}{
		"templateId": templateID,
		"data":       json.RawMessage(dataStr),
		"outputFormat": format,
	})
	resp, e := http.DefaultClient.Post(apiBase+"/render", "application/json", bytes.NewReader(body))
	if e != nil {
		return err("API request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("API returned status " + resp.Status)
}

	return ok("Document generated successfully")
}

func HandleConvertDocument(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	fileURL, _ :=getString(args, "fileUrl")
	format, _ :=getString(args, "outputFormat")
	if format == "" {
		format = "pdf"
	}
	apiBase := os.Getenv("CARBONE_API_URL")
	if apiBase == "" {
		apiBase = "https://api.carbone.io/v1"
	}
	body, _ := json.Marshal(map[string]interface{}{
		"fileUrl":      fileURL,
		"outputFormat": format,
	})
	resp, e := http.DefaultClient.Post(apiBase+"/convert", "application/json", bytes.NewReader(body))
	if e != nil {
		return err("API request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("API returned status " + resp.Status)
}

	return ok("Document converted successfully")
}