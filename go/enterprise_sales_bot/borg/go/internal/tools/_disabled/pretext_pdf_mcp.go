package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleGeneratePDF(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	content, _ :=getString(args, "content")
	filename, _ :=getString(args, "filename")
	if !strings.HasSuffix(filename, ".pdf") {
		filename = filename + ".pdf"
	}
	payload := map[string]interface{}{
		"content":  content,
		"filename": filename,
		"format":   "A4",
	}
	body, e := json.Marshal(payload)
	if e != nil {
		return err("failed to marshal json")
	}
	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.pretext.dev/v1/pdf", strings.NewReader(string(body)))
	if e != nil {
		return err("failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
	}
	defer resp.Body.Close()
	data, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
	}
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("server error: %d %s", resp.StatusCode, string(data)))
	}
	return success("PDF generated successfully")
}

func HandleValidateJSON(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	content, _ :=getString(args, "content")
	var dummy interface{}
	if e := json.Unmarshal([]byte(content), &dummy); e != nil {
		return err("invalid json: " + e.Error())
	}
	return success("JSON is valid")
}// touch 1781132138
