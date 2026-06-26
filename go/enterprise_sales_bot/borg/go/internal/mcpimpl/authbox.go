package mcpimpl

import (
	"context"
	"net/http"
)

func HandleVerifyToken(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	if token == "" {
		return err("token is required")
}

	resp, e := http.DefaultClient.Get("https://authbox.example.com/verify?token=" + token)
	if e != nil {
		return err("verification request failed: " + e.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err("verification failed: status " + resp.Status)
}

	return success("token verified")
}