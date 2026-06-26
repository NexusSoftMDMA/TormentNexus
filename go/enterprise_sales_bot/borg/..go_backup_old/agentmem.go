package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func HandleRemember(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	value, _ :=getString(args, "value")
	apiURL, _ :=getString(args, "api_url")
	body := map[string]string{"key": key, "value": value}
	jsonBody, e := json.Marshal(body)
	if e != nil {
		return err("failed to encode request body")
}

	req, e := http.NewRequestWithContext(ctx, "POST", apiURL+"/remember", strings.NewReader(string(jsonBody)))
	if e != nil {
		return err("failed to create request")
}

	req.Header.Set("Content-Type", "application/json")
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	_, _ = io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return err("server error")
}

	return success("memory stored")
}

func HandleRecall(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	apiURL, _ :=getString(args, "api_url")
	u, e := url.Parse(apiURL + "/recall")
	if e != nil {
		return err("invalid url")
}

	q := u.Query()
	q.Set("key", key)
	u.RawQuery = q.Encode()
	req, e := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if e != nil {
		return err("failed to create request")
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	respBody, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response")
}

	if resp.StatusCode != http.StatusOK {
		return err("server error")
}

	var result struct {
		Value string `json:"value"`
	}
	if e := json.Unmarshal(respBody, &result); e != nil {
		return err("invalid response")
}

	return ok(result.Value)
}