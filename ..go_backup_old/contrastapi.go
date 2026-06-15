package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleCVE(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getString(args, "id")
	if id == "" {
		return err("missing cve id")
}

	resp, e := http.DefaultClient.Get(fmt.Sprintf("https://cveapi.example.com/%s", id))
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var data map[string]interface{}
	if e := json.Unmarshal(body, &data); e != nil {
		return err("parse error")
}

	return ok(fmt.Sprintf("CVE %s: %v", id, data["description"]))
}

func HandleEmailSecurity(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	domain, _ :=getString(args, "domain")
	if domain == "" {
		return err("missing domain")
}

	resp, e := http.DefaultClient.Get(fmt.Sprintf("https://emailsec.example.com/%s", domain))
	if e != nil {
		return err("request failed")
}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if strings.Contains(string(body), "SPF") {
		return ok("SPF record found")
}

	return ok("no SPF record")
}