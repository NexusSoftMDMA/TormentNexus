package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
)

func HandleGeneratePDF(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	html, _ :=getString(args, "html")
	if html == "" {
		return err("html is required")
}

	apiKey := os.Getenv("PDFBROKER_API_KEY")
	if apiKey == "" {
		return err("PDFBROKER_API_KEY not set")
}

	body, _ := json.Marshal(map[string]string{"html": html})
	req, e := http.NewRequestWithContext(ctx, "POST", "https://pdfbroker.io/api/generate", bytes.NewReader(body))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return err("API returned status " + resp.Status)
}

	var result struct{ URL string `json:"url"` }
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("failed to decode response")
}

	return success("PDF generated: " + result.URL)
}