package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func HandleGetSignatures(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key is required")
	}
	url := fmt.Sprintf("https://api.esignatures.com/v1/signatures?api_key=%s", apiKey)
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("failed to fetch signatures: " + e.Error())
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
	}
	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response")
	}
	return ok(fmt.Sprintf("Signatures: %v", result))
}

func HandleCreateSignature(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey, _ :=getString(args, "api_key")
	if apiKey == "" {
		return err("api_key is required")
	}
	documentID, _ :=getString(args, "document_id")
	if documentID == "" {
		return err("document_id is required")
	}
	signers, _ :=getString(args, "signers")
	url := fmt.Sprintf("https://api.esignatures.com/v1/signatures?api_key=%s", apiKey)
	payload := map[string]interface{}{"document_id": documentID, "signers": signers}
	body, _ := json.Marshal(payload)
	resp, e := http.DefaultClient.Post(url, "application/json", bytes.NewReader(body))
	if e != nil {
		return err("failed to create signature: " + e.Error())
	}
	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
	}
	return success("Signature created: " + string(respBody))
}