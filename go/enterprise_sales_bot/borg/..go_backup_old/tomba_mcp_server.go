package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
)

func HandleEmailFinder(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	domain, _ :=getString(args, "domain")
	firstName, _ :=getString(args, "first_name")
	lastName, _ :=getString(args, "last_name")
	if domain == "" {
		return err("domain is required")
}

	url := fmt.Sprintf("https://api.tomba.io/v1/email-finder?domain=%s", domain)
	if firstName != "" {
		url += "&first_name=" + firstName
	}
	if lastName != "" {
		url += "&last_name=" + lastName
	}
	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("failed to create request: " + e.Error())
}

	req.Header.Set("X-Api-Key", os.Getenv("TOMBA_API_KEY"))
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("request failed: " + e.Error())
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err("failed to read response: " + e.Error())
}

	if resp.StatusCode != 200 {
		return err("API error: " + string(body))
}

	return success(string(body))
}