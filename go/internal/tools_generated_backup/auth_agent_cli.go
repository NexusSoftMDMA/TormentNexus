package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func HandleLogin(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	username, _ :=getString(args, "username")
	password, _ :=getString(args, "password")
	authURL, _ :=getString(args, "auth_url")
	if username == "" || password == "" || authURL == "" {
		return err("missing required arguments: username, password, auth_url")
}

	body, e := json.Marshal(map[string]string{"username": username, "password": password})
	if e != nil {
		return err("failed to marshal request: " + e.Error())
}

	resp, e := http.DefaultClient.Post(authURL, "application/json", bytes.NewBuffer(body))
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	respBody, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err("auth failed with status " + resp.Status)
}

	var result map[string]interface{}
	if e := json.Unmarshal(respBody, &result); e != nil {
		return err("invalid response json: " + e.Error())
}

	token, found := result["token"].(string)
	if !found {
		return err("response missing token")
}

	return ok("Authenticated. Token: " + token)
}