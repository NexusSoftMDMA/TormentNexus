package tools

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func HandleConnect(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serverURL, _ :=getString(args, "server_url")
	clientID, _ :=getString(args, "client_id")
	clientSecret, _ :=getString(args, "client_secret")
	tokenURL := serverURL + "/oauth/token"
	payload := url.Values{}
	payload.Set("grant_type", "client_credentials")
	payload.Set("client_id", clientID)
	payload.Set("client_secret", clientSecret)
	req, e := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(payload.Encode()))
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read body: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("token endpoint returned " + resp.Status)
}

	var result map[string]interface{}
	if e := json.Unmarshal(body, &result); e != nil {
		return err("failed to parse response: " + e.Error())
}

	accessToken, found := result["access_token"].(string)
	if !found {
		return err("no access_token in response")
}

	return ok("Connected with token: " + accessToken[:10] + "...")
}