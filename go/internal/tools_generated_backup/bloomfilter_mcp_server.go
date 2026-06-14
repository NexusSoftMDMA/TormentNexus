package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func HandleSearchDomain(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	domain, _ :=getString(args, "domain")
	if domain == "" {
		return err("domain is required")
}

	reqURL := fmt.Sprintf("https://api.bloomfilter.net/domains/check?domain=%s", url.QueryEscape(domain))
	resp, e := http.DefaultClient.Get(reqURL)
	if e != nil {
		return err("failed to check domain: " + e.Error())
}

	defer resp.Body.Close()
	var result struct {
		Available bool   `json:"available"`
		Message   string `json:"message"`
	}
	if e := json.NewDecoder(resp.Body).Decode(&result); e != nil {
		return err("invalid response: " + e.Error())
}

	if result.Available {
		return ok(fmt.Sprintf("Domain '%s' is available", domain))
}

	return err(fmt.Sprintf("Domain '%s' is taken: %s", domain, result.Message))
}

func HandleRegisterDomain(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	domain, _ :=getString(args, "domain")
	email, _ :=getString(args, "owner_email")
	years, _ :=getInt(args, "years")
	if domain == "" || email == "" {
		return err("domain and owner_email are required")
}

	body, _ := json.Marshal(map[string]interface{}{
		"domain":      domain,
		"owner_email": email,
		"years":       years,
	})
	resp, e := http.DefaultClient.Post(
		"https://api.bloomfilter.net/domains/register",
		"application/json",
		strings.NewReader(string(body)),
	)
	if e != nil {
		return err("failed to register domain: " + e.Error())
}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		var errResp struct{ Error string }
		json.NewDecoder(resp.Body).Decode(&errResp)
		msg := "registration failed"
		if errResp.Error != "" {
			msg = errResp.Error
		}
		return err(msg)
}

	return ok(fmt.Sprintf("Domain '%s' registered successfully for %d years", domain, years))
}