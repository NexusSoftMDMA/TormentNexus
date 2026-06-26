package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func HandleExecuteLisp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	reqBody, e := json.Marshal(map[string]string{"code": code})
	if e != nil {
		return err("failed to marshal request")
}

	resp, e := http.DefaultClient.Post("http://localhost:8080/execute-lisp", "application/json", bytes.NewBuffer(reqBody))
	if e != nil {
		return err("failed to send request: " + e.Error())
}

	defer resp.Body.Close()
	return success("lisp executed")
}

func HandleGetDocumentInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("document info retrieved")
}