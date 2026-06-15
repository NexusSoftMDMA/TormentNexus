package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
)

func HandleResolveIdentity(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	credential, _ :=getString(args, "credential")
	if credential == "" {
		return err("credential is required")
}

	base := os.Getenv("AGENT_IDENTITY_BASE_URL")
	if base == "" {
		base = "https://api.agent-identity.datacules.com"
	}
	url := base + "/resolve?credential=" + credential
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("HTTP request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	return success(string(body))
}

func HandleValidateToken(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	if token == "" {
		return err("token is required")
}

	base := os.Getenv("AGENT_IDENTITY_BASE_URL")
	if base == "" {
		base = "https://api.agent-identity.datacules.com"
	}
	url := base + "/validate?token=" + token
	resp, e := http.DefaultClient.Get(url)
	if e != nil {
		return err("HTTP request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("invalid JSON response: " + e.Error())
}

	valid, found := result["valid"].(bool)
	if !found {
		return err("response missing 'valid' field")
}

	if valid {
		return ok("token is valid")
}

	return err("token is invalid")
}