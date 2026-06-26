package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

func HandleGetOrganizations(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	token := os.Getenv("QONTO_API_TOKEN")
	if token == "" {
		return err("QONTO_API_TOKEN not set")
}

	base := "https://third-party.qonto.com/v2/organizations"
	page, _ :=getString(args, "page")
	if page == "" {
		page = "1"
	}
	u, _ := url.Parse(base)
	q := u.Query()
	q.Set("page", page)
	u.RawQuery = q.Encode()
	req, e := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	req.Header.Set("Authorization", token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	return success(string(body))
}

func HandleGetBankAccounts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	slug, _ :=getString(args, "organization_slug")
	if slug == "" {
		return err("organization_slug required")
}

	token := os.Getenv("QONTO_API_TOKEN")
	if token == "" {
		return err("QONTO_API_TOKEN not set")
}

	pageStr, _ :=getString(args, "page")
	if pageStr == "" {
		pageStr = "1"
	}
	perPage, _ :=getString(args, "per_page")
	u, _ := url.Parse(fmt.Sprintf("https://third-party.qonto.com/v2/organizations/%s/bank_accounts", slug))
	q := u.Query()
	q.Set("page", pageStr)
	if perPage != "" {
		q.Set("per_page", perPage)

	u.RawQuery = q.Encode()
	req, e := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if e != nil {
		return err(fmt.Sprintf("create request: %v", e))
}

	req.Header.Set("Authorization", token)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err(fmt.Sprintf("request failed: %v", e))
}

	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return err(fmt.Sprintf("read body: %v", e))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)))
}

	return success(string(body))
}
}