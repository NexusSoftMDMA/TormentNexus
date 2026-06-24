package tools

import (
	"context"
	"net/http"
	"strings"
)

func HandleClassifyVulnerability(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	desc, _ :=getString(args, "description")
	if desc == "" {
		return err("missing description")
}

	desc = strings.ToLower(desc)
	severity := "low"
	if strings.Contains(desc, "critical") || strings.Contains(desc, "remote code") {
		severity = "critical"
	} else if strings.Contains(desc, "high") || strings.Contains(desc, "privilege") {
		severity = "high"
	} else if strings.Contains(desc, "medium") || strings.Contains(desc, "xss") {
		severity = "medium"
	}
	return ok("classified as " + severity)
}

func HandleCheckSecurityHeaders(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	url, _ :=getString(args, "url")
	if url == "" {
		return err("missing url")
}

	req, e := http.NewRequestWithContext(ctx, "GET", url, nil)
	if e != nil {
		return err("request error: " + e.Error())
}

	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		return err("fetch error: " + e.Error())
}

	defer resp.Body.Close()
	headers := resp.Header
	var found []string
	var missing []string
	checks := []string{"X-Content-Type-Options", "X-Frame-Options", "Content-Security-Policy", "Strict-Transport-Security"}
	for _, h := range checks {
		if headers.Get(h) != "" {
			found = append(found, h)
		} else {
			missing = append(missing, h)

	}
	return ok("found: " + strings.Join(found, ", ") + "; missing: " + strings.Join(missing, ", "))
}
}