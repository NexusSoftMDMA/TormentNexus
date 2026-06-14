package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

func HandleOAuthAuthorize(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	clientID, _ :=getString(args, "client_id")
	redirectURI, _ :=getString(args, "redirect_uri")
	scope, _ :=getString(args, "scope")
	state, _ :=getString(args, "state")
	authEndpoint, _ :=getString(args, "authorization_endpoint")
	if authEndpoint == "" || clientID == "" || redirectURI == "" {
		return err("missing required parameters: authorization_endpoint, client_id, redirect_uri")
}

	v := url.Values{}
	v.Set("response_type", "code")
	v.Set("client_id", clientID)
	v.Set("redirect_uri", redirectURI)
	v.Set("scope", scope)
	v.Set("state", state)
	return success(authEndpoint + "?" + v.Encode())
}