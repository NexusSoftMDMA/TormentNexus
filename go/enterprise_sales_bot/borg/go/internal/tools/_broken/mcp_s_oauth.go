package mcpimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleAuthorize(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	clientID, _ :=getString(args, "client_id")
	redirectURI, _ :=getString(args, "redirect_uri")
	state, _ :=getString(args, "state")
	req, e := http.NewRequestWithContext(ctx, "GET", "https://auth.example.com/authorize", nil)
	if e != nil {
		return err("failed to create authorize request")
}

	q := req.URL.Query()
	q.Set("response_type", "code")
	q.Set("client_id", clientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("state", state)
	req.URL.RawQuery = q.Encode()
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("authorize request failed")
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e = json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("invalid response")
}

	return ok(fmt.Sprintf("authorization url: %s", req.URL.String()))
}

func HandleToken(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	clientID, _ :=getString(args, "client_id")
	clientSecret, _ :=getString(args, "client_secret")
	redirectURI, _ :=getString(args, "redirect_uri")
	body := fmt.Sprintf("grant_type=authorization_code&code=%s&client_id=%s&client_secret=%s&redirect_uri=%s",
		code, clientID, clientSecret, redirectURI)
	resp, e := http.DefaultClient.Post("https://auth.example.com/token", "application/x-www-form-urlencoded",
		nil)
	if e != nil {
		return err("token request failed")
}

	defer resp.Body.Close()
	var result map[string]interface{}
	if e = json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("invalid token response")
}

	accessToken, found := result["access_token"].(string)
	if !found {
		return err("access_token missing in response")
}

	return ok(fmt.Sprintf("access token: %s", accessToken))
}