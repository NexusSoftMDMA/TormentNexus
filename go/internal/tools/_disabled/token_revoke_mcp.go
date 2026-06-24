package tools

import (
	"context"
	"net/http"
	"net/url"
	"os"
)

func HandleRevokeToken(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token, _ :=getString(args, "token")
	if token == "" {
		return err("token is required")
}

	revokeURL := os.Getenv("REVOKE_URL")
	if revokeURL == "" {
		return err("REVOKE_URL environment variable not set")
}

	resp, e := http.DefaultClient.PostForm(revokeURL, url.Values{"token": {token}})
	if e != nil {
		return err("failed to revoke token: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("revoke endpoint returned status " + resp.Status)
}

	return success("Token revoked successfully")
}