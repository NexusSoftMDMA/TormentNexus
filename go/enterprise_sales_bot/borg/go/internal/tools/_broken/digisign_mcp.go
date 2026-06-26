package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func HandleSignDocument(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	docID, _ :=getString(args, "document_id")
	email, _ :=getString(args, "signer_email")
	if docID == "" || email == "" {
		return err("document_id and signer_email are required")
}

	payload := map[string]string{"document_id": docID, "signer_email": email}
	body, e := json.Marshal(payload)
	if e != nil {
		return err(fmt.Sprintf("marshal error: %v", e))
}

	resp, e := http.DefaultClient.Post("https://api.digisign.com/v1/sign", "application/json", bytes.NewReader(body))
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	respBody, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read response error: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error: %s", string(respBody)))
}

	return ok(string(respBody))
}