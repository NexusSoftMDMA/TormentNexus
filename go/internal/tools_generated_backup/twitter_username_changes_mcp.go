package tools

import (
	"context"
	"net/http"
)

func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	username, _ :=getString(args, "username")
	if username == "" {
		return err("username is required")
}

	resp, e := http.DefaultClient.Get("https://api.twitter.com/2/users/by/username/" + username)
	if e != nil {
		return err("failed to fetch user: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err("Twitter API returned status " + resp.Status)
}

	return success("Fetched username info for " + username)
}