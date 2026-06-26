package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func ValidateApi(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	spec, _ :=getString(args, "spec")
	if spec == "" {
		return err("spec is required")
}

	payload, _ := json.Marshal(map[string]string{"spec": spec})
	resp, e := http.DefaultClient.Post("https://api.apimatic.io/validator", "application/json", bytes.NewReader(payload))
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	return ok(string(body))
}

func GetValidators(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resp, e := http.DefaultClient.Get("https://api.apimatic.io/validators")
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body failed: %v", e))
}

	return ok(string(body))
}