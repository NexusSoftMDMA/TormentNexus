package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func HandleSendDocumentForSignature(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	filePath, _ :=getString(args, "filePath")
	signerEmail, _ :=getString(args, "signerEmail")
	signerName, _ :=getString(args, "signerName")
	if filePath == "" || signerEmail == "" {
		return err("filePath and signerEmail are required")
}

	apiKey := os.Getenv("BOLDSIGN_API_KEY")
	if apiKey == "" {
		return err("BOLDSIGN_API_KEY environment variable not set")
}

	body := map[string]interface{}{
		"files": []string{filePath},
		"recipients": []map[string]interface{}{
			{"email": signerEmail, "name": signerName, "role": "Signer"},
		},
	}
	payload, e := json.Marshal(body)
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	req, e := http.NewRequestWithContext(ctx, "POST", "https://api.boldsign.com/v1/document/send", bytes.NewReader(payload))
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
	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d", resp.StatusCode))
}

	return success("Document sent for signature successfully")
}